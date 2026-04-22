# Operations Dashboard Time Filter & Notification Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现运营看板时间周期筛选（快速选择 + 天/周/月维度）和通知模块（WebSocket实时推送 + 邮件通知）

**Architecture:**
- 模块一：前端时间筛选组件 + 后端 dashboardStats 接口扩展，支持 start_time/end_time/period 参数
- 模块二：通知服务（notification service）+ WebSocket 实时推送 + 邮件发送 + 定时任务cron，前端通知中心组件

**Tech Stack:** Go (gin, gorilla/websocket), React (Ant Design), MySQL, SMTP

---

## Phase 1: 看板时间周期筛选

### Task 1: 后端 - dashboardStats 接口扩展

**Files:**
- Modify: `internal/gateway/handler.go:515-529`
- Modify: `internal/gateway/service.go:131-184`

- [ ] **Step 1: 修改 handler.go 的 dashboardStats 函数，解析时间参数**

```go
// @Summary 运营看板统计
// @Tags Admin
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param start_time query int64 false "开始时间戳"
// @Param end_time query int64 false "结束时间戳"
// @Param period query string false "统计周期 day/week/month" default(day)
// @Success 200 {object} httpx.Response
// @Router /api/v1/admin/dashboard/stats [get]
func (h *Handler) dashboardStats(c *gin.Context) {
    startTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
    endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
    period := c.DefaultQuery("period", "day")

    resp, err := h.svc.GetDashboardStats(startTime, endTime, period)
    if err != nil {
        code, body := httpx.Fail(errx.CodeUpstreamUnavail, "order service unavailable")
        c.JSON(code, body)
        return
    }
    c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 2: 修改 service.go 的 GetDashboardStats 函数签名和实现**

在 `service.go` 中找到现有的 `GetDashboardStats` 函数（约在131-184行），修改为：

```go
type DashboardStats struct {
    TodayOrderCount     int64   `json:"today_order_count"`
    TodayOrderAmount    int64   `json:"today_order_amount"`
    PendingRefundCount  int64   `json:"pending_refund_count"`
    PaymentSuccessRate  float64 `json:"payment_success_rate"`
    LowStockSkuCount    int64   `json:"low_stock_sku_count"`
    Period              string  `json:"period"`
    StartTime           int64   `json:"start_time"`
    EndTime             int64   `json:"end_time"`
}

