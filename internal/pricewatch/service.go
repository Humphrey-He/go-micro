package pricewatch

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// SetPriceWatch creates or updates a price watch for a user
func (s *Service) SetPriceWatch(ctx context.Context, userID int64, req *SetPriceWatchRequest) (*SetPriceWatchResponse, error) {
	// Get current product info (mock - in production would call product service)
	currentPrice := 0.0

	// Check if watch already exists
	var existing PriceWatch
	err := s.db.GetContext(ctx, &existing, `SELECT * FROM user_price_watches WHERE user_id = ? AND sku_id = ?`, userID, req.SkuID)

	if err == nil {
		// Update existing watch
		targetPrice := req.TargetPrice
		_, err = s.db.ExecContext(ctx, `
			UPDATE user_price_watches
			SET target_price = ?, current_price = ?, status = 'active', updated_at = NOW()
			WHERE user_id = ? AND sku_id = ?`,
			targetPrice, currentPrice, userID, req.SkuID)
		if err != nil {
			return nil, err
		}
		return &SetPriceWatchResponse{
			WatchID:      existing.ID,
			CurrentPrice: currentPrice,
			TargetPrice:  targetPrice,
		}, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new watch
	targetPrice := req.TargetPrice
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO user_price_watches (user_id, sku_id, target_price, current_price, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'active', NOW(), NOW())`,
		userID, req.SkuID, targetPrice, currentPrice)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &SetPriceWatchResponse{
		WatchID:      id,
		CurrentPrice: currentPrice,
		TargetPrice:  targetPrice,
	}, nil
}

// CancelPriceWatch removes a price watch
func (s *Service) CancelPriceWatch(ctx context.Context, userID int64, skuID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE user_price_watches SET status = 'stopped', updated_at = NOW()
		WHERE user_id = ? AND sku_id = ?`,
		userID, skuID)
	return err
}

