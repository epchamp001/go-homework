// Package app инициализирует и запускает основное приложение,
// включая конфигурацию, зависимости, маршруты и graceful shutdown.
//
// Этот пакет связывает все модули проекта и является точкой входа при запуске бинарника.
package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"pvz-cli/internal/config"
	"pvz-cli/internal/handler"
	"pvz-cli/internal/handler/middleware"
	"pvz-cli/internal/repository/storage/filerepo"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/closer"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/logger"
	pvzpb "pvz-cli/pkg/pvz"
)

// Server позволяет удобно и аккуратно поднимать весь проект и его зависимости.
type Server struct {
	closer     *closer.Closer
	grpcServer *grpc.Server
	log        logger.Logger
	cfg        *config.Config
	hndl       handler.ReportsHandler
}

// NewServer создаёт новое приложение с инициализацией хранилища, бизнес-логики и REPL.
func NewServer(cfg *config.Config, log logger.Logger) *Server {
	c := closer.NewCloser()

	limiter := rate.NewLimiter(rate.Limit(5), 5)
	
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.RateLimitInterceptor(limiter)),
	)
	reflection.Register(grpcSrv)

	s := &Server{
		closer:     c,
		grpcServer: grpcSrv,
		log:        log,
		cfg:        cfg,
	}

	s.setupGRPC()

	return s
}

func (s *Server) setupGRPC() {
	repo, err := filerepo.NewFileRepo("data")
	if err != nil {
		s.log.Fatalw("could not create repo",
			"error", err,
		)
	}

	svc := usecase.NewService(repo)

	hndl := handler.NewReportsHandler(svc)

	s.hndl = hndl

	handler.RegisterOrderService(s.grpcServer, svc)
}

// Run запускает REPL-приложение, обрабатывающее пользовательские команды.
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

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := fmt.Sprintf("%s:%d", s.cfg.GRPCServer.Endpoint, s.cfg.GRPCServer.Port)
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return errs.Wrap(err, errs.CodeInternalError, "failed to dial gRPC endpoint")

	}

	s.closer.Add(func(ctx context.Context) error {
		s.log.Infow("Closing gRPC connection")
		return conn.Close()
	})

	if err := pvzpb.RegisterOrdersServiceHandler(ctx, mux, conn); err != nil {
		return errs.Wrap(err, errs.CodeInternalError, "failed to register PVZ service handler")
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

	return r
}

func setupHTTPRoutes(r *gin.Engine, svc usecase.Service) {
	r.GET("/v1/reports/clients/raw", func(c *gin.Context) {

		sortBy := c.Query("sortBy")

		dataBytes, err := svc.GenerateClientReportByte(sortBy)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to generate report: %v", err)
			return
		}
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=clients_report.xlsx")
		c.Header("Content-Length", fmt.Sprintf("%d", len(dataBytes)))
		c.Writer.Write(dataBytes)
	})
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.closer.Close(ctx)
}