func (s *Service) GetDashboardStats(startTime, endTime int64, period string) (httpx.Response, error) {
    // 如果未提供时间参数，使用默认值（今日0点 ~ 现在）
    now := time.Now()
    if startTime == 0 {
        startTime = now.Truncate(24 * time.Hour).Unix()
    }
    if endTime == 0 {
        endTime = now.Unix()
    }
    if period == "" {
        period = "day"
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var totalOrders, successOrders int64
    var totalAmount int64

    _ = s.cbOrder.Execute(func() error {
        orders, err := s.order.List(ctx, order.ListOrdersRequest{
            Page:      1,
            PageSize:  10000, // 增加 PageSize 以获取更多数据
            StartTime: startTime,
            EndTime:   endTime,
        })
        if err != nil {
            return err
        }
        totalOrders = int64(orders.Total)
        for _, o := range orders.Orders {
            if o.Status == "SUCCESS" {
                successOrders++
                totalAmount += o.TotalAmount
            }
        }
        return nil
    })

    var paymentSuccessRate float64
    if totalOrders > 0 {
        paymentSuccessRate = float64(successOrders) / float64(totalOrders) * 100
    }

    stats := DashboardStats{
        TodayOrderCount:    totalOrders,
        TodayOrderAmount:   totalAmount,
        PendingRefundCount: 0,
        PaymentSuccessRate: paymentSuccessRate,
        LowStockSkuCount:   0,
        Period:             period,
        StartTime:          startTime,
        EndTime:            endTime,
    }

    return httpx.Response{Code: 0, Message: "OK", Data: stats}, nil
}
```

- [ ] **Step 3: 验证编译**

Run: `cd E:/awesomeProject/go-micro && go build ./...`
Expected: 编译成功，无错误

- [ ] **Step 4: 提交**

```bash
git add internal/gateway/handler.go internal/gateway/service.go
git commit -m "feat(dashboard): add time period and period dimension params to dashboardStats API"
```

---

### Task 2: 前端 - dashboardApi 和 hooks 更新

**Files:**
- Modify: `order-admin/src/features/dashboard/dashboardApi.ts`
- Modify: `order-admin/src/features/dashboard/hooks/useDashboardStats.ts`

- [ ] **Step 1: 更新 dashboardApi.ts，添加时间参数支持**

```typescript
import { get } from '@/api/request'
import type { DashboardStats } from './types/stats'

export interface DashboardStatsParams {
  start_time?: number
  end_time?: number
  period?: 'day' | 'week' | 'month'
}

export const getDashboardStats = (params?: DashboardStatsParams) =>
  get<DashboardStats>('/admin/dashboard/stats', params)
```

- [ ] **Step 2: 更新 useDashboardStats.ts，支持传入时间参数**

```typescript
import { useState, useCallback, useEffect } from 'react'
import { getDashboardStats } from '../dashboardApi'
import type { DashboardStats } from '../types/stats'
import type { DashboardStatsParams } from '../dashboardApi'

interface UseDashboardStatsResult {
  loading: boolean
  data: DashboardStats | null
  error: string | null
  refresh: () => Promise<void>
}

export const useDashboardStats = (params?: DashboardStatsParams): UseDashboardStatsResult => {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<DashboardStats | null>(null)
  const [error, setError] = useState<string | null>(null)

  const fetchStats = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await getDashboardStats(params)
      setData(res)
    } catch {
      setError('获取统计数据失败')
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchStats()
  }, [fetchStats])

  return {
    loading,
    data,
    error,
    refresh: fetchStats,
  }
}
```

- [ ] **Step 3: 更新 stats.ts 类型定义**

```typescript
export interface DashboardStats {
  today_order_count: number
  today_order_amount: number
  pending_refund_count: number
  payment_success_rate: number
  low_stock_sku_count: number
  period?: 'day' | 'week' | 'month'
  start_time?: number
  end_time?: number
}
```

- [ ] **Step 4: 提交**

```bash
git add order-admin/src/features/dashboard/dashboardApi.ts order-admin/src/features/dashboard/hooks/useDashboardStats.ts order-admin/src/features/dashboard/types/stats.ts
git commit -m "feat(dashboard): add time period params support to dashboard API client"
```

---

### Task 3: 前端 - Dashboard 页面时间筛选 UI

**Files:**
- Modify: `order-admin/src/pages/Dashboard/index.tsx`

- [ ] **Step 1: 添加时间选择相关的 import**

在现有的 import 后面添加：

```typescript
import { Segmented, Space, Button } from 'antd'
import { CalendarOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
```

- [ ] **Step 2: 添加时间选择状态和快捷选项**

在 DashboardPage 组件内添加状态：

```typescript
const [selectedPeriod, setSelectedPeriod] = useState<'day' | 'week' | 'month'>('day')

const timeRanges = [
  { label: '今日', value: 'today' },
  { label: '本周', value: 'week' },
  { label: '本月', value: 'month' },
  { label: '本年', value: 'year' },
]
const [selectedRange, setSelectedRange] = useState('today')
```

- [ ] **Step 3: 添加计算时间戳的函数**

```typescript
const getTimeRange = (range: string) => {
  const now = dayjs()
  switch (range) {
    case 'today':
      return {
        start_time: now.startOf('day').unix(),
        end_time: now.unix(),
      }
    case 'week':
      return {
        start_time: now.startOf('week').unix(),
        end_time: now.unix(),
      }
    case 'month':
      return {
        start_time: now.startOf('month').unix(),
        end_time: now.unix(),
      }
    case 'year':
      return {
        start_time: now.startOf('year').unix(),
        end_time: now.unix(),
      }
    default:
      return {
        start_time: now.startOf('day').unix(),
        end_time: now.unix(),
      }
  }
}
```

- [ ] **Step 4: 修改 useDashboardStats 调用，传入时间参数**

将原来的：
```typescript
const { loading, data, error } = useDashboardStats()
```

改为：
```typescript
const timeRange = getTimeRange(selectedRange)
const { loading, data, error } = useDashboardStats({
  ...timeRange,
  period: selectedPeriod,
})
```

- [ ] **Step 5: 修改页面顶部的数据周期显示**

将原来的静态显示：
```tsx
<div style={{...}}>
  数据统计周期：今日 00:00 - 现在
</div>
```

改为动态显示：
```tsx
<div style={{...}}>
  <Space size={8}>
    <CalendarOutlined />
    <span>
      数据统计周期：{dayjs.unix(timeRange.start_time).format('MM月DD日 HH:mm')} - {dayjs.unix(timeRange.end_time).format('MM月DD日 HH:mm')}
    </span>
  </Space>
</div>
```

- [ ] **Step 6: 在页面头部添加快捷按钮组和维度切换**

在页面头部 `{dayjs().format('YYYY年MM月DD日 dddd')} · 数据实时更新` 那行下面添加：

```tsx
<div style={{ marginTop: 12, display: 'flex', alignItems: 'center', gap: 16 }}>
  <Space size={4}>
    {timeRanges.map((range) => (
      <Button
        key={range.value}
        type={selectedRange === range.value ? 'primary' : 'default'}
        size="small"
        onClick={() => setSelectedRange(range.value)}
      >
        {range.label}
      </Button>
    ))}
  </Space>
  <Segmented
    value={selectedPeriod}
    onChange={(value) => setSelectedPeriod(value as 'day' | 'week' | 'month')}
    options={[
      { label: '按天', value: 'day' },
      { label: '按周', value: 'week' },
      { label: '按月', value: 'month' },
    ]}
  />
</div>
```

- [ ] **Step 7: 验证前端编译**

Run: `cd order-admin && npm run build 2>&1 | head -50`
Expected: 编译成功或仅警告

- [ ] **Step 8: 提交**

```bash
git add order-admin/src/pages/Dashboard/index.tsx
git commit -m "feat(dashboard): add time range selector and period dimension switcher UI"
```

---

## Phase 2: 通知模块后端

### Task 4: 数据库 - 创建通知相关表

**Files:**
- Create: `internal/notification/model.go`
- Create: `internal/notification/store.go`

- [ ] **Step 1: 创建 notification 目录和 model.go**

```go
package notification

import "time"

type Notification struct {
    ID        int64     `json:"id" db:"id"`
    UserID    string    `json:"user_id" db:"user_id"`
    Type      string    `json:"type" db:"type"`
    Title     string    `json:"title" db:"title"`
    Content   string    `json:"content" db:"content"`
    IsRead    bool      `json:"is_read" db:"is_read"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type NotificationConfig struct {
    ID            int64  `json:"id" db:"id"`
    UserID        string `json:"user_id" db:"user_id"`
    Type          string `json:"type" db:"type"`
    EmailEnabled  bool   `json:"email_enabled" db:"email_enabled"`
    PushEnabled   bool   `json:"push_enabled" db:"push_enabled"`
    Threshold     int    `json:"threshold" db:"threshold"`
}

// NotificationType 常量
const (
    TypeRefundPending  = "refund_pending"
    TypeLowStock        = "low_stock"
    TypePaymentFailed   = "payment_failed"
    TypeDailyReport     = "daily_report"
    TypeWeeklyReport    = "weekly_report"
)
```

- [ ] **Step 2: 创建 store.go，实现数据库操作**

```go
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
```

- [ ] **Step 3: 创建数据库初始化 SQL**

创建 `sql/notification.sql`：

```sql
-- 通知记录表
CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) NOT NULL,
    type VARCHAR(32) NOT NULL,
    title VARCHAR(128) NOT NULL,
    content TEXT,
    is_read TINYINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_read (user_id, is_read),
    INDEX idx_created (created_at)
);

-- 通知订阅配置表
CREATE TABLE IF NOT EXISTS notification_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) NOT NULL,
    type VARCHAR(32) NOT NULL,
    email_enabled TINYINT DEFAULT 1,
    push_enabled TINYINT DEFAULT 1,
    threshold INT DEFAULT 0,
    UNIQUE INDEX idx_user_type (user_id, type)
);
```

- [ ] **Step 4: 验证编译**

Run: `cd E:/awesomeProject/go-micro && go build ./internal/notification/...`
Expected: 编译成功

- [ ] **Step 5: 提交**

```bash
git add internal/notification/model.go internal/notification/store.go sql/notification.sql
git commit -m "feat(notification): add notification model and store with DB schema"
```

---

### Task 5: 通知服务 - Service 和 Handler

**Files:**
- Create: `internal/notification/service.go`
- Create: `internal/notification/handler.go`

- [ ] **Step 1: 创建 service.go**

```go
package notification

import (
    "context"
    "log"
    "net/smtp"
    "strings"
    "time"

    "go-micro/pkg/config"
)

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
    // 这里只是占位，实际通过全局 hub 广播
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
```

- [ ] **Step 2: 创建 handler.go**

```go
package notification

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "go-micro/pkg/errx"
    "go-micro/pkg/httpx"
    "go-micro/pkg/middleware"
)

type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
    api := r.Group("/api/v1")
    api.Use(middleware.JWTAuth())
    api.GET("/notifications", h.listNotifications)
    api.GET("/notifications/unread-count", h.unreadCount)
    api.PUT("/notifications/:id/read", h.markRead)
    api.PUT("/notifications/read-all", h.markAllRead)
    api.GET("/notification/configs", h.getConfigs)
    api.PUT("/notification/configs", h.updateConfigs)
}

func (h *Handler) listNotifications(c *gin.Context) {
    userID, _ := c.Get(middleware.CtxUserID)
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

    notifications, unreadCount, err := h.svc.ListNotifications(c.Request.Context(), userID.(string), page, pageSize)
    if err != nil {
        code, body := httpx.Fail(errx.CodeInternal, "failed to get notifications")
        c.JSON(code, body)
        return
    }

    c.JSON(http.StatusOK, httpx.OK(gin.H{
        "notifications": notifications,
        "unread_count":  unreadCount,
        "page":          page,
        "page_size":     pageSize,
    }))
}

func (h *Handler) unreadCount(c *gin.Context) {
    userID, _ := c.Get(middleware.CtxUserID)
    count, err := h.svc.GetUnreadCount(c.Request.Context(), userID.(string))
    if err != nil {
        code, body := httpx.Fail(errx.CodeInternal, "failed to get unread count")
        c.JSON(code, body)
        return
    }
    c.JSON(http.StatusOK, httpx.OK(gin.H{"count": count}))
}

func (h *Handler) markRead(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
        c.JSON(code, body)
        return
    }
    if err := h.svc.MarkRead(c.Request.Context(), id); err != nil {
        code, body := httpx.Fail(errx.CodeInternal, "failed to mark read")
        c.JSON(code, body)
        return
    }
    c.JSON(http.StatusOK, httpx.OK(gin.H{"success": true}))
}

func (h *Handler) markAllRead(c *gin.Context) {
    userID, _ := c.Get(middleware.CtxUserID)
    if err := h.svc.MarkAllRead(c.Request.Context(), userID.(string)); err != nil {
        code, body := httpx.Fail(errx.CodeInternal, "failed to mark all read")
        c.JSON(code, body)
        return
    }
    c.JSON(http.StatusOK, httpx.OK(gin.H{"success": true}))
}

func (h *Handler) getConfigs(c *gin.Context) {
    userID, _ := c.Get(middleware.CtxUserID)
    notifType := c.Query("type")

    cfg, err := h.svc.GetConfig(c.Request.Context(), userID.(string), notifType)
    if err != nil {
        cfg = &NotificationConfig{
            UserID:       userID.(string),
            Type:         notifType,
            EmailEnabled: true,
            PushEnabled:  true,
            Threshold:    0,
        }
    }
    c.JSON(http.StatusOK, httpx.OK(cfg))
}

func (h *Handler) updateConfigs(c *gin.Context) {
    var req NotificationConfig
    if err := c.ShouldBindJSON(&req); err != nil {
        code, body := httpx.Fail(errx.CodeInvalidRequest, errx.MsgInvalidRequest)
        c.JSON(code, body)
        return
    }
    userID, _ := c.Get(middleware.CtxUserID)
    req.UserID = userID.(string)

    if err := h.svc.UpdateConfig(c.Request.Context(), &req); err != nil {
        code, body := httpx.Fail(errx.CodeInternal, "failed to update config")
        c.JSON(code, body)
        return
    }
    c.JSON(http.StatusOK, httpx.OK(gin.H{"success": true}))
}
```

- [ ] **Step 3: 验证编译**

Run: `cd E:/awesomeProject/go-micro && go build ./internal/notification/...`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add internal/notification/service.go internal/notification/handler.go
git commit -m "feat(notification): add notification service and HTTP handler"
```

---

### Task 6: WebSocket 实时推送

**Files:**
- Create: `internal/notification/websocket.go`

- [ ] **Step 1: 创建 websocket.go**

```go
package notification

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境应限制 origin
    },
}

