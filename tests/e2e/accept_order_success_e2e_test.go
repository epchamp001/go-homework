//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_AcceptOrder_Success(t *testing.T) {
	env := newTestEnv(t)
	defer env.cancel()

	reqJSON := `{
	  "order_id": "42",
	  "user_id":  "777",
	  "expires_at": "2030-12-31T12:00:00Z",
	  "package":   "PACKAGE_TYPE_BOX",
	  "weight":    1.2,
	  "price":     100.0
	}`

	resp, err := http.Post(env.BaseURL+"/v1/orders/accept",
		"application/json", bytes.NewBufferString(reqJSON))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got struct {
		Status  string `json:"status"`
		OrderID string `json:"order_id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))

	assert.Equal(t, "42", got.OrderID)
	assert.Equal(t, "ORDER_STATUS_EXPECTS", got.Status)
}
