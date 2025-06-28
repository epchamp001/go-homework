package e2e

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestE2E_AcceptOrder_ValidationError(t *testing.T) {
	env := newTestEnv(t)
	defer env.cancel()

	// order_id = 0
	badReq := `{"order_id":0,"user_id":1,"expires_at":"2030-01-01T00:00:00Z","weight":1,"price":1}`

	resp, err := http.Post(env.BaseURL+"/v1/orders/accept",
		"application/json", bytes.NewBufferString(badReq))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