type Client struct {
    userID string
    conn   *websocket.Conn
    send   chan []byte
}

type Hub struct {
    clients    map[string][]*Client
    register   chan *Client
    unregister chan *Client
    mutex      sync.RWMutex
}

var hub = &Hub{
    clients:    make(map[string][]*Client),
    register:   make(chan *Client),
    unregister: make(chan *Client),
}

func init() {
    go hub.run()
}

func (h *Hub) run() {
    for {
        select {
        case client := <-h.register:
            h.mutex.Lock()
            h.clients[client.userID] = append(h.clients[client.userID], client)
            h.mutex.Unlock()
            log.Printf("WebSocket client registered for user: %s", client.userID)

        case client := <-h.unregister:
            h.mutex.Lock()
            for i, c := range h.clients[client.userID] {
                if c == client {
                    h.clients[client.userID] = append(h.clients[client.userID][:i], h.clients[client.userID][i+1:]...)
                    close(c.send)
                    break
                }
            }
            h.mutex.Unlock()
            log.Printf("WebSocket client unregistered for user: %s", client.userID)
        }
    }
}

// BroadcastNotification 广播通知给指定用户
func BroadcastNotification(userID string, n *Notification) {
    hub.mutex.RLock()
    clients := hub.clients[userID]
    hub.mutex.RUnlock()

    msg, _ := json.Marshal(map[string]interface{}{
        "type": "notification",
        "data": n,
    })

    for _, client := range clients {
        select {
        case client.send <- msg:
        default:
            close(client.send)
        }
    }
}

