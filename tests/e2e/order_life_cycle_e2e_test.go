package e2e

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestE2E_OrderLifecycle(t *testing.T) {
	env := newTestEnv(t)
	defer env.cancel()

	const orderID = 1001
	const userID = 555

	// AcceptOrder
	accReq := `{
	  "order_id": ` + strconv.Itoa(orderID) + `,
	  "user_id":  ` + strconv.Itoa(userID) + `,
	  "expires_at":"2030-12-31T23:59:59Z",
	  "package":  "PACKAGE_TYPE_BAG",
	  "weight":   2.5,
	  "price":    199.0
	}`
	accResp := postJSON(t, env.BaseURL+"/v1/orders/accept", accReq)
	assert.Equal(t, "ORDER_STATUS_EXPECTS", accResp["status"])
	assert.Equal(t, strconv.Itoa(orderID), accResp["order_id"])

	// ListOrders
	listResp := getJSON(t, env.BaseURL+"/v1/orders?user_id="+strconv.Itoa(userID))
	orders := listResp["orders"].([]any)
	require.Len(t, orders, 1)
	assert.Equal(t, "ORDER_STATUS_EXPECTS",
		orders[0].(map[string]any)["status"])

	// ProcessOrders
	procReq := `{
	  "user_id":  ` + strconv.Itoa(userID) + `,
	  "action":   "ACTION_TYPE_ISSUE",
	  "order_ids":[` + strconv.Itoa(orderID) + `]
	}`
	procResp := postJSON(t, env.BaseURL+"/v1/orders/process", procReq)
	assert.Contains(t, procResp["processed"], strconv.Itoa(orderID))

	// ListOrders
	listAfterIssue := getJSON(t, env.BaseURL+"/v1/orders?user_id="+strconv.Itoa(userID))
	orders = listAfterIssue["orders"].([]any)
	require.Len(t, orders, 1)
	assert.Equal(t, "ORDER_STATUS_ACCEPTED",
		orders[0].(map[string]any)["status"])
}
