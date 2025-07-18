// Package app инициализирует и запускает основное приложение,
// включая конфигурацию, зависимости, маршруты и graceful shutdown.
//
// Этот пакет связывает все модули проекта и является точкой входа при запуске бинарника.
package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"pvz-cli/internal/config"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/handler"
	"pvz-cli/internal/handler/middleware"
	"pvz-cli/internal/infrastructure/kafka/producer"
	"pvz-cli/internal/repository/storage/postgres"
	"pvz-cli/internal/usecase/outboxworker"
	"pvz-cli/internal/usecase/packaging"
	"pvz-cli/internal/usecase/service"
	"pvz-cli/pkg/auth"
	"pvz-cli/pkg/cache"
	"pvz-cli/pkg/cache/lru"
	"pvz-cli/pkg/closer"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/logger"
	pvzpb "pvz-cli/pkg/pvz"
	"pvz-cli/pkg/txmanager"
	"pvz-cli/pkg/wpool"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

// Server позволяет удобно и аккуратно поднимать весь проект и его зависимости.
type Server struct {
	closer     *closer.Closer
	grpcServer *grpc.Server
	log        logger.Logger
	cfg        *config.Config
	hndl       handler.ReportsHandler
	txMgr      txmanager.TxManager
	wp         *wpool.Pool
}

// NewServer создаёт новое приложение с инициализацией хранилища, бизнес-логики и REPL.
func NewServer(cfg *config.Config, log logger.Logger) *Server {
	c := closer.NewCloser()

	limiter := rate.NewLimiter(rate.Limit(5), 5)

	masterPool, err := cfg.Storage.ConnectMaster(log)
	if err != nil {
		log.Fatalw("connect to master postgres",
			"error", err)
	}
	c.Add(func(ctx context.Context) error {
		log.Infow("Closing Master PostgreSQL pool")
		masterPool.Close()
		return nil
	})

	replPool1, err := cfg.Storage.ConnectReplica(1, log)
	if err != nil {
		log.Fatalw("connect to replica1 postgres",
			"error", err)
	}
	c.Add(func(ctx context.Context) error {
		log.Infow("Closing Replica 1 PostgreSQL pool")
		replPool1.Close()
		return nil
	})

	replPool2, err := cfg.Storage.ConnectReplica(1, log)
	if err != nil {
		log.Fatalw("connect to replica2 postgres",
			"error", err)
	}
	c.Add(func(ctx context.Context) error {
		log.Infow("Closing Replica 2 PostgreSQL pool")
		replPool2.Close()
		return nil
	})

	tx := txmanager.NewTransactor(masterPool, []*pgxpool.Pool{replPool1, replPool2}, log)

	wp := wpool.NewWorkerPool(cfg.Workers.Start, cfg.Workers.Queue, log)
	c.Add(func(ctx context.Context) error {
		log.Infow("Stopping worker-pool")
		wp.Stop()
		return nil
	})

	creds := auth.StaticCreds{User: cfg.Admin.User, Pass: cfg.Admin.Pass}
	basic := auth.NewUnaryBasicAuthWithFilter(
		creds,
		func(full string) bool { // применяю только к admin-методам
			return strings.HasPrefix(full, "/admin.AdminService/")
		},
	)

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.RateLimitInterceptor(limiter),
			basic,
		),
	)
	reflection.Register(grpcSrv)

	s := &Server{
		closer:     c,
		grpcServer: grpcSrv,
		log:        log,
		cfg:        cfg,
		txMgr:      tx,
		wp:         wp,
	}

	s.setupGRPC()

	prodCfg := producer.Config{
		Brokers:      s.cfg.Kafka.Brokers,
		Topic:        s.cfg.Kafka.Topic,
		Idempotent:   true,
		RequiredAcks: "all",
	}
	prod, err := producer.NewProducer(prodCfg)
	if err != nil {
		s.log.Fatalw("failed to create kafka producer", "error", err)
	}
	s.closer.Add(func(ctx context.Context) error {
		s.log.Infow("Closing Kafka producer")
		prod.Close()
		return nil
	})

	workerCtx, cancel := context.WithCancel(context.Background())
	s.closer.Add(func(_ context.Context) error {
		cancel()
		return nil
	})

	outboxRepo := postgres.NewOutboxPostgresRepo(s.txMgr)
	w := outboxworker.NewWorker(
		s.txMgr,
		outboxRepo,
		prod,
		s.cfg.Outbox.BatchSize,
		s.cfg.Outbox.Interval,
		s.log,
	)
	go func() {
		if err := w.Run(workerCtx); err != nil && !errors.Is(err, context.Canceled) {
			s.log.Errorw("outbox worker stopped with error", "error", err)
		}
	}()

	return s
}