func HandleWebSocket(c *gin.Context) {
    userID, _ := c.Get("user_id")
    if userID == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }

    client := &Client{
        userID: userID.(string),
        conn:   conn,
        send:   make(chan []byte, 256),
    }

    hub.register <- client

    go client.writePump()
    go client.readPump()
}

func (c *Client) writePump() {
    defer func() {
        c.conn.Close()
    }()

    for {
        message, ok := <-c.send
        if !ok {
            c.conn.WriteMessage(websocket.CloseMessage, []byte{})
            return
        }
        if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
            return
        }
    }
}

func (c *Client) readPump() {
    defer func() {
        hub.unregister <- c
        c.conn.Close()
    }()

    for {
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
    }
}
```

- [ ] **Step 2: 更新 handler.go 注册 WebSocket 路由**

在 Register 函数中添加：

```go
func (h *Handler) Register(r *gin.Engine) {
    // ... existing routes ...
    r.GET("/ws/notifications", HandleWebSocket)
}
```

- [ ] **Step 3: 验证编译**

Run: `cd E:/awesomeProject/go-micro && go build ./internal/notification/...`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add internal/notification/websocket.go internal/notification/handler.go
git commit -m "feat(notification): add WebSocket real-time notification support"
```

---

### Task 7: 定时任务

