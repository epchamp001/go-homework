//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"pvz-cli/internal/app"
	"pvz-cli/internal/config"
	"pvz-cli/internal/config/storage"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/infrastructure/kafka/producer"
	"pvz-cli/internal/repository/storage/postgres"
	"pvz-cli/internal/usecase/packaging"
	"pvz-cli/internal/usecase/service"
	"pvz-cli/pkg/cache"
	"pvz-cli/pkg/cache/lru"
	"pvz-cli/pkg/logger"
	"pvz-cli/pkg/txmanager"
	"pvz-cli/pkg/wpool"
	"pvz-cli/tests/integration/testutil"
	"strconv"
	"strings"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/twmb/franz-go/pkg/kadm"
)

var dbCleanupMu sync.Mutex

type TestSuite struct {
	suite.Suite
	psqlContainer *testutil.PostgreSQLContainer
	masterPool    *pgxpool.Pool
	svc           service.Service
	wp            *wpool.Pool

	outbox *postgres.OutboxPostgresRepo
	tx     txmanager.TxManager
	log    logger.Logger

	kafkaCont testcontainers.Container
	kafkaAddr string
	kadm      *kadm.Client // общий admin
	baseProd  producer.Config

	ordCache *cache.Cache[string, *models.Order]

	fixtureNow time.Time
}

func (s *TestSuite) SetupSuite() {
	cfg, err := config.LoadConfig("../../configs/config.yaml", "../../.env")
	s.Require().NoError(err)

	ctx := context.Background()

	pgC, err := testutil.NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)
	s.psqlContainer = pgC

	err = testutil.RunMigrations(pgC.GetDSN(), "../../migrations")
	s.Require().NoError(err)

	host, err := pgC.Host(ctx)
	s.Require().NoError(err)

	natPort, err := pgC.MappedPort(ctx, "5432/tcp")
	s.Require().NoError(err)

	port, err := strconv.Atoi(natPort.Port())
	s.Require().NoError(err)

	pgCfg := &cfg.Storage.Postgres
	ep := storage.PostgresEndpoint{Host: host, Port: port}

	pgCfg.Username = pgC.Config.User
	pgCfg.Password = pgC.Config.Password
	pgCfg.Database = pgC.Config.Database

	pgCfg.Master = ep
	pgCfg.Replicas = []storage.PostgresEndpoint{ep, ep} // две псевдореплики

	pgCfg.SSLMode = "disable" // на случай, если добавить tls, то в тестах убираю его

	log, err := app.SetupLogger(cfg.Logging)
	s.Require().NoError(err)

	master, err := cfg.Storage.ConnectMaster(log)
	s.Require().NoError(err)

	s.masterPool = master

	replica1, err := cfg.Storage.ConnectReplica(0, log)
	s.Require().NoError(err)

	replica2, err := cfg.Storage.ConnectReplica(1, log)
	s.Require().NoError(err)

	txmngr := txmanager.NewTransactor(master, []*pgxpool.Pool{replica1, replica2}, log)

	orderRepo := postgres.NewOrdersPostgresRepo(txmngr)
	hrRepo := postgres.NewHistoryAndReturnsPostgresRepo(txmngr)
	outboxRepo := postgres.NewOutboxPostgresRepo(txmngr)

	strategyProvider := packaging.NewDefaultProvider()

	wp := wpool.NewWorkerPool(4, 16, log)
	s.wp = wp

	cacheCfg := cache.Config[string]{
		Capacity: cfg.OrderCache.Capacity,
		TTL:      cfg.OrderCache.TTL,
		Strategy: lru.NewLRUStrategy[string](cfg.OrderCache.Capacity),
	}
	orderCache := cache.New[string, *models.Order](cacheCfg)
	s.ordCache = orderCache
	svc := service.NewService(txmngr, orderRepo, hrRepo, outboxRepo, strategyProvider, wp, orderCache)
	s.svc = svc

	s.tx = txmngr
	s.outbox = outboxRepo
	s.log = log

	kc, err := kafka.Run(ctx,
		"confluentinc/confluent-local:7.4.0",
		kafka.WithClusterID("test-cluster"))
	s.Require().NoError(err)
	s.kafkaCont = kc // сохраняем, чтобы убить в TearDown

	addrs, _ := kc.Brokers(ctx)
	s.kafkaAddr = addrs[0]

	// admin-клиент
	adminCli, _ := kgo.NewClient(kgo.SeedBrokers(s.kafkaAddr))
	s.kadm = kadm.NewClient(adminCli)
	// НЕ закрываем adminCli здесь – он живёт всю Suite

	// smoke-test порта
	conn, err := net.DialTimeout("tcp", s.kafkaAddr, 5*time.Second)
	s.Require().NoError(err)
	conn.Close()

	// базовый конфиг продюсера – пользуем в тестах и только меняем Topic
	s.baseProd = producer.Config{
		Brokers:      []string{s.kafkaAddr},
		Idempotent:   true,
		RequiredAcks: "all",
	}

	s.fixtureNow = time.Date(2025, 6, 28, 10, 0, 0, 0, time.UTC)
}