// UpdatePriceWatch updates watch settings
func (s *Service) UpdatePriceWatch(ctx context.Context, userID int64, skuID string, req *UpdatePriceWatchRequest) error {
	// Build update query dynamically
	query := "UPDATE user_price_watches SET updated_at = NOW()"
	args := []interface{}{}

	if req.TargetPrice != nil {
		query += ", target_price = ?"
		args = append(args, *req.TargetPrice)
	}
	if req.NotifyEnabled != nil {
		query += ", notify_enabled = ?"
		args = append(args, *req.NotifyEnabled)
	}

	query += " WHERE user_id = ? AND sku_id = ? AND status = 'active'"
	args = append(args, userID, skuID)

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// GetPriceWatchList returns user's price watches
func (s *Service) GetPriceWatchList(ctx context.Context, userID int64, page, pageSize int, status string) (*PriceWatchListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Build query
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}

	if status == "active" {
		whereClause += " AND status = 'active'"
	}

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM user_price_watches %s", whereClause)
	if err := s.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, err
	}

	// Get watches
	query := fmt.Sprintf(`
		SELECT w.*, p.name as product_name, p.image
		FROM user_price_watches w
		LEFT JOIN products p ON w.sku_id = p.sku_id
		%s
		ORDER BY w.created_at DESC
		LIMIT ? OFFSET ?`, whereClause)
	args = append(args, pageSize, offset)

	type watchResult struct {
		PriceWatch
		ProductName sql.NullString `db:"product_name"`
		Image       sql.NullString `db:"image"`
	}

	var results []watchResult
	if err := s.db.SelectContext(ctx, &results, query, args...); err != nil {
		return nil, err
	}

	items := make([]PriceWatchItem, len(results))
	for i, r := range results {
		items[i] = PriceWatchItem{
			WatchID:       r.ID,
			SkuID:         r.SkuID,
			ProductName:   r.ProductName.String,
			Image:         r.Image.String,
			CurrentPrice:  r.CurrentPrice,
			OriginalPrice: r.CurrentPrice * 1.2, // Mock
			TargetPrice:   float64Ptr(r.TargetPrice),
			NotifyEnabled: r.NotifyEnabled,
			CreatedAt:     r.CreatedAt.Format("2006-01-02 15:04:05"),
			PriceTrend:    "stable",
		}
	}

	return &PriceWatchListResponse{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// GetPriceHistory returns price history for a product
func (s *Service) GetPriceHistory(ctx context.Context, skuID string, period string) (*PriceHistoryResponse, error) {
	days := 30
	switch period {
	case "90d":
		days = 90
	case "1y":
		days = 365
	}

	query := `
		SELECT price, created_at
		FROM product_price_snapshots
		WHERE sku_id = ? AND snapshot_date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		ORDER BY snapshot_date ASC`

	var snapshots []struct {
		Price     float64   `db:"price"`
		CreatedAt time.Time `db:"created_at"`
	}
	if err := s.db.SelectContext(ctx, &snapshots, query, skuID, days); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// If no snapshots, return mock data
	if len(snapshots) == 0 {
		// Generate mock price points
		now := time.Now()
		points := make([]PricePoint, 30)
		basePrice := 99.0
		for i := 29; i >= 0; i-- {
			date := now.AddDate(0, 0, -i)
			price := basePrice + float64(i%10) - 5
			points[29-i] = PricePoint{
				Date:  date.Format("2006-01-02"),
				Price: price,
			}
		}
		return &PriceHistoryResponse{
			SkuID:        skuID,
			CurrentPrice: points[len(points)-1].Price,
			LowestPrice:  89.0,
			HighestPrice: 109.0,
			AveragePrice: 99.0,
			PricePoints:  points,
		}, nil
	}

	// Calculate stats
	var total, min, max float64
	min = snapshots[0].Price
	max = snapshots[0].Price
	var minDate, maxDate time.Time

	points := make([]PricePoint, len(snapshots))
	for i, s := range snapshots {
		price := s.Price
		total += price
		points[i] = PricePoint{
			Date:  s.CreatedAt.Format("2006-01-02"),
			Price: price,
		}
		if price < min {
			min = price
			minDate = s.CreatedAt
		}
		if price > max {
			max = price
			maxDate = s.CreatedAt
		}
	}

	avg := total / float64(len(snapshots))

	return &PriceHistoryResponse{
		SkuID:        skuID,
		CurrentPrice: snapshots[len(snapshots)-1].Price,
		LowestPrice:  min,
		LowestDate:   minDate.Format("2006-01-02"),
		HighestPrice: max,
		HighestDate:  maxDate.Format("2006-01-02"),
		AveragePrice: avg,
		PricePoints:  points,
	}, nil
}

// GetNotifications returns user's price watch notifications
func (s *Service) GetNotifications(ctx context.Context, userID int64, page, pageSize int) (*NotificationListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Count total and unread
	var total, unread int
	s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM price_watch_notifications WHERE user_id = ?`, userID)
	s.db.GetContext(ctx, &unread, `SELECT COUNT(*) FROM price_watch_notifications WHERE user_id = ? AND is_read = 0`, userID)

	// Get notifications
	query := `
		SELECT n.*, p.name as product_name, p.image
		FROM price_watch_notifications n
		LEFT JOIN products p ON n.sku_id = p.sku_id
		WHERE n.user_id = ?
		ORDER BY n.created_at DESC
		LIMIT ? OFFSET ?`

	type notifResult struct {
		PriceWatchNotification
		ProductName sql.NullString `db:"product_name"`
		Image       sql.NullString `db:"image"`
	}

	var results []notifResult
	if err := s.db.SelectContext(ctx, &results, query, userID, pageSize, offset); err != nil {
		return nil, err
	}

	items := make([]NotificationItem, len(results))
	for i, r := range results {
		items[i] = NotificationItem{
			NotificationID: r.ID,
			SkuID:          r.SkuID,
			ProductName:    r.ProductName.String,
			Image:          r.Image.String,
			OldPrice:       r.OldPrice,
			NewPrice:       r.NewPrice,
			DiscountAmount: r.DiscountAmount,
			DiscountRate:   r.DiscountRate,
			IsRead:         r.IsRead,
			CreatedAt:      r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &NotificationListResponse{
		Items:       items,
		UnreadCount: unread,
		Page:        page,
		PageSize:    pageSize,
		Total:       total,
	}, nil
}

// MarkNotificationRead marks a notification as read
func (s *Service) MarkNotificationRead(ctx context.Context, userID int64, notificationID int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE price_watch_notifications SET is_read = 1, read_at = NOW()
		WHERE user_id = ? AND id = ?`,
		userID, notificationID)
	return err
}

func float64Ptr(v float64) *float64 {
	return &v
}