**Files:**
- Create: `internal/notification/cron.go`

- [ ] **Step 1: 创建 cron.go**

```go
package notification

import (
    "log"
    "time"

    "github.com/robfig/cron/v3"
)

type CronService struct {
    notificationSvc *Service
    cron            *cron.Cron
}

func NewCronService(notificationSvc *Service) *CronService {
    return &CronService{
        notificationSvc: notificationSvc,
        cron:            cron.New(),
    }
}

func (s *CronService) Start() error {
    // 检查退款告警 - 每5分钟
    _, err := s.cron.AddFunc("*/5 * * * *", func() {
        s.checkRefundAlert()
    })
    if err != nil {
        return err
    }

    // 检查库存告警 - 每5分钟
    _, err = s.cron.AddFunc("*/5 * * * *", func() {
        s.checkLowStockAlert()
    })
    if err != nil {
        return err
    }

    // 每日报告 - 每天早上9点
    _, err = s.cron.AddFunc("0 9 * * *", func() {
        if err := s.notificationSvc.SendDailyReport(); err != nil {
            log.Printf("Failed to send daily report: %v", err)
        }
    })
    if err != nil {
        return err
    }

    // 每周报告 - 每周一早上9点
    _, err = s.cron.AddFunc("0 9 * * 1", func() {
        if err := s.notificationSvc.SendWeeklyReport(); err != nil {
            log.Printf("Failed to send weekly report: %v", err)
        }
    })
    if err != nil {
        return err
    }

    s.cron.Start()
    log.Println("Notification cron service started")
    return nil
}

func (s *CronService) Stop() {
    s.cron.Stop()
}

func (s *CronService) checkRefundAlert() {
    // TODO: 从数据库或服务获取当前待处理退款数量
    // 这里需要接入实际的退款服务
    pendingCount := 0 // 暂时设为0，后续接入实际服务

    if pendingCount > 10 {
        err := s.notificationSvc.CreateNotification(
            nil,
            "admin",
            TypeRefundPending,
            "退款告警",
            "待处理退款数量超过阈值，当前："+string(rune(pendingCount))+"件",
        )
        if err != nil {
            log.Printf("Failed to create refund alert: %v", err)
        }
    }
}

func (s *CronService) checkLowStockAlert() {
    // TODO: 从库存服务获取低库存SKU数量
    lowStockCount := 0

    if lowStockCount > 5 {
        err := s.notificationSvc.CreateNotification(
            nil,
            "admin",
            TypeLowStock,
            "库存告警",
            "低库存SKU数量超过阈值，当前："+string(rune(lowStockCount))+"个",
        )
        if err != nil {
            log.Printf("Failed to create low stock alert: %v", err)
        }
    }
}
```

- [ ] **Step 2: 检查是否有 robfig/cron 依赖**

如果没有，添加到 go.mod：

Run: `cd E:/awesomeProject/go-micro && go get github.com/robfig/cron/v3`

- [ ] **Step 3: 验证编译**

Run: `cd E:/awesomeProject/go-micro && go build ./internal/notification/...`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add internal/notification/cron.go
git commit -m "feat(notification): add cron jobs for periodic alerts and reports"
```

---

## Phase 3: 通知模块前端

### Task 8: 前端 - 通知 API 和 Hooks

**Files:**
- Create: `order-admin/src/features/notification/notificationApi.ts`
- Create: `order-admin/src/features/notification/hooks/useNotification.ts`
- Create: `order-admin/src/features/notification/types/notification.ts`

- [ ] **Step 1: 创建 notificationApi.ts**

```typescript
import { get, put } from '@/api/request'
import type { Notification, NotificationConfig } from './types/notification'

