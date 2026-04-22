package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderCreationFlow(t *testing.T) {
	suite, cleanup := SetupSuite(t)
	defer cleanup()

	db := suite.MySQL
	rdb := suite.Red
	ctx := context.Background()

	userID := SetupUser(t, db, "testuser-order", "buyer")
	skuID := "SKU-TEST-001"
	SetupInventory(t, db, skuID, 100, 0)

	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())
	bizNo := fmt.Sprintf("BIZ-%d", time.Now().UnixNano())

	_, err := db.ExecContext(ctx,
		`INSERT INTO orders(order_id,biz_no,user_id,status,total_amount,idempotent_key,version,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,NOW(),NOW())`,
		orderID, bizNo, userID, "PENDING", 10000, "idem-"+orderID, 0)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO order_items(order_id,sku_id,quantity,unit_price,subtotal,created_at)
		VALUES(?,?,?,?,?,NOW())`,
		orderID, skuID, 2, 5000, 10000)
	require.NoError(t, err)

	VerifyOrderStatus(t, db, orderID, "PENDING")

	t.Logf("Order created: %s, BizNo: %s", orderID, bizNo)
}

func TestOrderCancelFlow(t *testing.T) {
	suite, cleanup := SetupSuite(t)
	defer cleanup()

	db := suite.MySQL
	ctx := context.Background()

	userID := SetupUser(t, db, "testuser-cancel", "buyer")
	skuID := "SKU-TEST-002"
	SetupInventory(t, db, skuID, 50, 0)

	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())
	bizNo := fmt.Sprintf("BIZ-%d", time.Now().UnixNano())

	_, err := db.ExecContext(ctx,
		`INSERT INTO orders(order_id,biz_no,user_id,status,total_amount,idempotent_key,version,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,NOW(),NOW())`,
		orderID, bizNo, userID, "PENDING", 5000, "idem-"+orderID, 0)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO order_items(order_id,sku_id,quantity,unit_price,subtotal,created_at)
		VALUES(?,?,?,?,?,NOW())`,
		orderID, skuID, 1, 5000, 5000)
	require.NoError(t, err)

	VerifyOrderStatus(t, db, orderID, "PENDING")

	_, err = db.ExecContext(ctx,
		`UPDATE orders SET status = 'CANCELED', updated_at = NOW() WHERE order_id = ?`,
		orderID)
	require.NoError(t, err)

	VerifyOrderStatus(t, db, orderID, "CANCELED")
}

func TestPaymentSuccessFlow(t *testing.T) {
	suite, cleanup := SetupSuite(t)
	defer cleanup()

	db := suite.MySQL
	ctx := context.Background()

	userID := SetupUser(t, db, "testuser-payment", "buyer")
	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())
	bizNo := fmt.Sprintf("BIZ-%d", time.Now().UnixNano())

	_, err := db.ExecContext(ctx,
		`INSERT INTO orders(order_id,biz_no,user_id,status,total_amount,idempotent_key,version,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,NOW(),NOW())`,
		orderID, bizNo, userID, "PENDING", 10000, "idem-"+orderID, 0)
	require.NoError(t, err)

	paymentID := fmt.Sprintf("pay-%d", time.Now().UnixNano())
	_, err = db.ExecContext(ctx,
		`INSERT INTO payments(payment_id,order_id,amount,status,request_id,created_at,updated_at)
		VALUES(?,?,?,?,?,NOW(),NOW())`,
		paymentID, orderID, 10000, "PENDING", "req-"+paymentID)
	require.NoError(t, err)

	VerifyPaymentStatus(t, db, paymentID, "PENDING")

	_, err = db.ExecContext(ctx,
		`UPDATE payments SET status = 'SUCCESS', updated_at = NOW() WHERE payment_id = ?`,
		paymentID)
	require.NoError(t, err)

	VerifyPaymentStatus(t, db, paymentID, "SUCCESS")

	VerifyOrderStatus(t, db, orderID, "PENDING")
}

