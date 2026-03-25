package inventory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go-micro/pkg/cache"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go.uber.org/zap"
)

var (
	ErrInsufficient = errors.New("insufficient inventory")
	ErrNotFound     = errors.New("reservation not found")
	ErrInvalidState = errors.New("invalid reservation state")
	ErrSkuNotFound  = errors.New("sku not found")
	ErrResvNotFound = errors.New("reservation by order not found")
)

const (
	resvReserved  = "RESERVED"
	resvReleased  = "RELEASED"
	resvConfirmed = "CONFIRMED"
)

type Service struct {
	ctx   context.Context
	db    *sqlx.DB
	cache *cacheClient
}

func NewService(dbx *sqlx.DB, rdb *redis.Client) *Service {
	return &Service{
		ctx:   context.Background(),
		db:    dbx,
		cache: &cacheClient{rdb: rdb},
	}
}

func (s *Service) Reserve(req ReserveRequest) (ReserveResponse, error) {
	if req.OrderID == "" || len(req.Items) == 0 {
		return ReserveResponse{}, errors.New("invalid request")
	}

	reservedID := uuid.NewString()
	err := db.Tx(s.db, func(tx *sqlx.Tx) error {
		for _, it := range req.Items {
			var inv Inventory
			if err := tx.Get(&inv, `SELECT * FROM inventory WHERE sku_id = ? FOR UPDATE`, it.SkuID); err != nil {
				if err == sql.ErrNoRows {
					return ErrInsufficient
				}
				return err
			}
			if inv.Available < it.Quantity {
				return ErrInsufficient
			}
		}

		for _, it := range req.Items {
			_, err := tx.Exec(`UPDATE inventory SET available = available - ?, reserved = reserved + ?, updated_at = NOW() WHERE sku_id = ?`, it.Quantity, it.Quantity, it.SkuID)
			if err != nil {
				return err
			}
		}

		_, err := tx.Exec(`INSERT INTO inventory_reserved(reserved_id,order_id,status,created_at,updated_at) VALUES(?,?,?,NOW(),NOW())`, reservedID, req.OrderID, resvReserved)
		if err != nil {
			return err
		}
		for _, it := range req.Items {
			_, err := tx.Exec(`INSERT INTO inventory_reserved_item(reserved_id,sku_id,quantity) VALUES(?,?,?)`, reservedID, it.SkuID, it.Quantity)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ReserveResponse{}, err
	}

	for _, it := range req.Items {
		_ = s.cache.refreshInventory(s.ctx, s.db, it.SkuID)
	}

	return ReserveResponse{ReservedID: reservedID}, nil
}

func (s *Service) Release(reservedID string) error {
	alreadyReleased := false
	err := db.Tx(s.db, func(tx *sqlx.Tx) error {
		resv := Reservation{}
		if err := tx.Get(&resv, `SELECT * FROM inventory_reserved WHERE reserved_id = ? FOR UPDATE`, reservedID); err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}
			return err
		}
		if resv.Status == resvReleased {
			alreadyReleased = true
			return nil
		}
		if resv.Status != resvReserved {
			return ErrInvalidState
		}

		var items []Item
		if err := tx.Select(&items, `SELECT sku_id,quantity FROM inventory_reserved_item WHERE reserved_id = ?`, reservedID); err != nil {
			return err
		}
		for _, it := range items {
			_, err := tx.Exec(`UPDATE inventory SET available = available + ?, reserved = reserved - ?, updated_at = NOW() WHERE sku_id = ?`, it.Quantity, it.Quantity, it.SkuID)
			if err != nil {
				return err
			}
		}

		_, err := tx.Exec(`UPDATE inventory_reserved SET status=?, updated_at=NOW() WHERE reserved_id = ?`, resvReleased, reservedID)
		return err
	})
	if err != nil {
		return err
	}
	if alreadyReleased {
		return nil
	}

	var items []Item
	if err := s.db.Select(&items, `SELECT sku_id,quantity FROM inventory_reserved_item WHERE reserved_id = ?`, reservedID); err != nil {
		return err
	}
	for _, it := range items {
		_ = s.cache.refreshInventory(s.ctx, s.db, it.SkuID)
	}
	return nil
}

func (s *Service) ReleaseByOrder(orderID string) error {
	if orderID == "" {
		return ErrNotFound
	}
	resv, err := s.GetReservation(orderID)
	if err != nil {
		return err
	}
	return s.Release(resv.ReservedID)
}

func (s *Service) Confirm(reservedID string) error {
	err := db.Tx(s.db, func(tx *sqlx.Tx) error {
		resv := Reservation{}
		if err := tx.Get(&resv, `SELECT * FROM inventory_reserved WHERE reserved_id = ? FOR UPDATE`, reservedID); err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}
			return err
		}
		if resv.Status != resvReserved {
			return ErrInvalidState
		}
		_, err := tx.Exec(`UPDATE inventory_reserved SET status=?, updated_at=NOW() WHERE reserved_id = ?`, resvConfirmed, reservedID)
		return err
	})
	if err != nil {
		return err
	}
	var items []Item
	if err := s.db.Select(&items, `SELECT sku_id,quantity FROM inventory_reserved_item WHERE reserved_id = ?`, reservedID); err != nil {
		return err
	}
	for _, it := range items {
		_ = s.cache.refreshInventory(s.ctx, s.db, it.SkuID)
	}
	return nil
}