export interface NotificationListParams {
  page?: number
  page_size?: number
}

export const getNotifications = (params?: NotificationListParams) =>
  get<{
    notifications: Notification[]
    unread_count: number
    page: number
    page_size: number
  }>('/notifications', params)

export const getUnreadCount = () =>
  get<{ count: number }>('/notifications/unread-count')

export const markAsRead = (id: number) =>
  put<{ success: boolean }>(`/notifications/${id}/read`)

export const markAllAsRead = () =>
  put<{ success: boolean }>('/notifications/read-all')

export const getNotificationConfig = (type: string) =>
  get<NotificationConfig>('/notification/configs', { type })

export const updateNotificationConfig = (config: NotificationConfig) =>
  put<{ success: boolean }>('/notification/configs', config)
```

- [ ] **Step 2: 创建 types/notification.ts**

```typescript
export interface Notification {
  id: number
  user_id: string
  type: 'refund_pending' | 'low_stock' | 'payment_failed' | 'daily_report' | 'weekly_report'
  title: string
  content: string
  is_read: boolean
  created_at: string
}

export interface NotificationConfig {
  id?: number
  user_id: string
  type: string
  email_enabled: boolean
  push_enabled: boolean
  threshold: number
}
```

- [ ] **Step 3: 创建 hooks/useNotification.ts**

```typescript
import { useState, useCallback, useEffect } from 'react'
import {
  getNotifications,
  getUnreadCount,
  markAsRead,
  markAllAsRead,
} from '../notificationApi'
import type { Notification } from '../types/notification'

interface UseNotificationsResult {
  notifications: Notification[]
  unreadCount: number
  loading: boolean
  error: string | null
  page: number
  pageSize: number
  refresh: () => Promise<void>
  loadMore: () => Promise<void>
  markRead: (id: number) => Promise<void>
  markAllRead: () => Promise<void>
}

export const useNotifications = (): UseNotificationsResult => {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)

  const fetchNotifications = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await getNotifications({ page: 1, page_size: pageSize })
      setNotifications(res.notifications || [])
      setUnreadCount(res.unread_count || 0)
      setPage(1)
    } catch {
      setError('获取通知失败')
    } finally {
      setLoading(false)
    }
  }, [pageSize])

  const loadMore = useCallback(async () => {
    if (loading) return
    setLoading(true)
    try {
      const nextPage = page + 1
      const res = await getNotifications({ page: nextPage, page_size: pageSize })
      setNotifications((prev) => [...prev, ...(res.notifications || [])])
      setPage(nextPage)
    } catch {
      setError('加载更多失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, loading])

  const markRead = useCallback(async (id: number) => {
    try {
      await markAsRead(id)
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, is_read: true } : n))
      )
      setUnreadCount((prev) => Math.max(0, prev - 1))
    } catch {
      setError('标记已读失败')
    }
  }, [])

  const markAllRead = useCallback(async () => {
    try {
      await markAllAsRead()
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })))
      setUnreadCount(0)
    } catch {
      setError('标记全部已读失败')
    }
  }, [])

  useEffect(() => {
    fetchNotifications()
  }, [fetchNotifications])

  return {
    notifications,
    unreadCount,
    loading,
    error,
    page,
    pageSize,
    refresh: fetchNotifications,
    loadMore,
    markRead,
    markAllRead,
  }
}
```

- [ ] **Step 4: 创建 hooks/useUnreadCount.ts（用于 Bell 组件）**

```typescript
import { useState, useCallback, useEffect } from 'react'
import { getUnreadCount } from '../notificationApi'

export const useUnreadCount = () => {
  const [count, setCount] = useState(0)
  const [loading, setLoading] = useState(false)

  const fetch = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getUnreadCount()
      setCount(res.count || 0)
    } catch {
      // 静默失败
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetch()
    // 轮询每30秒刷新一次
    const interval = setInterval(fetch, 30000)
    return () => clearInterval(interval)
  }, [fetch])

  return { count, loading, refresh: fetch }
}
```

- [ ] **Step 5: 提交**

```bash
git add order-admin/src/features/notification/notificationApi.ts order-admin/src/features/notification/hooks/useNotification.ts order-admin/src/features/notification/hooks/useUnreadCount.ts order-admin/src/features/notification/types/notification.ts
git commit -m "feat(notification): add notification API client and hooks"
```

---

### Task 9: 前端 - NotificationBell 组件

**Files:**
- Create: `order-admin/src/features/notification/components/NotificationBell.tsx`

- [ ] **Step 1: 创建 NotificationBell.tsx**

```typescript
import React from 'react'
import { Badge, Popover, List, Button, Typography, Space, Empty } from 'antd'
import { BellOutlined, CheckOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useNotifications } from '../hooks/useNotification'
import { useUnreadCount } from '../hooks/useUnreadCount'

