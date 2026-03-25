package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go-micro/pkg/cache"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/resilience"
)

var (
	ErrIdempotentHit = errors.New("idempotent hit")
	ErrInventoryFail = errors.New("inventory reserve failed")
	ErrNotFound      = errors.New("order not found")
	ErrInvalidState  = errors.New("invalid order state")
	ErrNotCancelable = errors.New("order not cancelable")
)

const (
	statusCreated    = "CREATED"
	statusReserved   = "RESERVED"
	statusProcessing = "PROCESSING"
	statusSuccess    = "SUCCESS"
	statusFailed     = "FAILED"
	statusCanceled   = "CANCELED"

	outboxPending = "PENDING"
	outboxSent    = "SENT"
)

type Service struct {
	ctx       context.Context
	db        *sqlx.DB
	cache     *cacheClient
	invClient InventoryClient
	publisher Publisher
	cbInv     *resilience.CircuitBreaker
}

type InventoryClient interface {
	Reserve(ctx context.Context, orderID string, items []Item) (string, error)
}

type Publisher interface {
	Publish(ctx context.Context, body []byte) error
}

var allowedTransitions = map[string]map[string]bool{
	statusCreated: {
		statusReserved: true,
		statusCanceled: true,
	},
	statusReserved: {
		statusProcessing: true,
		statusFailed:     true,
		statusCanceled:   true,
	},
	statusProcessing: {
		statusSuccess: true,
		statusFailed:  true,
	},
}

var outboxPendingGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "order",
	Subsystem: "outbox",
	Name:      "pending_total",
	Help:      "Pending outbox messages",
})

func init() {
	prometheus.MustRegister(outboxPendingGauge)
}

func NewService(dbx *sqlx.DB, rdb *redis.Client, invClient InventoryClient, publisher Publisher) *Service {
	return &Service{
		ctx:       context.Background(),
		db:        dbx,
		cache:     &cacheClient{rdb: rdb},
		invClient: invClient,
		publisher: publisher,
		cbInv:     newBreakerFromEnv(),
	}
}

