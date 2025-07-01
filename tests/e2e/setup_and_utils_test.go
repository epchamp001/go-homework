//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"pvz-cli/internal/app"
	"pvz-cli/internal/config"
	"pvz-cli/internal/config/storage"
	"pvz-cli/tests/integration/testutil"
)

func freePort(t *testing.T) int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

type testEnv struct {
	BaseURL string
	cancel  context.CancelFunc
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	ctx := context.Background()
	pgC, err := testutil.NewPostgreSQLContainer(ctx)
	require.NoError(t, err)
	require.NoError(t, testutil.RunMigrations(pgC.GetDSN(), "../../migrations"))

	host, _ := pgC.Host(ctx)
	natPort, _ := pgC.MappedPort(ctx, "5432/tcp")
	port, _ := strconv.Atoi(natPort.Port())

	cfg, err := config.LoadConfig("../../configs/config.yaml", "../../.env")
	require.NoError(t, err)

	pgCfg := &cfg.Storage.Postgres
	pgCfg.Username = pgC.Config.User
	pgCfg.Password = pgC.Config.Password
	pgCfg.Database = pgC.Config.Database
	pgCfg.SSLMode = "disable"
	ep := storage.PostgresEndpoint{Host: host, Port: port}
	pgCfg.Master = ep
	pgCfg.Replicas = []storage.PostgresEndpoint{ep, ep}

	cfg.GRPCServer.Port = freePort(t)
	cfg.Gateway.Port = freePort(t)

	log, err := app.SetupLogger(cfg.Logging)
	require.NoError(t, err)

	srv := app.NewServer(cfg, log)

	runCtx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = srv.Run(runCtx)
	}()

	// ждем gateway
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", cfg.Gateway.Port)
	require.Eventually(t, func() bool {
		req, _ := http.NewRequest(http.MethodGet, baseURL+"/v1/orders?user_id=1", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode != http.StatusNotFound // роут зарегистрирован, БД пуста
	}, 10*time.Second, 200*time.Millisecond, "gateway не поднялся")

	return &testEnv{BaseURL: baseURL, cancel: cancel}
}

// postJSON делает POST с JSON-payload и возвращает тело как map[string]any
func postJSON(t *testing.T, url, body string) map[string]any {
	t.Helper()
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

func getJSON(t *testing.T, url string) map[string]any {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}