func (s *Server) setupGRPC() {

	orderRepo := postgres.NewOrdersPostgresRepo(s.txMgr)
	hrRepo := postgres.NewHistoryAndReturnsPostgresRepo(s.txMgr)
	outboxRepo := postgres.NewOutboxPostgresRepo(s.txMgr)

	strategyProvider := packaging.NewDefaultProvider()

	cacheCfg := cache.Config[string]{
		Capacity: s.cfg.OrderCache.Capacity,
		TTL:      s.cfg.OrderCache.TTL,
		Strategy: lru.NewLRUStrategy[string](s.cfg.OrderCache.Capacity),
	}
	orderCache := cache.New[string, *models.Order](cacheCfg)
	s.closer.Add(func(_ context.Context) error {
		orderCache.Close()
		return nil
	})

	svc := service.NewService(s.txMgr, orderRepo, hrRepo, outboxRepo, strategyProvider, s.wp, orderCache)

	hndl := handler.NewReportsHandler(svc)

	s.hndl = hndl

	handler.RegisterOrderService(s.grpcServer, svc, s.log, s.wp)
	handler.RegisterAdminService(s.grpcServer, s.wp)
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.runGRPC(ctx); err != nil {
		return err
	}

	if err := s.runGateway(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Server) runGRPC(ctx context.Context) error {
	grpcPort := s.cfg.GRPCServer.Port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		return errs.Wrap(err, errs.CodeInternalError, fmt.Sprintf("failed to listen on port %d", grpcPort))
	}

	s.closer.Add(func(ctx context.Context) error {
		s.log.Infow("Shutting down gRPC server")
		s.grpcServer.GracefulStop()
		return nil
	})

	go func() {
		s.log.Infow("Starting gRPC server",
			"address", lis.Addr().String(),
		)
		if err := s.grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			s.log.Fatalw("gRPC server error",
				"error", err,
			)
		}
	}()

	return nil
}

func (s *Server) runGateway(ctx context.Context) error {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	endpoint := fmt.Sprintf("%s:%d", s.cfg.GRPCServer.Endpoint, s.cfg.GRPCServer.Port)

	// вместо grpc.DialContext использую grpc.NewClient, так как просит линтер и рекомендуют разрабы)
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return errs.Wrap(err, errs.CodeInternalError, "failed to create gRPC client")
	}

	s.closer.Add(func(ctx context.Context) error {
		s.log.Infow("Closing gRPC connection")
		return conn.Close()
	})

	if err := pvzpb.RegisterOrdersServiceHandler(ctx, mux, conn); err != nil {
		return errs.Wrap(err, errs.CodeInternalError, "failed to register PVZ service handler")
	}

	if err := pvzpb.RegisterAdminServiceHandler(ctx, mux, conn); err != nil {
		return errs.Wrap(err, errs.CodeInternalError, "register Admin handler")
	}

	gatewayRouter := s.setupRoutes(mux)

	gwAddr := fmt.Sprintf(":%d", s.cfg.Gateway.Port)
	gwServer := &http.Server{
		Addr:    gwAddr,
		Handler: gatewayRouter,
	}

	s.closer.Add(func(ctx context.Context) error {
		s.log.Infow("Shutting down gRPC gateway")
		return gwServer.Shutdown(ctx)
	})

	go func() {
		s.log.Infow("Starting gRPC gateway",
			"address", gwAddr,
		)
		if err := gwServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatalw("gRPC gateway error",
				"error", err,
			)
		}
	}()

	return nil
}

func (s *Server) setupRoutes(gw http.Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.StaticFile(
		"/swagger/spec/http/swagger.json",
		"./api/swagger/apidocs.swagger.json",
	)

	r.GET(
		"/swagger/http/*any",
		ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("/swagger/spec/http/swagger.json"),
		),
	)

	corsHandler := middleware.EnableCORS(gw)

	// Принять заказ от курьера
	r.POST("/v1/orders/accept", gin.WrapH(corsHandler))

	// Вернуть заказ курьеру (динамический order_id)
	r.POST("/v1/orders/:order_id/return", gin.WrapH(corsHandler))

	// Выдать заказы или принять возврат клиента
	r.POST("/v1/orders/process", gin.WrapH(corsHandler))

	// Получить список заказов клиента
	r.GET("/v1/orders", gin.WrapH(corsHandler))

	// Получить список возвратов клиентов
	r.GET("/v1/orders/returns", gin.WrapH(corsHandler))

	// Получить историю изменения заказов
	r.GET("/v1/orders/history", gin.WrapH(corsHandler))

	// Импортировать заказы из JSON-файла
	r.POST("/v1/orders/import", gin.WrapH(corsHandler))

	r.GET("/v1/reports/clients", s.hndl.DownloadClientReport)

	r.POST("/v1/admin/resizePool", gin.WrapH(corsHandler))

	return r
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.closer.Close(ctx)
}
