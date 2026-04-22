package notification

import (
    "context"
    "log"
    "net/smtp"
    "strings"
    "time"

    "go-micro/pkg/config"
)

// BroadcastNotification is called to push notifications to connected WebSocket clients.
// Implementation is in websocket.go (see task #6 for WebSocket real-time push).
var BroadcastNotification func(userID string, n *Notification)

type Service struct {
    store *Store
    email *EmailService
}

func NewService(store *Store, email *EmailService) *Service {
    return &Service{store: store, email: email}
}

type EmailService struct {
    host     string
    port     string
    username string
    password string
    from     string
}

func NewEmailService() *EmailService {
    return &EmailService{
        host:     config.GetEnv("SMTP_HOST", "smtp.gmail.com"),
        port:     config.GetEnv("SMTP_PORT", "587"),
        username: config.GetEnv("SMTP_USERNAME", ""),
        password: config.GetEnv("SMTP_PASSWORD", ""),
        from:     config.GetEnv("SMTP_FROM", "noreply@example.com"),
    }
}

func (e *EmailService) Send(to, subject, body string) error {
    if e.username == "" || e.password == "" {
        log.Printf("Email not configured, skipping send to %s", to)
        return nil
    }

    msg := buildMessage(e.from, to, subject, body)
    auth := smtp.PlainAuth("", e.username, e.password, e.host)
    addr := e.host + ":" + e.port
    return smtp.SendMail(addr, auth, e.from, []string{to}, msg)
}

func buildMessage(from, to, subject, body string) []byte {
    headers := make(map[string]string)
    headers["From"] = from
    headers["To"] = to
    headers["Subject"] = subject
    headers["MIME-Version"] = "1.0"
    headers["Content-Type"] = "text/html; charset=\"utf-8\""

    var msg strings.Builder
    for k, v := range headers {
        msg.WriteString(k + ": " + v + "\r\n")
    }
    msg.WriteString("\r\n")
    msg.WriteString(body)
    return []byte(msg.String())
}

func (s *Service) CreateNotification(ctx context.Context, userID, notifType, title, content string) error {
    n := &Notification{
        UserID:  userID,
        Type:    notifType,
        Title:   title,
        Content: content,
        IsRead:  false,
    }
    if err := s.store.Create(ctx, n); err != nil {
        return err
    }

    // 异步发送通知（WebSocket + 邮件）
    go s.sendNotification(n)
    return nil
}

func (s *Service) sendNotification(n *Notification) {
    // 获取用户配置，决定发送方式
    ctx := context.Background()
    cfg, err := s.store.GetConfig(ctx, n.UserID, n.Type)
    if err != nil {
        // 配置不存在，使用默认配置（都开启）
        cfg = &NotificationConfig{PushEnabled: true, EmailEnabled: true}
    }

    if cfg.PushEnabled {
        s.broadcastToUser(n.UserID, n)
    }

    if cfg.EmailEnabled {
        s.sendEmailNotification(n)
    }
}

func (s *Service) broadcastToUser(userID string, n *Notification) {
    // WebSocket 广播逻辑在 websocket.go 中实现
    BroadcastNotification(userID, n)
}

func (s *Service) sendEmailNotification(n *Notification) {
    if s.email == nil {
        return
    }
    subject := "【通知】" + n.Title
    err := s.email.Send("admin@example.com", subject, n.Content)
    if err != nil {
        log.Printf("Failed to send email notification: %v", err)
    }
}

func (s *Service) ListNotifications(ctx context.Context, userID string, page, pageSize int) ([]*Notification, int, error) {
    offset := (page - 1) * pageSize
    notifications, err := s.store.ListByUser(ctx, userID, pageSize, offset)
    if err != nil {
        return nil, 0, err
    }
    unreadCount, err := s.store.CountUnread(ctx, userID)
    return notifications, unreadCount, err
}

func (s *Service) MarkRead(ctx context.Context, id int64) error {
    return s.store.MarkRead(ctx, id)
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
    return s.store.MarkAllRead(ctx, userID)
}

func (s *Service) GetUnreadCount(ctx context.Context, userID string) (int, error) {
    return s.store.CountUnread(ctx, userID)
}

func (s *Service) UpdateConfig(ctx context.Context, cfg *NotificationConfig) error {
    return s.store.UpsertConfig(ctx, cfg)
}

func (s *Service) GetConfig(ctx context.Context, userID, notifType string) (*NotificationConfig, error) {
    return s.store.GetConfig(ctx, userID, notifType)
}

// SendDailyReport 发送每日报告
func (s *Service) SendDailyReport() error {
    ctx := context.Background()
    now := time.Now()
    content := generateReportContent("每日", now)
    return s.CreateNotification(ctx, "admin", TypeDailyReport, "每日运营报告", content)
}

// SendWeeklyReport 发送每周报告
func (s *Service) SendWeeklyReport() error {
    ctx := context.Background()
    now := time.Now()
    content := generateReportContent("每周", now)
    return s.CreateNotification(ctx, "admin", TypeWeeklyReport, "每周运营报告", content)
}

func generateReportContent(period string, t time.Time) string {
    return "<h2>运营" + period + "报告</h2>" +
        "<p>报告时间：" + t.Format("2006年01月02日 15:04") + "</p>" +
        "<p>统计周期：详见运营看板</p>"
}