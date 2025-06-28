package integration

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"pvz-cli/internal/app"
	"pvz-cli/internal/config"
	"pvz-cli/internal/config/storage"
	"pvz-cli/internal/repository/storage/postgres"
	"pvz-cli/internal/usecase/packaging"
	"pvz-cli/internal/usecase/service"
	"pvz-cli/pkg/txmanager"
	"pvz-cli/tests/integration/testutil"
	"strconv"
	"sync"
	"testing"
	"time"
)

var dbCleanupMu sync.Mutex

type TestSuite struct {
	suite.Suite
	psqlContainer *testutil.PostgreSQLContainer
	masterPool    *pgxpool.Pool
	svc           service.Service
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

	strategyProvider := packaging.NewDefaultProvider()

	svc := service.NewService(txmngr, orderRepo, hrRepo, strategyProvider)
	s.svc = svc
}

func (s *TestSuite) TearDownTest() {
	dbCleanupMu.Lock()
	defer dbCleanupMu.Unlock()

	ctx := context.Background()
	if err := s.truncateAll(ctx, s.masterPool); err != nil {
		s.T().Fatalf("truncate: %v", err)
	}
}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.masterPool.Close()
	s.Require().NoError(s.psqlContainer.Terminate(ctx))
}

func TestSuite_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) loadFixtures() {
	dbCleanupMu.Lock()
	defer dbCleanupMu.Unlock()

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer db.Close()

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures/storage"),
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