func (s *Service) GetInventory(skuID string) (*Inventory, error) {
	if skuID == "" {
		return nil, ErrSkuNotFound
	}
	if inv, ok := s.cache.getInventory(s.ctx, skuID); ok {
		return inv, nil
	}

	lockKey := "lock:inv:" + skuID
	locked, _ := cache.TryLock(s.ctx, s.cache.rdb, lockKey, 5*time.Second)
	if locked {
		defer func() { _ = cache.Unlock(s.ctx, s.cache.rdb, lockKey) }()
		if inv, ok := s.cache.getInventory(s.ctx, skuID); ok {
			return inv, nil
		}
	}

	inv := Inventory{}
	if err := s.db.Get(&inv, `SELECT * FROM inventory WHERE sku_id = ?`, skuID); err != nil {
		if err == sql.ErrNoRows {
			_ = s.cache.setNil(s.ctx, skuID)
			return nil, ErrSkuNotFound
		}
		return nil, err
	}
	_ = s.cache.setInventory(s.ctx, skuID, inv)
	return &inv, nil
}

func (s *Service) GetReservation(orderID string) (*Reservation, error) {
	if orderID == "" {
		return nil, ErrResvNotFound
	}
	resv := Reservation{}
	if err := s.db.Get(&resv, `SELECT * FROM inventory_reserved WHERE order_id = ? ORDER BY created_at DESC LIMIT 1`, orderID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrResvNotFound
		}
		return nil, err
	}
	return &resv, nil
}

type cacheClient struct {
	rdb *redis.Client
}

func (c *cacheClient) getInventory(ctx context.Context, skuID string) (*Inventory, bool) {
	key := "inventory:" + skuID
	// Check nil cache
	nilKey := "inventory:nil:" + skuID
	if val, err := c.rdb.Get(ctx, nilKey).Result(); err == nil && val == cache.NilValue {
		cache.IncHit("inventory")
		return nil, false
	}
	fields, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		logx.L().Warn("redis hgetall failed", zap.String("key", key), zap.Error(err))
		if v, ok, localNil := cache.GetLocalString(key); ok {
			cache.IncHit("inventory")
			if localNil {
				return nil, false
			}
			var inv Inventory
			if err := json.Unmarshal([]byte(v), &inv); err == nil && inv.SkuID != "" {
				return &inv, true
			}
		}
		cache.IncMiss("inventory")
		return nil, false
	}
	if len(fields) == 0 {
		cache.IncMiss("inventory")
		return nil, false
	}
	available, _ := strconv.Atoi(fields["available"])
	reserved, _ := strconv.Atoi(fields["reserved"])
	inv := Inventory{
		SkuID:     fields["sku_id"],
		Available: available,
		Reserved:  reserved,
	}
	if inv.SkuID == "" {
		cache.IncMiss("inventory")
		return nil, false
	}
	cache.IncHit("inventory")
	return &inv, true
}

func (c *cacheClient) setInventory(ctx context.Context, skuID string, inv Inventory) error {
	key := "inventory:" + skuID
	fields := map[string]interface{}{
		"sku_id":    inv.SkuID,
		"available": strconv.Itoa(inv.Available),
		"reserved":  strconv.Itoa(inv.Reserved),
		"version":   strconv.FormatInt(time.Now().UnixNano(), 10),
	}
	_ = c.rdb.Del(ctx, "inventory:nil:"+skuID).Err()
	_ = c.rdb.HSet(ctx, key, fields).Err()
	_ = c.rdb.Expire(ctx, key, ttlWithJitter(2*time.Minute, 20*time.Second)).Err()
	if b, err := json.Marshal(inv); err == nil {
		cache.SetLocalString(key, string(b), ttlWithJitter(2*time.Minute, 20*time.Second), false)
	}
	return nil
}

func (c *cacheClient) setNil(ctx context.Context, skuID string) error {
	key := "inventory:nil:" + skuID
	return cache.SetNil(ctx, c.rdb, key, ttlWithJitter(30*time.Second, 10*time.Second))
}

func (c *cacheClient) delInventory(ctx context.Context, skuID string) error {
	key := "inventory:" + skuID
	_ = c.rdb.Del(ctx, key).Err()
	_ = c.rdb.Del(ctx, "inventory:nil:"+skuID).Err()
	cache.SetLocalString(key, "", 1*time.Nanosecond, true)
	return nil
}

func (c *cacheClient) refreshInventory(ctx context.Context, dbx *sqlx.DB, skuID string) error {
	inv := Inventory{}
	if err := dbx.Get(&inv, `SELECT * FROM inventory WHERE sku_id = ?`, skuID); err != nil {
		return err
	}
	return c.setInventory(ctx, skuID, inv)
}

func ttlWithJitter(base, jitter time.Duration) time.Duration {
	return base + time.Duration(time.Now().UnixNano()%int64(jitter))
}