const { Text, Title } = Typography

export const NotificationBell: React.FC = () => {
  const { count } = useUnreadCount()
  const { notifications, markRead, markAllRead, refresh } = useNotifications()

  const popoverContent = (
    <div style={{ width: 360, maxHeight: 480, overflow: 'auto' }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 12,
          padding: '8px 0',
          borderBottom: '1px solid #f0f0f0',
        }}
      >
        <Title level={5} style={{ margin: 0 }}>
          通知中心
        </Title>
        {count > 0 && (
          <Button type="link" size="small" onClick={markAllRead}>
            全部已读
          </Button>
        )}
      </div>

      {notifications.length === 0 ? (
        <Empty description="暂无通知" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      ) : (
        <List
          dataSource={notifications.slice(0, 10)}
          renderItem={(item) => (
            <List.Item
              style={{
                padding: '10px 0',
                opacity: item.is_read ? 0.6 : 1,
                cursor: 'pointer',
              }}
              onClick={() => !item.is_read && markRead(item.id)}
            >
              <List.Item.Meta
                title={
                  <Space>
                    {!item.is_read && (
                      <Badge status="processing" color="#1677ff" />
                    )}
                    <Text strong={!item.is_read}>{item.title}</Text>
                  </Space>
                }
                description={
                  <div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {item.content}
                    </Text>
                    <div style={{ marginTop: 4 }}>
                      <Text type="secondary" style={{ fontSize: 11 }}>
                        {dayjs(item.created_at).format('MM-DD HH:mm')}
                      </Text>
                    </div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      )}

      <div style={{ textAlign: 'center', marginTop: 8 }}>
        <Button type="link" size="small" onClick={refresh}>
          刷新
        </Button>
      </div>
    </div>
  )

  return (
    <Popover
      content={popoverContent}
      trigger="click"
      placement="bottomRight"
      arrow={{ pointAtCenter: true }}
    >
      <Badge count={count} size="small" offset={[-2, 2]}>
        <Button
          type="text"
          icon={<BellOutlined style={{ fontSize: 20 }} />}
          style={{ color: '#fff' }}
        />
      </Badge>
    </Popover>
  )
}
```

- [ ] **Step 2: 提交**

```bash
git add order-admin/src/features/notification/components/NotificationBell.tsx
git commit -m "feat(notification): add NotificationBell component with popover list"
```

---

### Task 10: 前端 - Header 集成 NotificationBell

**Files:**
- Modify: `order-admin/src/layouts/BasicLayout/index.tsx`

- [ ] **Step 1: 添加 import**

```typescript
import { NotificationBell } from '@/features/notification/components/NotificationBell'
```

- [ ] **Step 2: 在 Header 区域添加 NotificationBell**

找到 Header 组件中合适的位置（通常是右侧区域），添加：

```tsx
<NotificationBell />
```

- [ ] **Step 3: 验证编译**

Run: `cd order-admin && npm run build 2>&1 | head -50`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add order-admin/src/layouts/BasicLayout/index.tsx
git commit -m "feat(notification): integrate NotificationBell into BasicLayout header"
```

---

### Task 11: 前端 - 通知列表页面

**Files:**
- Create: `order-admin/src/pages/Notifications/index.tsx`

- [ ] **Step 1: 创建 Notifications/index.tsx**

```typescript
import React, { useState } from 'react'
import {
  Card,
  List,
  Tabs,
  Button,
  Typography,
  Space,
  Badge,
  Empty,
  Spin,
} from 'antd'
import { CheckOutlined, BellOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useNotifications } from '@/features/notification/hooks/useNotification'

const { Title, Text } = Typography

const NOTIFICATION_TYPE_MAP: Record<string, string> = {
  refund_pending: '退款告警',
  low_stock: '库存告警',
  payment_failed: '支付失败',
  daily_report: '每日报告',
  weekly_report: '每周报告',
}

const NotificationItem: React.FC<{
  item: any
  onMarkRead: (id: number) => void
}> = ({ item, onMarkRead }) => (
  <List.Item
    style={{
      padding: '16px 20px',
      opacity: item.is_read ? 0.7 : 1,
      background: item.is_read ? '#fafafa' : '#fff',
      transition: 'all 0.3s',
    }}
    actions={
      !item.is_read
        ? [
            <Button
              key="read"
              type="link"
              size="small"
              icon={<CheckOutlined />}
              onClick={() => onMarkRead(item.id)}
            >
              标记已读
            </Button>,
          ]
        : []
    }
  >
    <List.Item.Meta
      avatar={
        <Badge
          status={item.is_read ? 'default' : 'processing'}
          text={
            <Text strong={!item.is_read} style={{ fontSize: 15 }}>
              {item.title}
            </Text>
          }
        />
      }
      title={
        <Space>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {NOTIFICATION_TYPE_MAP[item.type] || item.type}
          </Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {dayjs(item.created_at).format('YYYY-MM-DD HH:mm')}
          </Text>
        </Space>
      }
      description={
        <div style={{ marginTop: 8 }}>
          <Text style={{ fontSize: 14 }}>{item.content}</Text>
        </div>
      }
    />
  </List.Item>
)

export const NotificationsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState('all')
  const {
    notifications,
    unreadCount,
    loading,
    markRead,
    markAllRead,
    refresh,
  } = useNotifications()

  const filteredNotifications =
    activeTab === 'all'
      ? notifications
      : activeTab === 'unread'
      ? notifications.filter((n) => !n.is_read)
      : notifications.filter((n) => n.type === activeTab)

  const tabItems = [
    {
      key: 'all',
      label: (
        <Space>
          全部
          <Badge count={notifications.length} size="small" />
        </Space>
      ),
    },
    {
      key: 'unread',
      label: (
        <Space>
          未读
          <Badge count={unreadCount} size="small" />
        </Space>
      ),
    },
    {
      key: 'refund_pending',
      label: '退款告警',
    },
    {
      key: 'low_stock',
      label: '库存告警',
    },
    {
      key: 'daily_report',
      label: '报告',
    },
  ]

  return (
    <div style={{ padding: '24px 28px' }}>
      <div
        style={{
          marginBottom: 24,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div>
          <h1
            style={{
              fontSize: 22,
              fontWeight: 700,
              color: '#1f2937',
              marginBottom: 4,
            }}
          >
            通知中心
          </h1>
          <Text type="secondary">
            共 {notifications.length} 条通知，{unreadCount} 条未读
          </Text>
        </div>
        {unreadCount > 0 && (
          <Button icon={<CheckOutlined />} onClick={markAllRead}>
            全部已读
          </Button>
        )}
      </div>

      <Card
        style={{ borderRadius: 12, border: '1px solid #e5e7eb' }}
        bodyStyle={{ padding: 0 }}
      >
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          style={{ padding: '0 20px' }}
        />

        <Spin spinning={loading}>
          {filteredNotifications.length === 0 ? (
            <div style={{ padding: 60, textAlign: 'center' }}>
              <Empty
                description={activeTab === 'unread' ? '暂无未读通知' : '暂无通知'}
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            </div>
          ) : (
            <List
              dataSource={filteredNotifications}
              renderItem={(item) => (
                <NotificationItem item={item} onMarkRead={markRead} />
              )}
            />
          )}
        </Spin>
      </Card>
    </div>
  )
}

export default NotificationsPage
```

- [ ] **Step 2: 提交**

```bash
git add order-admin/src/pages/Notifications/index.tsx
git commit -m "feat(notification): add notifications list page"
```

---

### Task 12: 路由注册

**Files:**
- Modify: `order-admin/src/routes/index.tsx`

- [ ] **Step 1: 添加路由配置**

添加通知页面路由：

```tsx
{
  path: '/notifications',
  component: () => import('@/pages/Notifications'),
}
```

- [ ] **Step 2: 提交**

```bash
git add order-admin/src/routes/index.tsx
git commit -m "feat(notification): add notifications route"
```

---

## 验收检查清单

### 模块一（看板时间筛选）
- [ ] 后端 dashboardStats 接口支持 start_time, end_time, period 参数
- [ ] 前端 dashboardApi 支持传入时间参数
- [ ] useDashboardStats 支持传入时间参数
- [ ] Dashboard 页面显示时间范围选择快捷按钮
- [ ] Dashboard 页面支持天/周/月维度切换
- [ ] 选择不同时间范围后数据正确更新

### 模块二（通知模块）
- [ ] 数据库表 notifications 和 notification_configs 创建
- [ ] notification store 支持 CRUD 操作
- [ ] notification service 支持创建通知和查询
- [ ] WebSocket 支持用户订阅和实时推送
- [ ] cron 定时任务正确注册
- [ ] 前端 notificationApi 和 hooks 工作正常
- [ ] NotificationBell 组件在 Header 显示
- [ ] 点击铃铛可查看通知列表
- [ ] 通知列表页面 /notifications 正常显示
- [ ] 支持标记已读和全部已读
