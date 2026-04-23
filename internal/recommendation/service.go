package recommendation

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	db    *sqlx.DB
	redis *redis.Client
}

func NewService(db *sqlx.DB, redis *redis.Client) *Service {
	return &Service{
		db:    db,
		redis: redis,
	}
}

// ReportBehavior - Report user behavior
func (s *Service) ReportBehavior(ctx context.Context, req *BehaviorReportRequest, userID int64) error {
	// Calculate 5-minute time bucket
	timeBucket := int(time.Now().Unix() / 300)

	_, err := s.db.Exec(`
		INSERT INTO user_behavior_logs (user_id, anonymous_id, sku_id, behavior_type, source, stay_duration, time_bucket, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE id=id
	`, userID, req.AnonymousID, req.SkuID, req.BehaviorType, req.Source, req.StayDuration, timeBucket)

	return err
}