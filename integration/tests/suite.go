package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

type IntegrationSuite struct {
	MySQL     *sqlx.DB
	Redis     *redis.Client
	RabbitMQ  *rabbitmq.Container
	Cleanup   func()
	RequestID string
}

func SetupSuite(t *testing.T) (*IntegrationSuite, func()) {
	ctx := context.Background()
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3308)/go_micro_integration?parseTime=true&charset=utf8mb4&loc=Local"
	}

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		t.Skipf("MySQL not available: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       15,
	})
	if rdb.Addr() == ":6379" {
		rdb = redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:6380",
		})
	}
	if err := rdb.Ping(ctx).Err(); err != nil {
		db.Close()
		t.Skipf("Redis not available: %v", err)
	}

	teardown := func() {
		cleanupDB(db)
		db.Close()
		rdb.Close()
	}

	return &IntegrationSuite{
		MySQL:    db,
		Redis:    rdb,
		Cleanup:  teardown,
		RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()),
	}, teardown
}

func cleanupDB(db *sqlx.DB) {
	tables := []string{
		"order_outbox",
		"order_items",
		"orders",
		"payments",
		"refunds",
		"inventory_reserved_item",
		"inventory_reserved",
		"inventory",
		"users",
		"activity_coupons",
		"activity_seckill",
		"seckill_orders",
		"saga_steps",
		"sagas",
		"tasks",
		"price_history",
	}
	ctx := context.Background()
	for _, table := range tables {
		db.ExecContext(ctx, "DELETE FROM "+table)
	}
}

func SetupUser(t *testing.T, db *sqlx.DB, username, role string) string {
	ctx := context.Background()
	var count int
	db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE username = ?", username)
	if count > 0 {
		var uid string
		db.GetContext(ctx, &uid, "SELECT user_id FROM users WHERE username = ?", username)
		return uid
	}

	userID := fmt.Sprintf("user-%d", time.Now().UnixNano())
	pwdHash := "$2a$10$N9qo8uLOickgx2ZMRZoMye0IqrQ3EQkLqgBJXCn1pYzN9pq7p9XKa"

	_, err := db.ExecContext(ctx,
		"INSERT INTO users(user_id,username,password_hash,role,status,created_at,updated_at) VALUES(?,?,?,?,?,NOW(),NOW())",
		userID, username, pwdHash, role, 1)
	require.NoError(t, err)
	return userID
}

func SetupInventory(t *testing.T, db *sqlx.DB, skuID string, stock, reserved int) {
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO inventory(sku_id, stock, reserved, version, created_at, updated_at)
		VALUES(?, ?, ?, 0, NOW(), NOW())
		ON DUPLICATE KEY UPDATE stock=VALUES(stock), reserved=VALUES(reserved)`,
		skuID, stock, reserved)
	require.NoError(t, err)
}

func VerifyInventory(t *testing.T, db *sqlx.DB, skuID string, expectedStock, expectedReserved int) {
	ctx := context.Background()
	var inv struct {
		Stock    int `db:"stock"`
		Reserved int `db:"reserved"`
	}
	err := db.GetContext(ctx, &inv, "SELECT stock, reserved FROM inventory WHERE sku_id = ?", skuID)
	require.NoError(t, err)
	assert.Equal(t, expectedStock, inv.Stock)
	assert.Equal(t, expectedReserved, inv.Reserved)
}

func VerifyOrderStatus(t *testing.T, db *sqlx.DB, orderID, expectedStatus string) {
	ctx := context.Background()
	var status string
	err := db.GetContext(ctx, &status, "SELECT status FROM orders WHERE order_id = ?", orderID)
	require.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
}

func VerifyPaymentStatus(t *testing.T, db *sqlx.DB, paymentID, expectedStatus string) {
	ctx := context.Background()
	var status string
	err := db.GetContext(ctx, &status, "SELECT status FROM payments WHERE payment_id = ?", paymentID)
	require.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
}

func CountOutboxPending(t *testing.T, db *sqlx.DB) int {
	ctx := context.Background()
	var count int
	err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM order_outbox WHERE status = 'PENDING'")
	require.NoError(t, err)
	return count
}

func PublishEvent(t *testing.T, rdb *redis.Client, topic string, payload any) {
	ctx := context.Background()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	err = rdb.Publish(ctx, topic, data).Err()
	require.NoError(t, err)
}