func TestPaymentFailureCompensation(t *testing.T) {
	suite, cleanup := SetupSuite(t)
	defer cleanup()

	db := suite.MySQL
	ctx := context.Background()

	userID := SetupUser(t, db, "testuser-comp", "buyer")
	skuID := "SKU-TEST-003"
	SetupInventory(t, db, skuID, 20, 0)

	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())
	bizNo := fmt.Sprintf("BIZ-%d", time.Now().UnixNano())

	_, err := db.ExecContext(ctx,
		`INSERT INTO orders(order_id,biz_no,user_id,status,total_amount,idempotent_key,version,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,NOW(),NOW())`,
		orderID, bizNo, userID, "PENDING", 10000, "idem-"+orderID, 0)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO order_items(order_id,sku_id,quantity,unit_price,subtotal,created_at)
		VALUES(?,?,?,?,?,NOW())`,
		orderID, skuID, 2, 5000, 10000)
	require.NoError(t, err)

	reservedID := fmt.Sprintf("res-%d", time.Now().UnixNano())
	_, err = db.ExecContext(ctx,
		`INSERT INTO inventory_reserved(reserved_id,order_id,status,created_at,updated_at)
		VALUES(?,?,?,NOW(),NOW())`,
		reservedID, orderID, "RESERVED")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO inventory_reserved_item(reserved_id,sku_id,quantity,created_at)
		VALUES(?,?,?,NOW())`,
		reservedID, skuID, 2)
	require.NoError(t, err)

	VerifyInventory(t, db, skuID, 18, 2)

	paymentID := fmt.Sprintf("pay-%d", time.Now().UnixNano())
	_, err = db.ExecContext(ctx,
		`INSERT INTO payments(payment_id,order_id,amount,status,request_id,created_at,updated_at)
		VALUES(?,?,?,?,?,NOW(),NOW())`,
		paymentID, orderID, 10000, "PENDING", "req-"+paymentID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`UPDATE payments SET status = 'FAILED', updated_at = NOW() WHERE payment_id = ?`,
		paymentID)
	require.NoError(t, err)

	VerifyPaymentStatus(t, db, paymentID, "FAILED")

	_, err = db.ExecContext(ctx,
		`UPDATE inventory_reserved SET status = 'RELEASED', updated_at = NOW() WHERE reserved_id = ?`,
		reservedID)
	require.NoError(t, err)

	VerifyInventory(t, db, skuID, 20, 0)
}

func TestIdempotentOrderCreation(t *testing.T) {
	suite, cleanup := SetupSuite(t)
	defer cleanup()

	db := suite.MySQL
	ctx := context.Background()

	userID := SetupUser(t, db, "testuser-idemp", "buyer")
	skuID := "SKU-TEST-004"
	SetupInventory(t, db, skuID, 100, 0)

	idempotentKey := "idem-unique-123"

	orderID1 := fmt.Sprintf("order-%d", time.Now().UnixNano())
	bizNo1 := fmt.Sprintf("BIZ-%d", time.Now().UnixNano())

	res1, err := db.ExecContext(ctx,
		`INSERT INTO orders(order_id,biz_no,user_id,status,total_amount,idempotent_key,version,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,NOW(),NOW())`,
		orderID1, bizNo1, userID, "PENDING", 5000, idempotentKey, 0)
	require.NoError(t, err)
	rows1, _ := res1.RowsAffected()

	orderID2 := fmt.Sprintf("order-%d", time.Now().UnixNano()+1)
	bizNo2 := fmt.Sprintf("BIZ-%d", time.Now().UnixNano()+1)

	res2, err := db.ExecContext(ctx,
		`INSERT INTO orders(order_id,biz_no,user_id,status,total_amount,idempotent_key,version,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,NOW(),NOW())`,
		orderID2, bizNo2, userID, "PENDING", 5000, idempotentKey, 0)
	require.NoError(t, err)
	rows2, _ := res2.RowsAffected()

	assert.Equal(t, int64(1), rows1+rows2, "Only one order should be created with same idempotent key")

	var count int
	db.GetContext(ctx, &count, "SELECT COUNT(*) FROM orders WHERE idempotent_key = ?", idempotentKey)
	assert.Equal(t, 1, count, "Only one order with idempotent key should exist")
}

