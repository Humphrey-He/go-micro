package notification

import (
    "context"
    "database/sql"
    "time"
)

type Store struct {
    db *sql.DB
}

func NewStore(db *sql.DB) *Store {
    return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, n *Notification) error {
    query := `INSERT INTO notifications (user_id, type, title, content, is_read, created_at) VALUES (?, ?, ?, ?, ?, ?)`
    result, err := s.db.ExecContext(ctx, query, n.UserID, n.Type, n.Title, n.Content, n.IsRead, time.Now())
    if err != nil {
        return err
    }
    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    n.ID = id
    return nil
}

func (s *Store) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Notification, error) {
    query := `SELECT id, user_id, type, title, content, is_read, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
    rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var notifications []*Notification
    for rows.Next() {
        n := &Notification{}
        err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Content, &n.IsRead, &n.CreatedAt)
        if err != nil {
            return nil, err
        }
        notifications = append(notifications, n)
    }
    return notifications, nil
}

func (s *Store) CountUnread(ctx context.Context, userID string) (int, error) {
    query := `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`
    var count int
    err := s.db.QueryRowContext(ctx, query, userID).Scan(&count)
    return count, err
}

func (s *Store) MarkRead(ctx context.Context, id int64) error {
    query := `UPDATE notifications SET is_read = 1 WHERE id = ?`
    _, err := s.db.ExecContext(ctx, query, id)
    return err
}

func (s *Store) MarkAllRead(ctx context.Context, userID string) error {
    query := `UPDATE notifications SET is_read = 1 WHERE user_id = ?`
    _, err := s.db.ExecContext(ctx, query, userID)
    return err
}

func (s *Store) GetConfig(ctx context.Context, userID, notifType string) (*NotificationConfig, error) {
    query := `SELECT id, user_id, type, email_enabled, push_enabled, threshold FROM notification_configs WHERE user_id = ? AND type = ?`
    c := &NotificationConfig{}
    err := s.db.QueryRowContext(ctx, query, userID, notifType).Scan(&c.ID, &c.UserID, &c.Type, &c.EmailEnabled, &c.PushEnabled, &c.Threshold)
    if err != nil {
        return nil, err
    }
    return c, nil
}

func (s *Store) UpsertConfig(ctx context.Context, c *NotificationConfig) error {
    query := `INSERT INTO notification_configs (user_id, type, email_enabled, push_enabled, threshold) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE email_enabled = ?, push_enabled = ?, threshold = ?`
    _, err := s.db.ExecContext(ctx, query, c.UserID, c.Type, c.EmailEnabled, c.PushEnabled, c.Threshold, c.EmailEnabled, c.PushEnabled, c.Threshold)
    return err
}