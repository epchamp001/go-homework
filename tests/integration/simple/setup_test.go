package simple

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"os"
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
)

var dbCleanupMu sync.Mutex

var (
	masterPool *pgxpool.Pool
	svc        service.Service
	ctx        = context.Background()
)

// TestMain поднимает контейнер, миграции и инициализирует svc
func TestMain(m *testing.M) {
	cfg, err := config.LoadConfig("../../../configs/config.yaml", "../../../.env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load: %v\n", err)
		os.Exit(1)
	}

	pgC, err := testutil.NewPostgreSQLContainer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "container start: %v\n", err)
		os.Exit(1)
	}
	if err := testutil.RunMigrations(pgC.GetDSN(), "../../../migrations"); err != nil {
		fmt.Fprintf(os.Stderr, "migrations: %v\n", err)
		os.Exit(1)
	}

	host, _ := pgC.Host(ctx)

	natPort, _ := pgC.MappedPort(ctx, "5432/tcp")

	port, _ := strconv.Atoi(natPort.Port())

	pgCfg := &cfg.Storage.Postgres
	ep := storage.PostgresEndpoint{Host: host, Port: port}

	pgCfg.Username = pgC.Config.User
	pgCfg.Password = pgC.Config.Password
	pgCfg.Database = pgC.Config.Database

	pgCfg.Master = ep
	pgCfg.Replicas = []storage.PostgresEndpoint{ep, ep} // две псевдореплики

	pgCfg.SSLMode = "disable"

	log, err := app.SetupLogger(cfg.Logging)

	masterPool, _ = cfg.Storage.ConnectMaster(log)
	replica1, _ := cfg.Storage.ConnectReplica(0, log)
	replica2, _ := cfg.Storage.ConnectReplica(1, log)

	txmngr := txmanager.NewTransactor(masterPool, []*pgxpool.Pool{replica1, replica2}, log)

	orderRepo := postgres.NewOrdersPostgresRepo(txmngr)
	hrRepo := postgres.NewHistoryAndReturnsPostgresRepo(txmngr)
	stratProv := packaging.NewDefaultProvider()
	svc = service.NewService(txmngr, orderRepo, hrRepo, stratProv)

	code := m.Run()

	masterPool.Close()
	pgC.Terminate(ctx)
	os.Exit(code)
}

func cleanDB(t *testing.T) {
	t.Helper()
	dbCleanupMu.Lock()
	defer dbCleanupMu.Unlock()

	ctx := context.Background()
	// Можно прямо вызвать ваш truncateAll:
	require.NoError(t, truncateAll(ctx, masterPool))
}

func truncateAll(ctx context.Context, pool *pgxpool.Pool) error {
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
