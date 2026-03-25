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
	"github.com/redis/go-redis/v9"
	"go-micro/pkg/cache"
	"go-micro/pkg/db"
)

var (
	ErrIdempotentHit = errors.New("idempotent hit")
	ErrInventoryFail = errors.New("inventory reserve failed")
	ErrNotFound      = errors.New("order not found")
)

const (
	statusPending = "PENDING"
	statusCreated = "CREATED"
	statusFailed  = "FAILED"

	outboxPending = "PENDING"
	outboxSent    = "SENT"
)

type Service struct {
	ctx       context.Context
	db        *sqlx.DB
	cache     *cacheClient
	invClient InventoryClient
	publisher Publisher
}

type InventoryClient interface {
	Reserve(ctx context.Context, orderID string, items []Item) (string, error)
}

type Publisher interface {
	Publish(ctx context.Context, body []byte) error
}

func NewService(dbx *sqlx.DB, rdb *redis.Client, invClient InventoryClient, publisher Publisher) *Service {
	return &Service{
		ctx:       context.Background(),
		db:        dbx,
		cache:     &cacheClient{rdb: rdb},
		invClient: invClient,
		publisher: publisher,
	}
}

func (s *Service) Create(req CreateOrderRequest) (CreateOrderResponse, error) {
	if req.RequestID == "" || req.UserID == "" || len(req.Items) == 0 {
		return CreateOrderResponse{}, errors.New("invalid request")
	}

	orderID := uuid.NewString()
	bizNo := "BIZ-" + uuid.NewString()
	total := int64(0)
	for _, it := range req.Items {
		total += int64(it.Quantity) * it.Price
	}

	insertErr := db.Tx(s.db, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO orders(order_id,biz_no,user_id,status,total_amount,idempotent_key,reserved_id,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,NOW(),NOW())`,
			orderID, bizNo, req.UserID, statusPending, total, req.RequestID, "",
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

	reservedID, err := s.reserveInventory(orderID, req.Items)
	if err != nil {
		_ = s.updateStatus(orderID, statusFailed, "")
		return CreateOrderResponse{}, ErrInventoryFail
	}

	event := OrderCreatedEvent{OrderID: orderID, BizNo: bizNo, Status: statusCreated, UserID: req.UserID, ReservedID: reservedID}
	txErr := db.Tx(s.db, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec(`UPDATE orders SET status=?, reserved_id=?, updated_at=NOW() WHERE order_id=?`, statusCreated, reservedID, orderID); err != nil {
			return err
		}
		payload, _ := json.Marshal(event)
		_, err := tx.Exec(`INSERT INTO order_outbox(event_type,payload,status,created_at) VALUES(?,?,?,NOW())`, "order.created", payload, outboxPending)
		return err
	})
	if txErr != nil {
		return CreateOrderResponse{}, txErr
	}

	resp := CreateOrderResponse{OrderID: orderID, BizNo: bizNo, Status: statusCreated}
	_ = s.cache.setOrder(s.ctx, orderID, resp, req.Items)
	return resp, nil
}

func (s *Service) Get(orderID string) (*Order, error) {
	if orderID == "" {
		return nil, ErrNotFound
	}
	if ord, ok := s.cache.getOrder(s.ctx, orderID); ok {
		return ord, nil
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

func (s *Service) updateStatus(orderID, status, reservedID string) error {
	_, err := s.db.Exec(`UPDATE orders SET status=?, reserved_id=?, updated_at=NOW() WHERE order_id=?`, status, reservedID, orderID)
	return err
}

func (s *Service) reserveInventory(orderID string, items []Item) (string, error) {
	if s.invClient == nil {
		return "", ErrInventoryFail
	}
	ctx, cancel := context.WithTimeout(s.ctx, 3*time.Second)
	defer cancel()
	reservedID, err := s.invClient.Reserve(ctx, orderID, items)
	if err != nil {
		return "", err
	}
	return reservedID, nil
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
	if s.publisher == nil {
		return nil
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
			_ = tx.Rollback()
			return err
		}
		if _, err := tx.Exec(`UPDATE order_outbox SET status=?, sent_at=NOW() WHERE id=?`, outboxSent, r.ID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) StartOutboxPublisher(stop <-chan struct{}) {
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

func ttlWithJitter(base, jitter time.Duration) time.Duration {
	return base + time.Duration(time.Now().UnixNano()%int64(jitter))
}