func (s *TestSuite) TearDownTest() {
	dbCleanupMu.Lock()
	defer dbCleanupMu.Unlock()

	s.ordCache.Flush()

	ctx := context.Background()
	if err := s.truncateAll(ctx, s.masterPool); err != nil {
		s.T().Fatalf("truncate: %v", err)
	}
}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.ordCache.Close()
	s.masterPool.Close()
	s.Require().NoError(s.psqlContainer.Terminate(ctx))

	s.wp.Stop()
}

func TestSuite_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) createUniqueTopic(ctx context.Context) string {
	topic := "pvz.events-log-" + uuid.NewString()[:8] // 8 символов достаточно
	_, err := s.kadm.CreateTopic(ctx, 1, 1, nil, topic)
	s.Require().NoError(err)
	return topic
}

func (s *TestSuite) loadFixtures() {
	dbCleanupMu.Lock()
	defer dbCleanupMu.Unlock()

	// создаём временную директорию
	tmpDir, err := os.MkdirTemp("", "fixtures-")
	s.Require().NoError(err)
	defer os.RemoveAll(tmpDir)

	// FuncMap возвращает готовые строки в формате RFC3339
	fm := template.FuncMap{
		"now": func() string {
			return s.fixtureNow.Format(time.RFC3339)
		},
		"add": func(d string) string {
			dur, _ := time.ParseDuration(d)
			return s.fixtureNow.Add(dur).Format(time.RFC3339)
		},
		"sub": func(d string) string {
			dur, _ := time.ParseDuration(d)
			return s.fixtureNow.Add(-dur).Format(time.RFC3339)
		},
	}

	// обходим только файлы-шаблоны *.yaml.tmpl
	err = filepath.Walk("fixtures/storage", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !(strings.HasSuffix(info.Name(), ".yml.tmpl") ||
			strings.HasSuffix(info.Name(), ".yaml.tmpl")) {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		tmpl := template.Must(template.New(info.Name()).Funcs(fm).Parse(string(src)))

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			return err
		}

		rel, _ := filepath.Rel("fixtures/storage", path)
		outPath := filepath.Join(tmpDir, strings.TrimSuffix(rel, ".tmpl"))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(outPath, buf.Bytes(), info.Mode())
	})
	s.Require().NoError(err)

	// подключаемся к БД и загружаем отрендеренные фикстуры
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer db.Close()

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory(tmpDir),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func (s *TestSuite) truncateAll(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `
        SELECT string_agg(format('%I.%I', schemaname, tablename), ',')
        FROM pg_catalog.pg_tables
        WHERE schemaname = 'public';`
	var tbls pgtype.Text
	if err := pool.QueryRow(ctx, q).Scan(&tbls); err != nil {
		return err
	}
	if !tbls.Valid || tbls.String == "" {
		return nil
	}
	_, err := pool.Exec(
		ctx,
		fmt.Sprintf("TRUNCATE %s RESTART IDENTITY CASCADE", tbls.String),
	)
	return err
}