func TestRefundFlow(t *testing.T) {
	suite, cleanup := SetupSuite(t)
	defer cleanup()

	db := suite.MySQL
	ctx := context.Background()

	userID := SetupUser(t, db, "testuser-refund", "buyer")
	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())
	bizNo := fmt.Sprintf("BIZ-%d", time.Now().UnixNano())

	_, err := db.ExecContext(ctx,
		`INSERT INTO orders(order_id,biz_no,user_id,status,total_amount,idempotent_key,version,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,NOW(),NOW())`,
		orderID, bizNo, userID, "PAID", 10000, "idem-"+orderID, 0)
	require.NoError(t, err)

	paymentID := fmt.Sprintf("pay-%d", time.Now().UnixNano())
	_, err = db.ExecContext(ctx,
		`INSERT INTO payments(payment_id,order_id,amount,status,request_id,created_at,updated_at)
		VALUES(?,?,?,?,?,NOW(),NOW())`,
		paymentID, orderID, 10000, "SUCCESS", "req-"+paymentID)
	require.NoError(t, err)

	refundID := fmt.Sprintf("refund-%d", time.Now().UnixNano())
	_, err = db.ExecContext(ctx,
		`INSERT INTO refunds(refund_id,order_id,payment_id,amount,status,request_id,created_at,updated_at)
		VALUES(?,?,?,?,?,?,NOW(),NOW())`,
		refundID, orderID, paymentID, 10000, "PENDING", "req-"+refundID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`UPDATE refunds SET status = 'SUCCESS', updated_at = NOW() WHERE refund_id = ?`,
		refundID)
	require.NoError(t, err)

	var status string
	db.GetContext(ctx, &status, "SELECT status FROM refunds WHERE refund_id = ?", refundID)
	assert.Equal(t, "SUCCESS", status)
}

func TestInventoryReservation(t *testing.T) {
	suite, cleanup := SetupSuite(t)
	defer cleanup()

	db := suite.MySQL
	ctx := context.Background()

	skuID := "SKU-TEST-005"
	SetupInventory(t, db, skuID, 100, 0)

	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())
	reservedID := fmt.Sprintf("res-%d", time.Now().UnixNano())

	_, err := db.ExecContext(ctx,
		`INSERT INTO inventory_reserved(reserved_id,order_id,status,created_at,updated_at)
		VALUES(?,?,?,NOW(),NOW())`,
		reservedID, orderID, "RESERVED")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO inventory_reserved_item(reserved_id,sku_id,quantity,created_at)
		VALUES(?,?,?,NOW())`,
		reservedID, skuID, 10)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`UPDATE inventory SET reserved = reserved + 10, updated_at = NOW() WHERE sku_id = ?`,
		skuID)
	require.NoError(t, err)

	VerifyInventory(t, db, skuID, 90, 10)

	_, err = db.ExecContext(ctx,
		`UPDATE inventory_reserved SET status = 'CONFIRMED', updated_at = NOW() WHERE reserved_id = ?`,
		reservedID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`UPDATE inventory SET stock = stock - 10, reserved = reserved - 10, updated_at = NOW() WHERE sku_id = ?`,
		skuID)
	require.NoError(t, err)

	VerifyInventory(t, db, skuID, 90, 0)
}

func TestOutboxMessagePublishing(t *testing.T) {
	suite, cleanup := SetupSuite(t)
	defer cleanup()

	db := suite.MySQL
	ctx := context.Background()

	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())
	payload, _ := json.Marshal(map[string]string{
		"order_id": orderID,
		"event":    "order_created",
	})

	_, err := db.ExecContext(ctx,
		`INSERT INTO order_outbox(event_type,payload,status,retry_count,last_error,created_at)
		VALUES(?,?,?,?,?,NOW())`,
		"order_created", payload, "PENDING", 0, "")
	require.NoError(t, err)

	pending := CountOutboxPending(t, db)
	assert.GreaterOrEqual(t, pending, 1)

	var msgPayload string
	db.GetContext(ctx, &msgPayload, "SELECT payload FROM order_outbox WHERE status = 'PENDING' LIMIT 1")
	assert.NotEmpty(t, msgPayload)
}