func (s *Service) Create(req CreateOrderRequest) (CreateOrderResponse, error) {
	if req.RequestID == "" || req.UserID == "" || len(req.Items) == 0 {
		return CreateOrderResponse{}, errors.New("invalid request")
	}

	// Idempotency: request_id is unique, return existing order on duplicate.
	orderID := uuid.NewString()
	bizNo := "BIZ-" + uuid.NewString()
	total := int64(0)
	for _, it := range req.Items {
		total += int64(it.Quantity) * it.Price
	}

	insertErr := db.Tx(s.db, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO orders(order_id,biz_no,user_id,status,total_amount,idempotent_key,reserved_id,version,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,NOW(),NOW())`,
			orderID, bizNo, req.UserID, statusCreated, total, req.RequestID, "", 0,
		)
		if err != nil {
			return err
		}
		for _, it := range req.Items {
			_, err := tx.Exec(`INSERT INTO order_items(order_id,sku_id,quantity,price) VALUES(?,?,?,?)`, orderID, it.SkuID, it.Quantity, it.Price)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if insertErr != nil {
		if isDuplicate(insertErr) {
			order, err := s.getByIdempotentKey(req.RequestID)
			if err == nil {
				return CreateOrderResponse{OrderID: order.OrderID, BizNo: order.BizNo, Status: order.Status}, ErrIdempotentHit
			}
		}
		return CreateOrderResponse{}, insertErr
	}

	// Reserve inventory synchronously; failure marks order FAILED.
	reservedID, err := s.reserveInventory(orderID, req.Items)
	if err != nil {
		_ = s.updateStatusWithVersion(orderID, statusCreated, statusFailed, "", 0)
		return CreateOrderResponse{}, ErrInventoryFail
	}

	event := OrderCreatedEvent{OrderID: orderID, BizNo: bizNo, Status: statusReserved, UserID: req.UserID, ReservedID: reservedID}
	txErr := db.Tx(s.db, func(tx *sqlx.Tx) error {
		// Transactionally update order status and persist outbox event.
		if err := updateStatusTx(tx, orderID, statusCreated, statusReserved, reservedID, 0); err != nil {
			return err
		}
		payload, _ := json.Marshal(event)
		_, err := tx.Exec(`INSERT INTO order_outbox(event_type,payload,status,retry_count,last_error,created_at) VALUES(?,?,?,?,?,NOW())`, "order_reserved", payload, outboxPending, 0, "")
		return err
	})
	if txErr != nil {
		return CreateOrderResponse{}, txErr
	}

	resp := CreateOrderResponse{OrderID: orderID, BizNo: bizNo, Status: statusReserved}
	_ = s.cache.setOrder(s.ctx, orderID, resp, req.Items)
	_ = s.cache.setOrderByBizNo(s.ctx, bizNo, &Order{OrderID: orderID, BizNo: bizNo, UserID: req.UserID, Status: statusReserved, TotalAmount: total, Items: req.Items})
	return resp, nil
}

func (s *Service) Get(orderID string) (*Order, error) {
	if orderID == "" {
		return nil, ErrNotFound
	}
	if ord, ok := s.cache.getOrder(s.ctx, orderID); ok {
		return ord, nil
	}

	// Cache breakdown protection: singleflight via redis lock
	lockKey := "lock:order:" + orderID
	locked, _ := cache.TryLock(s.ctx, s.cache.rdb, lockKey, 5*time.Second)
	if locked {
		defer func() { _ = cache.Unlock(s.ctx, s.cache.rdb, lockKey) }()
		if ord, ok := s.cache.getOrder(s.ctx, orderID); ok {
			return ord, nil
		}
	}

	order := Order{}
	err := s.db.Get(&order, `SELECT * FROM orders WHERE order_id = ?`, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			_ = s.cache.setNil(s.ctx, orderID)
			return nil, ErrNotFound
		}
		return nil, err
	}

	var items []Item
	if err := s.db.Select(&items, `SELECT sku_id,quantity,price FROM order_items WHERE order_id = ?`, orderID); err != nil {
		return nil, err
	}
	order.Items = items
	_ = s.cache.setOrder(s.ctx, orderID, CreateOrderResponse{OrderID: order.OrderID, BizNo: order.BizNo, Status: order.Status}, items)
	return &order, nil
}

func (s *Service) GetByBizNo(bizNo string) (*Order, error) {
	if bizNo == "" {
		return nil, ErrNotFound
	}
	// Try cache by bizNo
	if ord, ok := s.cache.getOrderByBizNo(s.ctx, bizNo); ok {
		return ord, nil
	}
	lockKey := "lock:orderbiz:" + bizNo
	locked, _ := cache.TryLock(s.ctx, s.cache.rdb, lockKey, 5*time.Second)
	if locked {
		defer func() { _ = cache.Unlock(s.ctx, s.cache.rdb, lockKey) }()
		if ord, ok := s.cache.getOrderByBizNo(s.ctx, bizNo); ok {
			return ord, nil
		}
	}
	order := Order{}
	if err := s.db.Get(&order, `SELECT * FROM orders WHERE biz_no = ?`, bizNo); err != nil {
		if err == sql.ErrNoRows {
			_ = s.cache.setNilByBizNo(s.ctx, bizNo)
			return nil, ErrNotFound
		}
		return nil, err
	}
	var items []Item
	if err := s.db.Select(&items, `SELECT sku_id,quantity,price FROM order_items WHERE order_id = ?`, order.OrderID); err != nil {
		return nil, err
	}
	order.Items = items
	_ = s.cache.setOrder(s.ctx, order.OrderID, CreateOrderResponse{OrderID: order.OrderID, BizNo: order.BizNo, Status: order.Status}, items)
	_ = s.cache.setOrderByBizNo(s.ctx, bizNo, &order)
	return &order, nil
}

func (s *Service) UpdateStatus(orderID, from, to string) error {
	if orderID == "" {
		return ErrNotFound
	}
	var status string
	var version int64
	if err := s.db.QueryRow(`SELECT status,version FROM orders WHERE order_id = ?`, orderID).Scan(&status, &version); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	if status == to {
		return nil
	}
	if status != from {
		return ErrInvalidState
	}
	if err := s.updateStatusWithVersion(orderID, from, to, "", version); err != nil {
		return err
	}
	_ = s.cache.delOrder(s.ctx, orderID)
	_ = s.cache.delOrderBizIndex(s.ctx, orderID)
	return nil
}

func (s *Service) Cancel(orderID string) error {
	if orderID == "" {
		return ErrNotFound
	}
	var status string
	var version int64
	if err := s.db.QueryRow(`SELECT status,version FROM orders WHERE order_id = ?`, orderID).Scan(&status, &version); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	if status == statusCanceled {
		return nil
	}
	if status == statusSuccess {
		return ErrNotCancelable
	}
	if status != statusCreated && status != statusReserved {
		return ErrNotCancelable
	}
	if err := s.updateStatusWithVersion(orderID, status, statusCanceled, "", version); err != nil {
		return err
	}
	_ = s.cache.delOrder(s.ctx, orderID)
	_ = s.cache.delOrderBizIndex(s.ctx, orderID)
	return nil
}

func (s *Service) getByIdempotentKey(key string) (*Order, error) {
	order := Order{}
	err := s.db.Get(&order, `SELECT * FROM orders WHERE idempotent_key = ?`, key)
	if err != nil {
		return nil, err
	}
	var items []Item
	_ = s.db.Select(&items, `SELECT sku_id,quantity,price FROM order_items WHERE order_id = ?`, order.OrderID)
	order.Items = items
	return &order, nil
}

func (s *Service) updateStatusWithVersion(orderID, from, to, reservedID string, version int64) error {
	if !canTransition(from, to) {
		return ErrInvalidState
	}
	res, err := s.db.Exec(`UPDATE orders SET status=?, reserved_id=?, version=version+1, updated_at=NOW() WHERE order_id=? AND status=? AND version=?`, to, reservedID, orderID, from, version)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrInvalidState
	}
	return nil
}

func updateStatusTx(tx *sqlx.Tx, orderID, from, to, reservedID string, version int64) error {
	if !canTransition(from, to) {
		return ErrInvalidState
	}
	res, err := tx.Exec(`UPDATE orders SET status=?, reserved_id=?, version=version+1, updated_at=NOW() WHERE order_id=? AND status=? AND version=?`, to, reservedID, orderID, from, version)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrInvalidState
	}
	return nil
}

func canTransition(from, to string) bool {
	if from == to {
		return true
	}
	nexts, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	return nexts[to]
}

func (s *Service) reserveInventory(orderID string, items []Item) (string, error) {
	if s.invClient == nil {
		return "", ErrInventoryFail
	}
	// Remote reserve call with circuit breaker and timeout.
	ctx, cancel := context.WithTimeout(s.ctx, 3*time.Second)
	defer cancel()
	var reservedID string
	err := s.cbInv.Execute(func() error {
		var callErr error
		reservedID, callErr = s.invClient.Reserve(ctx, orderID, items)
		return callErr
	})
	if err != nil {
		return "", err
	}
	return reservedID, nil
}

func newBreakerFromEnv() *resilience.CircuitBreaker {
	fail := getInt("CB_FAIL_THRESHOLD", 5)
	reset := getInt("CB_RESET_SECONDS", 10)
	half := getInt("CB_HALF_OPEN_SUCCESS", 1)
	return resilience.NewCircuitBreaker(fail, time.Duration(reset)*time.Second, half)
}

func getInt(key string, def int) int {
	v := config.GetEnv(key, "")
	if v == "" {
		return def
	}
	n := 0
	for i := 0; i < len(v); i++ {
		ch := v[i]
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	if n == 0 {
		return def
	}
	return n
}

func isDuplicate(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return false
}

type cacheClient struct {
	rdb *redis.Client
}

type OrderCreatedEvent struct {
	OrderID    string `json:"order_id"`
	BizNo      string `json:"biz_no"`
	Status     string `json:"status"`
	UserID     string `json:"user_id"`
	ReservedID string `json:"reserved_id"`
}

func (s *Service) publishOutboxBatch(limit int) error {
	// Outbox publisher: fetch PENDING rows, publish, then mark SENT.
	if s.publisher == nil {
		return nil
	}

	var pendingCount int64
	if err := s.db.Get(&pendingCount, `SELECT COUNT(1) FROM order_outbox WHERE status = ?`, outboxPending); err == nil {
		outboxPendingGauge.Set(float64(pendingCount))
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	var rows []struct {
		ID      int64           `db:"id"`
		Payload json.RawMessage `db:"payload"`
	}
	if err := tx.Select(&rows, `SELECT id,payload FROM order_outbox WHERE status = ? ORDER BY id ASC LIMIT ? FOR UPDATE`, outboxPending, limit); err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, r := range rows {
		if err := s.publisher.Publish(s.ctx, r.Payload); err != nil {
			_, _ = tx.Exec(`UPDATE order_outbox SET retry_count = retry_count + 1, last_error = ? WHERE id = ?`, err.Error(), r.ID)
			continue
		}
		if _, err := tx.Exec(`UPDATE order_outbox SET status=?, sent_at=NOW() WHERE id=?`, outboxSent, r.ID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) StartOutboxPublisher(stop <-chan struct{}) {
	// Background worker for reliable event delivery.
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			_ = s.publishOutboxBatch(50)
		}
	}
}

func (c *cacheClient) getOrder(ctx context.Context, orderID string) (*Order, bool) {
	key := "order:" + orderID
	var ord Order
	hit, isNil, err := cache.GetJSON(ctx, c.rdb, key, &ord)
	if err != nil || isNil || !hit {
		return nil, false
	}
	if ord.OrderID == "" {
		return nil, false
	}
	return &ord, true
}

func (c *cacheClient) setOrder(ctx context.Context, orderID string, resp CreateOrderResponse, items []Item) error {
	key := "order:" + orderID
	ord := Order{OrderID: resp.OrderID, BizNo: resp.BizNo, Status: resp.Status, Items: items}
	return cache.SetJSON(ctx, c.rdb, key, ord, ttlWithJitter(3*time.Minute, 30*time.Second))
}

func (c *cacheClient) setNil(ctx context.Context, orderID string) error {
	key := "order:" + orderID
	return cache.SetNil(ctx, c.rdb, key, ttlWithJitter(30*time.Second, 10*time.Second))
}

func (c *cacheClient) delOrder(ctx context.Context, orderID string) error {
	key := "order:" + orderID
	_ = c.rdb.Del(ctx, key).Err()
	return nil
}

func (c *cacheClient) delOrderBizIndex(ctx context.Context, orderID string) error {
	key := "order:biz:index:" + orderID
	bizNo, err := c.rdb.Get(ctx, key).Result()
	if err == nil && bizNo != "" {
		_ = c.rdb.Del(ctx, "order:biz:"+bizNo).Err()
	}
	_ = c.rdb.Del(ctx, key).Err()
	return nil
}

func (c *cacheClient) getOrderByBizNo(ctx context.Context, bizNo string) (*Order, bool) {
	key := "order:biz:" + bizNo
	var ord Order
	hit, isNil, err := cache.GetJSON(ctx, c.rdb, key, &ord)
	if err != nil || isNil || !hit {
		return nil, false
	}
	if ord.OrderID == "" {
		return nil, false
	}
	return &ord, true
}

func (c *cacheClient) setOrderByBizNo(ctx context.Context, bizNo string, ord *Order) error {
	key := "order:biz:" + bizNo
	_ = c.rdb.Set(ctx, "order:biz:index:"+ord.OrderID, bizNo, ttlWithJitter(5*time.Minute, 30*time.Second)).Err()
	return cache.SetJSON(ctx, c.rdb, key, ord, ttlWithJitter(3*time.Minute, 30*time.Second))
}

func (c *cacheClient) setNilByBizNo(ctx context.Context, bizNo string) error {
	key := "order:biz:" + bizNo
	return cache.SetNil(ctx, c.rdb, key, ttlWithJitter(30*time.Second, 10*time.Second))
}

func ttlWithJitter(base, jitter time.Duration) time.Duration {
	return base + time.Duration(time.Now().UnixNano()%int64(jitter))
}
