package mysqlit

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	_ "github.com/go-sql-driver/mysql"
)

const TestDBIndex = 15

type TestDB struct {
	DB  *sqlx.DB
	RDB *redis.Client
}

func getDSN() string {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3307)/go_micro?parseTime=true&charset=utf8mb4&loc=Local"
	}
	return dsn
}

func getRedisAddr() string {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	return addr
}

func NewDB(t *testing.T) (*TestDB, func()) {
	dsn := getDSN()
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		t.Skipf("MySQL not available: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	rdb := redis.NewClient(&redis.Options{
		Addr:     getRedisAddr(),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       TestDBIndex,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		db.Close()
		t.Skipf("Redis not available: %v", err)
	}

	teardown := func() {
		if err := cleanupDB(db); err != nil {
			log.Printf("Cleanup error: %v", err)
		}
		db.Close()
		rdb.Close()
	}

	return &TestDB{DB: db, RDB: rdb}, teardown
}

func cleanupDB(db *sqlx.DB) error {
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
		if _, err := db.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return fmt.Errorf("failed to cleanup %s: %w", table, err)
		}
	}
	return nil
}

func randomID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}

func SetupUser(t *testing.T, db *sqlx.DB, username, role string) (userID string) {
	ctx := context.Background()

	var count int
	db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE username = ?", username)
	if count > 0 {
		var uid string
		db.GetContext(ctx, &uid, "SELECT user_id FROM users WHERE username = ?", username)
		return uid
	}

	userID = fmt.Sprintf("user-%d", time.Now().UnixNano())
	pwdHash := "$2a$10$N9qo8uLOickgx2ZMRZoMye0IqrQ3EQkLqgBJXCn1pYzN9pq7p9XKa"

	_, err := db.ExecContext(ctx,
		"INSERT INTO users(user_id,username,password_hash,role,status,created_at,updated_at) VALUES(?,?,?,?,?,NOW(),NOW())",
		userID, username, pwdHash, role, 1)
	if err != nil {
		t.Fatalf("Failed to setup user: %v", err)
	}
	return
}

func SetupOrder(t *testing.T, db *sqlx.DB, userID, status string) (orderID string) {
	ctx := context.Background()
	orderID = fmt.Sprintf("order-%d", time.Now().UnixNano())
	bizNo := fmt.Sprintf("BIZ-%d", time.Now().UnixNano())

	_, err := db.ExecContext(ctx,
		"INSERT INTO orders(order_id,biz_no,user_id,status,total_amount,idempotent_key,reserved_id,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,NOW(),NOW())",
		orderID, bizNo, userID, status, 10000, "idem-"+orderID, "", 0)
	if err != nil {
		t.Fatalf("Failed to setup order: %v", err)
	}
	return
}

func SetupPayment(t *testing.T, db *sqlx.DB, orderID, status string) (paymentID string) {
	ctx := context.Background()
	paymentID = fmt.Sprintf("pay-%d", time.Now().UnixNano())

	_, err := db.ExecContext(ctx,
		"INSERT INTO payments(payment_id,order_id,amount,status,request_id,created_at,updated_at) VALUES(?,?,?,?,?,NOW(),NOW())",
		paymentID, orderID, 10000, status, "req-"+paymentID)
	if err != nil {
		t.Fatalf("Failed to setup payment: %v", err)
	}
	return
}

func CountRows(t *testing.T, db *sqlx.DB, table string) int {
	var count int
	if err := db.Get(&count, "SELECT COUNT(*) FROM "+table); err != nil {
		t.Fatalf("Failed to count %s: %v", table, err)
	}
	return count
}
