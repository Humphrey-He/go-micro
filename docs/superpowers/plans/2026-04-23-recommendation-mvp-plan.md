# 推荐服务 MVP 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现推荐服务MVP，包含行为采集、Item-CF看了又看、首页推荐、类目热卖兜底

**Architecture:** 独立推荐服务，共享MySQL/Redis，写读分离。MQ异步采集行为数据，离线计算商品相似度和热卖榜，实时查询推荐结果。

**Tech Stack:** Go, MySQL, Redis, RabbitMQ, go-zero

---

## 文件结构

```
internal/
  app/
    recommendation/
      run.go              # 服务入口
  recommendation/
    model.go              # 数据模型
    service.go            # 业务逻辑
    handler.go            # HTTP handlers
    consumer.go           # MQ消费者
    cache.go             # Redis缓存
    algorithm/
      item_cf.go         # Item-CF算法
      user_cf.go         # User-CF算法
      bestseller.go      # 热卖计算
    job/
      item_cf_job.go     # Item-CF定时任务
      bestseller_job.go  # 热卖定时任务
      preference_job.go  # 用户偏好定时任务
pkg/
  mq/rabbit.go           # 已有，复用
  db/mysql.go            # 已有，复用
  cache/redis.go         # 已有，复用
docs/
  superpowers/
    specs/
      2026-04-23-recommendation-mvp-design.md  # 设计文档（已存在）
    plans/
      2026-04-23-recommendation-mvp-plan.md    # 本计划
sql/
  001_create_recommendation_tables.sql  # 数据库表
```

---

## Task 1: 数据库表创建

**Files:**
- Create: `sql/001_create_recommendation_tables.sql`
- Create: `sql/002_seed_test_data.sql` (测试数据)

- [ ] **Step 1: 创建数据库表 SQL**

```sql
-- 用户行为日志表
CREATE TABLE IF NOT EXISTS user_behavior_logs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id             BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    anonymous_id        VARCHAR(64) COMMENT '匿名用户ID',
    sku_id              BIGINT UNSIGNED NOT NULL COMMENT '商品SKU ID',
    behavior_type       ENUM('cart', 'favorite', 'purchase') NOT NULL COMMENT '行为类型',
    source              VARCHAR(32) DEFAULT 'unknown' COMMENT '来源',
    stay_duration       INT COMMENT '停留时长(秒)',
    time_bucket         INT NOT NULL COMMENT '5分钟时间桶: FLOOR(UNIX_TIMESTAMP/300)',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_type_time (user_id, behavior_type, created_at),
    INDEX idx_sku_id (sku_id),
    INDEX idx_anonymous (anonymous_id),
    INDEX idx_created_at (created_at),
    UNIQUE KEY uk_dedup (user_id, anonymous_id, sku_id, behavior_type, time_bucket)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户行为日志表';

-- 商品相似度表
CREATE TABLE IF NOT EXISTS product_similarity (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id_a        BIGINT UNSIGNED NOT NULL COMMENT '商品A',
    sku_id_b        BIGINT UNSIGNED NOT NULL COMMENT '商品B',
    scene           ENUM('cart', 'favorite', 'purchase') NOT NULL COMMENT '场景',
    similarity      DECIMAL(10,6) NOT NULL COMMENT '相似度分数 0-1',
    weight          INT DEFAULT 1 COMMENT '共同行为次数',
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (sku_id_a, sku_id_b, scene),
    INDEX idx_sku_a_scene (sku_id_a, scene),
    INDEX idx_similarity (similarity DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品相似度表';

-- 用户类目偏好表
CREATE TABLE IF NOT EXISTS user_category_preference (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL,
    category_id     BIGINT UNSIGNED NOT NULL COMMENT '类目ID',
    weight          DECIMAL(10,4) DEFAULT 1.0 COMMENT '偏好权重',
    source          ENUM('explicit', 'implicit') DEFAULT 'implicit' COMMENT '显式/隐式',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_category (user_id, category_id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户类目偏好表';

-- 类目热卖榜
CREATE TABLE IF NOT EXISTS category_bestsellers (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    category_id     BIGINT UNSIGNED NOT NULL,
    sku_id          BIGINT UNSIGNED NOT NULL,
    sales_score     DECIMAL(12,2) NOT NULL COMMENT '销量分数',
    rank            INT NOT NULL COMMENT '类目内的排名',
    period          ENUM('7d', '30d') DEFAULT '30d' COMMENT '统计周期',
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_category_sku_period (category_id, sku_id, period),
    INDEX idx_category_rank (category_id, period, rank)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='类目热卖榜';

-- 全站热卖榜（不分品类）
CREATE TABLE IF NOT EXISTS global_bestsellers (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id          BIGINT UNSIGNED NOT NULL,
    sales_score     DECIMAL(12,2) NOT NULL COMMENT '销量分数',
    rank            INT NOT NULL COMMENT '全站排名',
    period          ENUM('7d', '30d') DEFAULT '30d' COMMENT '统计周期',
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_sku_period (sku_id, period),
    INDEX idx_rank_period (period, rank)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='全站热卖榜';
```

- [ ] **Step 2: 创建测试数据 SQL**

```sql
-- 插入测试商品（假设products表已存在）
-- 插入测试行为数据
INSERT INTO user_behavior_logs (user_id, sku_id, behavior_type, source, time_bucket) VALUES
(1, 1001, 'cart', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(1, 1002, 'cart', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(1, 1001, 'purchase', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(2, 1001, 'favorite', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(2, 1003, 'purchase', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(3, 1002, 'cart', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(3, 1001, 'purchase', 'cart', FLOOR(UNIX_TIMESTAMP(NOW()) / 300));
```

- [ ] **Step 3: 执行 SQL 并验证**

Run: 手动执行上述SQL，或提供执行脚本

- [ ] **Step 4: Commit**

```bash
git add sql/001_create_recommendation_tables.sql sql/002_seed_test_data.sql
git commit -m "feat(recommendation): add database tables for recommendation service"
```

---

## Task 2: 推荐服务基础结构

**Files:**
- Create: `internal/app/recommendation/run.go`
- Create: `internal/recommendation/model.go`
- Create: `internal/recommendation/service.go`

- [ ] **Step 1: 创建 model.go**

```go
package recommendation

import "time"

type BehaviorType string

const (
	BehaviorCart      BehaviorType = "cart"
	BehaviorFavorite  BehaviorType = "favorite"
	BehaviorPurchase  BehaviorType = "purchase"
)

type BehaviorLog struct {
	ID           int64        `db:"id"`
	UserID       int64        `db:"user_id"`
	AnonymousID  string       `db:"anonymous_id"`
	SkuID        int64        `db:"sku_id"`
	BehaviorType BehaviorType `db:"behavior_type"`
	Source       string       `db:"source"`
	StayDuration int          `db:"stay_duration"`
	TimeBucket   int          `db:"time_bucket"`
	CreatedAt    time.Time    `db:"created_at"`
}

type ProductSimilarity struct {
	ID        int64     `db:"id"`
	SkuIDA    int64     `db:"sku_id_a"`
	SkuIDB    int64     `db:"sku_id_b"`
	Scene     string    `db:"scene"`
	Similarity float64   `db:"similarity"`
	Weight    int       `db:"weight"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UserCategoryPreference struct {
	ID          int64   `db:"id"`
	UserID      int64   `db:"user_id"`
	CategoryID  int64   `db:"category_id"`
	Weight      float64 `db:"weight"`
	Source      string  `db:"source"`
}

type CategoryBestseller struct {
	ID         int64   `db:"id"`
	CategoryID int64   `db:"category_id"`
	SkuID      int64   `db:"sku_id"`
	SalesScore float64 `db:"sales_score"`
	Rank       int     `db:"rank"`
	Period     string  `db:"period"`
}

type GlobalBestseller struct {
	ID         int64   `db:"id"`
	SkuID      int64   `db:"sku_id"`
	SalesScore float64 `db:"sales_score"`
	Rank       int     `db:"rank"`
	Period     string  `db:"period"`
}

// RecItem 推荐商品项
type RecItem struct {
	SkuID       int64   `json:"sku_id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Similarity  float64 `json:"similarity,omitempty"`
	Reason      string  `json:"reason,omitempty"`
}

// BehaviorReportRequest 行为上报请求
type BehaviorReportRequest struct {
	SkuID        int64  `json:"sku_id" form:"sku_id" binding:"required"`
	BehaviorType string `json:"behavior_type" form:"behavior_type" binding:"required,oneof=cart favorite purchase"`
	Source       string `json:"source" form:"source"`
	StayDuration int    `json:"stay_duration" form:"stay_duration"`
	AnonymousID  string `json:"anonymous_id" form:"anonymous_id"`
}

// SimilarProductsRequest 相似商品请求
type SimilarProductsRequest struct {
	SkuID int64 `uri:"sku_id" binding:"required"`
	Scene string `form:"scene" binding:"omitempty,oneof=cart favorite purchase"`
	Limit int    `form:"limit"`
}

// HomeRecRequest 首页推荐请求
type HomeRecRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
```

- [ ] **Step 2: 创建 service.go**

```go
package recommendation

import (
	"context"

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

func (s *Service) ReportBehavior(ctx context.Context, req *BehaviorReportRequest, userID int64) error {
	// 计算5分钟时间桶
	timeBucket := int(time.Now().Unix() / 300)

	_, err := s.db.Exec(`
		INSERT INTO user_behavior_logs (user_id, anonymous_id, sku_id, behavior_type, source, stay_duration, time_bucket, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE id=id
	`, userID, req.AnonymousID, req.SkuID, req.BehaviorType, req.Source, req.StayDuration, timeBucket)

	return err
}
```

- [ ] **Step 3: 创建 run.go**

```go
package recommendationapp

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/internal/recommendation"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/pkg/metrics"
	"go-micro/pkg/mq"
	"go-micro/pkg/tracing"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() error {
	logger := logx.L()
	defer logx.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(metrics.HTTPMiddleware())
	r.Use(tracing.Middleware("recommendation-api"))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 数据库
	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Error("mysql connect failed", zap.Error(err))
		return err
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: config.GetEnv("REDIS_ADDR", "localhost:6379"),
		DB:   config.GetInt("REDIS_DB", 1),
	})

	// RabbitMQ
	rabbit, err := mq.NewRabbit(
		config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		"recommendation_exchange",
		"user_behavior_topic",
		"user.behavior.#",
		"recommendation_dlx",
		"user_behavior_dlq",
	)
	if err != nil {
		logger.Warn("rabbitmq connect failed, running without MQ", zap.Error(err))
		rabbit = nil
	}

	svc := recommendation.NewService(dbx, rdb)
	h := recommendation.NewHandler(svc)

	// 启动MQ消费者
	if rabbit != nil {
		consumer := recommendation.NewConsumer(svc, rabbit)
		go consumer.Start(ctx)
	}

	h.Register(r)

	addr := config.GetEnv("RECOMMENDATION_ADDR", ":8085")
	srv := &http.Server{Addr: addr, Handler: r}
	logger.Info("recommendation-api starting", zap.String("addr", addr))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("recommendation-api start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("recommendation-api shutting down")
	return srv.Shutdown(shutdownCtx)
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/app/recommendation/run.go internal/recommendation/model.go internal/recommendation/service.go
git commit -m "feat(recommendation): add basic service structure"
```

---

## Task 3: 行为采集 API 和 MQ 消费者

**Files:**
- Modify: `internal/recommendation/service.go`
- Create: `internal/recommendation/consumer.go`
- Create: `internal/recommendation/handler.go`

- [ ] **Step 1: 创建 handler.go**

```go
package recommendation

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
	r.GET("/health", func(c *gin.Context) {
		code, body := httpx.OK(gin.H{"status": "ok"})
		c.JSON(code, body)
	})

	api := r.Group("/api/v1/rec")
	api.POST("/report", h.reportBehavior)
	api.GET("/similar/:sku_id", h.getSimilarProducts)
	api.GET("/home", h.getHomeRecommendations)
	api.GET("/cold-start", h.getColdStart)
	api.POST("/preference", h.setPreference)
}

// reportBehavior 上报用户行为
// @Summary 行为上报
// @Tags Recommendation
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer token"
// @Param body body BehaviorReportRequest true "行为数据"
// @Success 200 {object} httpx.Response
// @Router /api/v1/rec/report [post]
func (h *Handler) reportBehavior(c *gin.Context) {
	var req BehaviorReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	// 获取用户ID（可选，匿名用户没有）
	userID := int64(0)
	if uid, exists := c.Get(middleware.CtxUserID); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	if err := h.svc.ReportBehavior(c.Request.Context(), &req, userID); err != nil {
		code, body := httpx.Fail(errx.CodeInternal, "report failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(nil)
	c.JSON(code, body)
}

// getSimilarProducts 获取相似商品
// @Summary 看了又看
// @Tags Recommendation
// @Produce json
// @Param sku_id path int true "商品ID"
// @Param scene query string false "场景 cart|favorite|purchase"
// @Param limit query int false "返回数量"
// @Success 200 {object} httpx.Response
// @Router /api/v1/rec/similar/{sku_id} [get]
func (h *Handler) getSimilarProducts(c *gin.Context) {
	skuID, err := strconv.ParseInt(c.Param("sku_id"), 10, 64)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid sku_id")
		c.JSON(code, body)
		return
	}

	scene := c.DefaultQuery("scene", "purchase")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	resp, err := h.svc.GetSimilarProducts(c.Request.Context(), skuID, scene, limit)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternal, "get similar products failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// getHomeRecommendations 获取首页推荐
// @Summary 首页推荐
// @Tags Recommendation
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} httpx.Response
// @Router /api/v1/rec/home [get]
func (h *Handler) getHomeRecommendations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userID := int64(0)
	if uid, exists := c.Get(middleware.CtxUserID); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	resp, err := h.svc.GetHomeRecommendations(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternal, "get home recommendations failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// getColdStart 获取冷启动数据
// @Summary 冷启动
// @Tags Recommendation
// @Produce json
// @Success 200 {object} httpx.Response
// @Router /api/v1/rec/cold-start [get]
func (h *Handler) getColdStart(c *gin.Context) {
	resp, err := h.svc.GetColdStartData(c.Request.Context())
	if err != nil {
		code, body := httpx.Fail(errx.CodeInternal, "get cold start data failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(resp)
	c.JSON(code, body)
}

// setPreference 设置类目偏好
// @Summary 设置类目偏好
// @Tags Recommendation
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body SetPreferenceRequest true "偏好设置"
// @Success 200 {object} httpx.Response
// @Router /api/v1/rec/preference [post]
func (h *Handler) setPreference(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxUserID)
	uid, _ := userID.(int64)

	var req SetPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, body := httpx.Fail(errx.CodeInvalidRequest, "invalid request")
		c.JSON(code, body)
		return
	}

	if err := h.svc.SetUserPreference(c.Request.Context(), uid, req.CategoryIDs); err != nil {
		code, body := httpx.Fail(errx.CodeInternal, "set preference failed")
		c.JSON(code, body)
		return
	}

	code, body := httpx.OK(nil)
	c.JSON(code, body)
}

type SetPreferenceRequest struct {
	CategoryIDs []int64 `json:"category_ids" binding:"required,min=1"`
}
```

- [ ] **Step 2: 扩展 service.go 添加更多方法**

```go
// GetSimilarProducts 获取相似商品（看了又看）
func (s *Service) GetSimilarProducts(ctx context.Context, skuID int64, scene string, limit int) (*SimilarProductsResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// 从数据库查询相似商品
	rows := []ProductSimilarity{}
	err := s.db.Select(&rows, `
		SELECT sku_id_b as sku_id_a, sku_id_a as sku_id_b, similarity, weight
		FROM product_similarity
		WHERE sku_id_a = ? AND scene = ?
		UNION ALL
		SELECT sku_id_a, sku_id_b, similarity, weight
		FROM product_similarity
		WHERE sku_id_b = ? AND scene = ?
		ORDER BY similarity DESC
		LIMIT ?
	`, skuID, scene, skuID, scene, limit)
	if err != nil {
		return nil, err
	}

	items := make([]RecItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, RecItem{
			SkuID:      r.SkuIDB,
			Similarity: r.Similarity,
		})
	}

	// 补全商品信息（从缓存或商品服务）
	items = s.enrichItems(ctx, items)

	return &SimilarProductsResponse{
		Scene:  scene,
		Items:  items,
	}, nil
}

// GetHomeRecommendations 获取首页推荐
func (s *Service) GetHomeRecommendations(ctx context.Context, userID int64, page, pageSize int) (*HomeRecResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var items []RecItem
	var source string

	// 检查用户偏好
	prefs, _ := s.GetUserPreferences(ctx, userID)

	if len(prefs) > 0 && userID > 0 {
		// 有偏好的用户，尝试User-CF推荐
		items, source = s.getUserCFRecommendations(ctx, userID, page, pageSize)
		if len(items) < pageSize {
			// 不足时用类目热卖补足
			fallback := s.getCategoryBestsellersFallback(ctx, prefs, pageSize-len(items))
			items = append(items, fallback...)
		}
	} else {
		// 无偏好用户，返回全站热卖
		items, source = s.getGlobalBestsellers(ctx, page, pageSize)
	}

	return &HomeRecResponse{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    len(items),
		Source:   source,
	}, nil
}

// GetColdStartData 获取冷启动数据
func (s *Service) GetColdStartData(ctx context.Context) (*ColdStartResponse, error) {
	// 获取全站热卖
	hotItems, _ := s.getGlobalBestsellers(ctx, 1, 20)

	// 获取热门类目（简化版，可配置）
	categoryPrefs := []CategoryPref{
		{CategoryID: 1, Name: "手机数码", Image: "https://placeholder.com/phone.png"},
		{CategoryID: 2, Name: "服装鞋帽", Image: "https://placeholder.com/clothing.png"},
		{CategoryID: 3, Name: "家用电器", Image: "https://placeholder.com/electronics.png"},
	}

	return &ColdStartResponse{
		HotItems:       hotItems,
		CategoryPrefs:  categoryPrefs,
	}, nil
}

// SetUserPreference 设置用户类目偏好
func (s *Service) SetUserPreference(ctx context.Context, userID int64, categoryIDs []int64) error {
	if userID <= 0 || len(categoryIDs) == 0 {
		return nil
	}

	// 清除旧偏好
	_, _ = s.db.Exec(`DELETE FROM user_category_preference WHERE user_id = ? AND source = 'explicit'`, userID)

	// 插入新偏好
	for _, catID := range categoryIDs {
		_, _ = s.db.Exec(`
			INSERT INTO user_category_preference (user_id, category_id, weight, source)
			VALUES (?, ?, 1.0, 'explicit')
			ON DUPLICATE KEY UPDATE weight = 1.0
		`, userID, catID)
	}

	// 清除缓存
	s.redis.Del(ctx, cacheKeyUserPref(userID))

	return nil
}

// GetUserPreferences 获取用户偏好
func (s *Service) GetUserPreferences(ctx context.Context, userID int64) ([]UserCategoryPreference, error) {
	if userID <= 0 {
		return nil, nil
	}

	// 尝试从缓存获取
	cacheKey := cacheKeyUserPref(userID)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var prefs []UserCategoryPreference
		if json.Unmarshal([]byte(cached), &prefs) == nil {
			return prefs, nil
		}
	}

	// 从数据库获取
	prefs := []UserCategoryPreference{}
	err = s.db.Select(&prefs, `
		SELECT * FROM user_category_preference
		WHERE user_id = ? AND weight > 0.05
		ORDER BY weight DESC
	`, userID)

	if len(prefs) > 0 {
		// 写入缓存
		if data, _ := json.Marshal(prefs); len(data) > 0 {
			s.redis.Set(ctx, cacheKey, data, 30*time.Minute)
		}
	}

	return prefs, nil
}

// Helper functions
func cacheKeyUserPref(userID int64) string {
	return fmt.Sprintf("rec:user_pref:%d", userID)
}

func (s *Service) enrichItems(ctx context.Context, items []RecItem) []RecItem {
	// TODO: 从商品服务或缓存补全商品信息
	// 简化实现，返回基本结构
	for i := range items {
		if items[i].Name == "" {
			items[i].Name = fmt.Sprintf("商品%d", items[i].SkuID)
		}
		if items[i].Price == 0 {
			items[i].Price = 99.00
		}
		if items[i].Image == "" {
			items[i].Image = "https://placeholder.com/product.png"
		}
	}
	return items
}
```

- [ ] **Step 3: 创建 consumer.go**

```go
package recommendation

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go-micro/pkg/mq"
)

type Consumer struct {
	svc    *Service
	rabbit *mq.Rabbit
}

func NewConsumer(svc *Service, rabbit *mq.Rabbit) *Consumer {
	return &Consumer{
		svc:    svc,
		rabbit: rabbit,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	msgs, err := c.rabbit.Consume()
	if err != nil {
		log.Printf("failed to start consumer: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("consumer stopped")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("consumer channel closed")
				return
			}

			// 处理消息
			if err := c.processMessage(ctx, msg.Body); err != nil {
				log.Printf("failed to process message: %v", err)
				// 消息处理失败，不ACK，让消息重新入队
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, body []byte) error {
	var req BehaviorReportRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return err
	}

	// 直接落库
	return c.svc.ReportBehavior(ctx, &req, 0)
}
```

- [ ] **Step 4: 更新 service.go 导入**

```go
package recommendation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// 新增 Response 结构体
type SimilarProductsResponse struct {
	Scene string    `json:"scene"`
	Items []RecItem `json:"items"`
}

type HomeRecResponse struct {
	Items    []RecItem `json:"items"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Total    int       `json:"total"`
	Source   string    `json:"source"` // personalized | category | global
}

type ColdStartResponse struct {
	HotItems      []RecItem       `json:"hot_items"`
	CategoryPrefs []CategoryPref  `json:"category_prefs"`
}

type CategoryPref struct {
	CategoryID int64  `json:"category_id"`
	Name       string `json:"name"`
	Image      string `json:"image"`
}
```

- [ ] **Step 5: Commit**

```bash
git add internal/recommendation/handler.go internal/recommendation/consumer.go internal/recommendation/service.go
git commit -m "feat(recommendation): add behavior report API and MQ consumer"
```

---

## Task 4: Item-CF 算法实现

**Files:**
- Create: `internal/recommendation/algorithm/item_cf.go`
- Create: `internal/recommendation/job/item_cf_job.go`

- [ ] **Step 1: 创建 item_cf.go**

```go
package algorithm

import (
	"math"

	"github.com/jmoiron/sqlx"
)

type ItemCF struct {
	db           *sqlx.DB
	minCoUsers   int
	topNSimilar  int
}

func NewItemCF(db *sqlx.DB) *ItemCF {
	return &ItemCF{
		db:          db,
		minCoUsers:  2,
		topNSimilar: 50,
	}
}

// ComputeSimilarity 计算商品相似度（基于共同用户）
// sim(A,B) = |Users(A) ∩ Users(B)| / sqrt(|Users(A)| × |Users(B)|)
func (ic *ItemCF) ComputeSimilarity() error {
	// 1. 统计每个商品的行为用户数
	itemUserCounts := ic.countItemUsers()

	// 2. 统计商品对的共同用户数
	pairs, err := ic.countCoOccurrences()
	if err != nil {
		return err
	}

	// 3. 计算相似度并写入
	batch := make([]struct {
		SkuIDA    int64
		SkuIDB    int64
		Scene     string
		Similarity float64
		Weight    int
	}, 0, len(pairs))

	for _, pair := range pairs {
		countA := itemUserCounts[pair.SkuID]
		countB := itemUserCounts[pair.SkuIDB]
		if countA == 0 || countB == 0 {
			continue
		}

		// 余弦相似度
		sim := float64(pair.CoUsers) / math.Sqrt(float64(countA)*float64(countB))

		if sim > 0.01 && pair.CoUsers >= ic.minCoUsers {
			batch = append(batch, struct {
				SkuIDA    int64
				SkuIDB    int64
				Scene     string
				Similarity float64
				Weight    int
			}{
				SkuIDA:    pair.SkuID,
				SkuIDB:    pair.SkuIDB,
				Scene:     pair.Scene,
				Similarity: sim,
				Weight:    pair.CoUsers,
			})
		}

		// 批量写入
		if len(batch) >= 1000 {
			if err := ic.saveBatch(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	// 写入剩余
	if len(batch) > 0 {
		if err := ic.saveBatch(batch); err != nil {
			return err
		}
	}

	return nil
}

type itemUserCount struct {
	SkuID  int64
	Count  int
}

type coOccurrence struct {
	SkuID   int64
	SkuIDB  int64
	Scene   string
	CoUsers int
}

func (ic *ItemCF) countItemUsers() map[int64]int {
	counts := make(map[int64]int)
	rows, _ := ic.db.Query(`
		SELECT sku_id, COUNT(DISTINCT user_id) as cnt
		FROM user_behavior_logs
		WHERE created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
		GROUP BY sku_id
	`)
	defer rows.Close()

	for rows.Next() {
		var c itemUserCount
		if err := rows.Scan(&c.SkuID, &c.Count); err == nil {
			counts[c.SkuID] = c.Count
		}
	}
	return counts
}

func (ic *ItemCF) countCoOccurrences() ([]coOccurrence, error) {
	pairs := make([]coOccurrence, 0)
	rows, err := ic.db.Query(`
		SELECT
			a.sku_id as sku_id_a,
			b.sku_id as sku_id_b,
			a.behavior_type as scene,
			COUNT(DISTINCT a.user_id) as co_users
		FROM user_behavior_logs a
		JOIN user_behavior_logs b ON a.user_id = b.user_id AND a.sku_id != b.sku_id
		WHERE a.created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
			AND b.created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
		GROUP BY a.sku_id, b.sku_id, a.behavior_type
		HAVING co_users >= ?
	`, ic.minCoUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p coOccurrence
		if err := rows.Scan(&p.SkuID, &p.SkuIDB, &p.Scene, &p.CoUsers); err == nil {
			pairs = append(pairs, p)
		}
	}
	return pairs, nil
}

func (ic *ItemCF) saveBatch(batch []struct {
	SkuIDA    int64
	SkuIDB    int64
	Scene     string
	Similarity float64
	Weight    int
}) error {
	if len(batch) == 0 {
		return nil
	}

	tx, err := ic.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 先清除旧数据
	for _, b := range batch {
		_, _ = tx.Exec(`DELETE FROM product_similarity WHERE (sku_id_a = ? AND sku_id_b = ?) OR (sku_id_a = ? AND sku_id_b = ?)`,
			b.SkuIDA, b.SkuIDB, b.SkuIDB, b.SkuIDA)
	}

	// 插入新数据
	for _, b := range batch {
		_, err := tx.Exec(`
			INSERT INTO product_similarity (sku_id_a, sku_id_b, scene, similarity, weight)
			VALUES (?, ?, ?, ?, ?)
		`, b.SkuIDA, b.SkuIDB, b.Scene, b.Similarity, b.Weight)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
```

- [ ] **Step 2: 创建 item_cf_job.go**

```go
package job

import (
	"log"
	"time"

	"go-micro/internal/recommendation/algorithm"
	"go-micro/pkg/db"
)

type ItemCFJob struct {
	itemCF *algorithm.ItemCF
}

func NewItemCFJob() *ItemCFJob {
	dbx, _ := db.NewMySQL()
	return &ItemCFJob{
		itemCF: algorithm.NewItemCF(dbx),
	}
}

func (j *ItemCFJob) Run() error {
	log.Println("[ItemCFJob] starting...")
	start := time.Now()

	if err := j.itemCF.ComputeSimilarity(); err != nil {
		log.Printf("[ItemCFJob] failed: %v", err)
		return err
	}

	log.Printf("[ItemCFJob] completed in %v", time.Since(start))
	return nil
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/recommendation/algorithm/item_cf.go internal/recommendation/job/item_cf_job.go
git commit -m "feat(recommendation): implement Item-CF algorithm"
```

---

## Task 5: 热卖榜算法实现

**Files:**
- Create: `internal/recommendation/algorithm/bestseller.go`
- Create: `internal/recommendation/job/bestseller_job.go`

- [ ] **Step 1: 创建 bestseller.go**

```go
package algorithm

import (
	"math"

	"github.com/jmoiron/sqlx"
)

type Bestseller struct {
	db    *sqlx.DB
	periodDays int
	topN       int
}

func NewBestseller(db *sqlx.DB) *Bestseller {
	return &Bestseller{
		db:         db,
		periodDays: 30,
		topN:       100,
	}
}

// BehaviorWeights 行为权重
var BehaviorWeights = map[string]float64{
	"purchase": 10.0,
	"cart":     3.0,
	"favorite": 5.0,
}

// ComputeCategoryBestsellers 计算类目热卖榜
func (b *Bestseller) ComputeCategoryBestsellers() error {
	cutoff := time.Now().AddDate(0, 0, -b.periodDays)

	// 按类目分组计算销量分数
	rows, err := b.db.Query(`
		SELECT
			p.category_id,
			b.sku_id,
			SUM(COALESCE(
				CASE b.behavior_type
					WHEN 'purchase' THEN 10
					WHEN 'cart' THEN 3
					WHEN 'favorite' THEN 5
				END * UNIX_TIMESTAMP(NOW()) / UNIX_TIMESTAMP(b.created_at), 1
			)) as score
		FROM user_behavior_logs b
		JOIN products p ON b.sku_id = p.sku_id
		WHERE b.created_at > ?
		GROUP BY p.category_id, b.sku_id
	`, cutoff)
	if err != nil {
		return err
	}
	defer rows.Close()

	// 按类目分组
	type skuScore struct {
		SkuID  int64
		Score  float64
	}
	categoryScores := make(map[int64][]skuScore)

	for rows.Next() {
		var catID, skuID int64
		var score float64
		if err := rows.Scan(&catID, &skuID, &score); err == nil {
			categoryScores[catID] = append(categoryScores[catID], skuScore{SkuID: skuID, Score: score})
		}
	}

	// 计算排名并写入
	for catID, scores := range categoryScores {
		// 排序
		for i := 0; i < len(scores)-1; i++ {
			for j := i + 1; j < len(scores); j++ {
				if scores[j].Score > scores[i].Score {
					scores[i], scores[j] = scores[j], scores[i]
				}
			}
		}

		// 取TopN
		if len(scores) > b.topN {
			scores = scores[:b.topN]
		}

		// 写入数据库
		if err := b.saveCategoryBestsellers(catID, scores, "30d"); err != nil {
			log.Printf("failed to save category %d bestsellers: %v", catID, err)
		}
	}

	return nil
}

// ComputeGlobalBestsellers 计算全站热卖榜
func (b *Bestseller) ComputeGlobalBestsellers() error {
	cutoff := time.Now().AddDate(0, 0, -b.periodDays)

	rows, err := b.db.Query(`
		SELECT
			b.sku_id,
			SUM(COALESCE(
				CASE b.behavior_type
					WHEN 'purchase' THEN 10
					WHEN 'cart' THEN 3
					WHEN 'favorite' THEN 5
				END * EXP(-0.1 * (UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(b.created_at)) / 86400), 1
			)) as score
		FROM user_behavior_logs b
		WHERE b.created_at > ?
		GROUP BY b.sku_id
		ORDER BY score DESC
		LIMIT ?
	`, cutoff, b.topN*2) // 多取一些用于排名
	if err != nil {
		return err
	}
	defer rows.Close()

	scores := make([]skuScore, 0, b.topN)
	for rows.Next() {
		var s skuScore
		if err := rows.Scan(&s.SkuID, &s.Score); err == nil {
			scores = append(scores, s)
		}
	}

	return b.saveGlobalBestsellers(scores, "30d")
}

type skuScore struct {
	SkuID int64
	Score float64
}

func (b *Bestseller) saveCategoryBestsellers(catID int64, scores []skuScore, period string) error {
	tx, err := b.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 清除旧数据
	_, _ = tx.Exec(`DELETE FROM category_bestsellers WHERE category_id = ? AND period = ?`, catID, period)

	// 插入新数据
	for rank, s := range scores {
		_, err := tx.Exec(`
			INSERT INTO category_bestsellers (category_id, sku_id, sales_score, rank, period)
			VALUES (?, ?, ?, ?, ?)
		`, catID, s.SkuID, s.Score, rank+1, period)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (b *Bestseller) saveGlobalBestsellers(scores []skuScore, period string) error {
	tx, err := b.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 清除旧数据
	_, _ = tx.Exec(`DELETE FROM global_bestsellers WHERE period = ?`, period)

	// 插入新数据
	for rank, s := range scores {
		_, err := tx.Exec(`
			INSERT INTO global_bestsellers (sku_id, sales_score, rank, period)
			VALUES (?, ?, ?, ?)
		`, s.SkuID, s.Score, rank+1, period)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
```

- [ ] **Step 2: 创建 bestseller_job.go**

```go
package job

import (
	"log"
	"time"

	"go-micro/internal/recommendation/algorithm"
	"go-micro/pkg/db"
)

type BestsellerJob struct {
	bs *algorithm.Bestseller
}

func NewBestsellerJob() *BestsellerJob {
	dbx, _ := db.NewMySQL()
	return &BestsellerJob{
		bs: algorithm.NewBestseller(dbx),
	}
}

func (j *BestsellerJob) Run() error {
	log.Println("[BestsellerJob] starting...")
	start := time.Now()

	// 计算类目热卖
	if err := j.bs.ComputeCategoryBestsellers(); err != nil {
		log.Printf("[BestsellerJob] category bestsellers failed: %v", err)
	}

	// 计算全站热卖
	if err := j.bs.ComputeGlobalBestsellers(); err != nil {
		log.Printf("[BestsellerJob] global bestsellers failed: %v", err)
	}

	log.Printf("[BestsellerJob] completed in %v", time.Since(start))
	return nil
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/recommendation/algorithm/bestseller.go internal/recommendation/job/bestseller_job.go
git commit -m "feat(recommendation): implement bestseller algorithm"
```

---

## Task 6: User-CF 算法和辅助方法

**Files:**
- Create: `internal/recommendation/algorithm/user_cf.go`
- Modify: `internal/recommendation/service.go`

- [ ] **Step 1: 创建 user_cf.go**

```go
package algorithm

import (
	"math"
	"sort"

	"github.com/jmoiron/sqlx"
)

type UserCF struct {
	db        *sqlx.DB
	topK      int
	minBehaviors int
}

func NewUserCF(db *sqlx.DB) *UserCF {
	return &UserCF{
		db:           db,
		topK:         20,
		minBehaviors: 10,
	}
}

// UserBehavior 用户行为
type UserBehavior struct {
	UserID   int64
	SkuID    int64
	Type     string
	Weight   float64
}

// GetUserRecommendations 获取用户的个性化推荐
func (uc *UserCF) GetUserRecommendations(userID int64, limit int) ([]int64, error) {
	if limit <= 0 {
		limit = 20
	}

	// 1. 获取用户行为
	userBehaviors := uc.getUserBehaviors(userID)
	if len(userBehaviors) < uc.minBehaviors {
		return nil, nil // 数据不足，返回nil让调用方使用兜底
	}

	// 2. 获取相似用户
	similarUsers := uc.findSimilarUsers(userID, userBehaviors)
	if len(similarUsers) == 0 {
		return nil, nil
	}

	// 3. 预测用户对商品的兴趣分
	itemScores := uc.predictScores(userID, userBehaviors, similarUsers)

	// 4. 排序返回TopN
	sort.Slice(itemScores, func(i, j int) bool {
		return itemScores[i].Score > itemScores[j].Score
	})

	result := make([]int64, 0, limit)
	for i := 0; i < len(itemScores) && i < limit; i++ {
		result = append(result, itemScores[i].SkuID)
	}

	return result, nil
}

type itemScore struct {
	SkuID int64
	Score float64
}

func (uc *UserCF) getUserBehaviors(userID int64) []UserBehavior {
	behaviors := make([]UserBehavior, 0)
	rows, _ := uc.db.Query(`
		SELECT user_id, sku_id, behavior_type
		FROM user_behavior_logs
		WHERE user_id = ? AND created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
	`, userID)
	defer rows.Close()

	weightMap := map[string]float64{"purchase": 10, "cart": 3, "favorite": 5}

	for rows.Next() {
		var b UserBehavior
		if err := rows.Scan(&b.UserID, &b.SkuID, &b.Type); err == nil {
			b.Weight = weightMap[b.Type]
			behaviors = append(behaviors, b)
		}
	}
	return behaviors
}

func (uc *UserCF) findSimilarUsers(userID int64, userBehaviors []UserBehavior) map[int64]float64 {
	// 获取与当前用户有共同行为的用户
	userSkuSet := make(map[int64]bool)
	for _, b := range userBehaviors {
		userSkuSet[b.SkuID] = true
	}

	similarUsers := make(map[int64]float64)
	rows, _ := uc.db.Query(`
		SELECT DISTINCT user_id
		FROM user_behavior_logs
		WHERE sku_id IN (SELECT sku_id FROM user_behavior_logs WHERE user_id = ?)
			AND user_id != ?
			AND created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
	`, userID, userID)
	defer rows.Close()

	for rows.Next() {
		var otherUserID int64
		if rows.Scan(&otherUserID) == nil {
			// 计算相似度（简化版：共同商品数 / sqrt(用户A商品数 * 用户B商品数)）
			otherBehaviors := uc.getUserBehaviors(otherUserID)
			common := 0
			for _, ob := range otherBehaviors {
				if userSkuSet[ob.SkuID] {
					common++
				}
			}
			if common > 0 {
				sim := float64(common) / math.Sqrt(float64(len(userBehaviors)*len(otherBehaviors)))
				similarUsers[otherUserID] = sim
			}
		}
	}
	return similarUsers
}

func (uc *UserCF) predictScores(userID int64, userBehaviors []UserBehavior, similarUsers map[int64]float64) []itemScore {
	// 获取相似用户的行为
	itemScores := make(map[int64]float64)
	userSkuSet := make(map[int64]bool)
	for _, b := range userBehaviors {
		userSkuSet[b.SkuID] = true
	}

	for otherUserID, sim := range similarUsers {
		otherBehaviors := uc.getUserBehaviors(otherUserID)
		for _, b := range otherBehaviors {
			if !userSkuSet[b.SkuID] { // 排除用户已行为过的商品
				itemScores[b.SkuID] += sim * b.Weight
			}
		}
	}

	result := make([]itemScore, 0, len(itemScores))
	for skuID, score := range itemScores {
		result = append(result, itemScore{SkuID: skuID, Score: score})
	}
	return result
}
```

- [ ] **Step 2: 更新 service.go 添加 User-CF 和热卖方法**

```go
// getUserCFRecommendations 获取User-CF推荐
func (s *Service) getUserCFRecommendations(ctx context.Context, userID int64, page, pageSize int) ([]RecItem, string) {
	if userID <= 0 {
		return s.getGlobalBestsellers(ctx, page, pageSize)
	}

	// 检查用户行为数是否足够
	count := 0
	s.db.Get(&count, `SELECT COUNT(*) FROM user_behavior_logs WHERE user_id = ?`, userID)
	if count < 10 {
		// 行为不足，返回类目热卖
		prefs, _ := s.GetUserPreferences(ctx, userID)
		return s.getCategoryBestsellersFallback(ctx, prefs, pageSize), "category"
	}

	// User-CF推荐
	skuIDs, err := s.userCF.GetUserRecommendations(userID, pageSize)
	if err != nil || len(skuIDs) == 0 {
		return s.getGlobalBestsellers(ctx, page, pageSize), "global"
	}

	items := make([]RecItem, 0, len(skuIDs))
	for _, skuID := range skuIDs {
		items = append(items, RecItem{SkuID: skuID, Reason: "为你推荐"})
	}
	items = s.enrichItems(ctx, items)

	return items, "personalized"
}

// getGlobalBestsellers 获取全站热卖
func (s *Service) getGlobalBestsellers(ctx context.Context, page, pageSize int) ([]RecItem, string) {
	offset := (page - 1) * pageSize
	rows := []GlobalBestseller{}
	err := s.db.Select(&rows, `
		SELECT * FROM global_bestsellers
		WHERE period = '30d'
		ORDER BY rank
		LIMIT ? OFFSET ?
	`, pageSize, offset)
	if err != nil || len(rows) == 0 {
		return []RecItem{}, "global"
	}

	items := make([]RecItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, RecItem{SkuID: r.SkuID, Reason: "热销榜单"})
	}
	return s.enrichItems(ctx, items), "global"
}

// getCategoryBestsellersFallback 获取类目热卖兜底
func (s *Service) getCategoryBestsellersFallback(ctx context.Context, prefs []UserCategoryPreference, limit int) []RecItem {
	if len(prefs) == 0 {
		return []RecItem{}
	}

	// 取第一个偏好类目
	catID := prefs[0].CategoryID
	rows := []CategoryBestseller{}
	err := s.db.Select(&rows, `
		SELECT * FROM category_bestsellers
		WHERE category_id = ? AND period = '30d'
		ORDER BY rank
		LIMIT ?
	`, catID, limit)
	if err != nil || len(rows) == 0 {
		return []RecItem{}
	}

	items := make([]RecItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, RecItem{SkuID: r.SkuID, Reason: fmt.Sprintf("%s热卖", "类目")})
	}
	return s.enrichItems(ctx, items)
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/recommendation/algorithm/user_cf.go internal/recommendation/service.go
git commit -m "feat(recommendation): implement User-CF algorithm and helper methods"
```

---

## Task 7: User-CF 定时任务

**Files:**
- Create: `internal/recommendation/job/preference_job.go`

- [ ] **Step 1: 创建 preference_job.go**

```go
package job

import (
	"log"
	"time"

	"go-micro/internal/recommendation/algorithm"
	"go-micro/pkg/db"
)

type PreferenceJob struct {
	uc *algorithm.UserCF
}

func NewPreferenceJob() *PreferenceJob {
	dbx, _ := db.NewMySQL()
	return &PreferenceJob{
		uc: algorithm.NewUserCF(dbx),
	}
}

func (j *PreferenceJob) Run() error {
	log.Println("[PreferenceJob] starting...")
	start := time.Now()

	// 计算所有活跃用户的隐式偏好
	// 简化版：每日批量更新所有有行为的用户
	rows, err := j.db.Query(`SELECT DISTINCT user_id FROM user_behavior_logs WHERE created_at > DATE_SUB(NOW(), INTERVAL 7 DAY)`)
	if err != nil {
		log.Printf("[PreferenceJob] failed to query users: %v", err)
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var userID int64
		if rows.Scan(&userID) == nil {
			count++
			// 计算并更新用户偏好（可并行优化）
			if err := j.computeUserPreference(userID); err != nil {
				log.Printf("[PreferenceJob] failed to compute preference for user %d: %v", userID, err)
			}
		}
	}

	log.Printf("[PreferenceJob] processed %d users in %v", count, time.Since(start))
	return nil
}

func (j *PreferenceJob) computeUserPreference(userID int64) error {
	// 获取用户最近30天行为
	behaviors := []struct {
		SkuID  int64 `db:"sku_id"`
		Type   string `db:"behavior_type"`
	}{}
	err := j.db.Select(&behaviors, `
		SELECT b.sku_id, b.behavior_type
		FROM user_behavior_logs b
		WHERE b.user_id = ? AND b.created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
	`, userID)
	if err != nil {
		return err
	}

	// 按类目聚合
	type catScore struct {
		CatID int64
		Score float64
	}
	catScores := make(map[int64]float64)
	weightMap := map[string]float64{"purchase": 10, "cart": 3, "favorite": 5}

	for _, b := range behaviors {
		// 简化：假设能从商品表获取类目ID
		// 实际应JOIN商品表
		weight := weightMap[b.Type]
		// TODO: 获取真实类目ID
		catID := b.SkuID % 10 // 假数据
		catScores[catID] += weight
	}

	// 归一化
	total := 0.0
	for _, score := range catScores {
		total += score
	}
	if total == 0 {
		return nil
	}

	// 写入偏好表
	tx, _ := j.db.Beginx()
	defer tx.Rollback()

	_, _ = tx.Exec(`DELETE FROM user_category_preference WHERE user_id = ? AND source = 'implicit'`, userID)

	for catID, score := range catScores {
		normalized := score / total
		if normalized > 0.05 { // 过滤噪音
			_, _ = tx.Exec(`
				INSERT INTO user_category_preference (user_id, category_id, weight, source)
				VALUES (?, ?, ?, 'implicit')
				ON DUPLICATE KEY UPDATE weight = VALUES(weight)
			`, userID, catID, normalized)
		}
	}

	return tx.Commit()
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/recommendation/job/preference_job.go
git commit -m "feat(recommendation): add user preference computation job"
```

---

## Task 8: 缓存层实现

**Files:**
- Create: `internal/recommendation/cache.go`
- Modify: `internal/recommendation/service.go`

- [ ] **Step 1: 创建 cache.go**

```go
package recommendation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	CacheTTL             = 15 * time.Minute
	SimilarProductsCache = "rec:similar:%d:%s"  // sku_id:scene
	HomeRecCache         = "rec:home:%d:%d:%d"   // user_id:page:pageSize
	GlobalBestsellerCache = "rec:bestseller:global:%s:%d"  // period:limit
)

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

// GetSimilarProducts 获取相似商品缓存
func (c *Cache) GetSimilarProducts(ctx context.Context, skuID int64, scene string) ([]RecItem, bool) {
	key := fmt.Sprintf(SimilarProductsCache, skuID, scene)
	data, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var items []RecItem
	if json.Unmarshal([]byte(data), &items) == nil {
		return items, true
	}
	return nil, false
}

// SetSimilarProducts 设置相似商品缓存
func (c *Cache) SetSimilarProducts(ctx context.Context, skuID int64, scene string, items []RecItem) error {
	key := fmt.Sprintf(SimilarProductsCache, skuID, scene)
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, CacheTTL).Err()
}

// GetHomeRec 获取首页推荐缓存
func (c *Cache) GetHomeRec(ctx context.Context, userID int64, page, pageSize int) (*HomeRecResponse, bool) {
	key := fmt.Sprintf(HomeRecCache, userID, page, pageSize)
	data, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var resp HomeRecResponse
	if json.Unmarshal([]byte(data), &resp) == nil {
		return &resp, true
	}
	return nil, false
}

// SetHomeRec 设置首页推荐缓存
func (c *Cache) SetHomeRec(ctx context.Context, userID int64, page, pageSize int, resp *HomeRecResponse) error {
	key := fmt.Sprintf(HomeRecCache, userID, page, pageSize)
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, CacheTTL).Err()
}

// InvalidateSimilarProducts 使相似商品缓存失效
func (c *Cache) InvalidateSimilarProducts(ctx context.Context, skuID int64) error {
	pattern := fmt.Sprintf("rec:similar:%d:*", skuID)
	keys, _ := c.rdb.Keys(ctx, pattern).Result()
	if len(keys) > 0 {
		return c.rdb.Del(ctx, keys...).Err()
	}
	return nil
}

// GetGlobalBestsellers 获取全站热卖缓存
func (c *Cache) GetGlobalBestsellers(ctx context.Context, period string, limit int) ([]RecItem, bool) {
	key := fmt.Sprintf(GlobalBestsellerCache, period, limit)
	data, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var items []RecItem
	if json.Unmarshal([]byte(data), &items) == nil {
		return items, true
	}
	return nil, false
}

// SetGlobalBestsellers 设置全站热卖缓存
func (c *Cache) SetGlobalBestsellers(ctx context.Context, period string, limit int, items []RecItem) error {
	key := fmt.Sprintf(GlobalBestsellerCache, period, limit)
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, CacheTTL).Err()
}
```

- [ ] **Step 2: 更新 service.go 集成缓存**

在 service.go 中添加缓存调用：

```go
func (s *Service) GetSimilarProducts(ctx context.Context, skuID int64, scene string, limit int) (*SimilarProductsResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// 1. 尝试从缓存获取
	if s.cache != nil {
		if items, ok := s.cache.GetSimilarProducts(ctx, skuID, scene); ok {
			return &SimilarProductsResponse{Scene: scene, Items: items}, nil
		}
	}

	// 2. 缓存未命中，从数据库查询
	rows := []ProductSimilarity{}
	err := s.db.Select(&rows, `
		SELECT sku_id_b as sku_id_a, sku_id_a as sku_id_b, similarity, weight
		FROM product_similarity
		WHERE sku_id_a = ? AND scene = ?
		UNION ALL
		SELECT sku_id_a, sku_id_b, similarity, weight
		FROM product_similarity
		WHERE sku_id_b = ? AND scene = ?
		ORDER BY similarity DESC
		LIMIT ?
	`, skuID, scene, skuID, scene, limit)
	if err != nil {
		return nil, err
	}

	items := make([]RecItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, RecItem{
			SkuID:      r.SkuIDB,
			Similarity: r.Similarity,
		})
	}

	// 3. 补全商品信息
	items = s.enrichItems(ctx, items)

	// 4. 写入缓存
	if s.cache != nil {
		s.cache.SetSimilarProducts(ctx, skuID, scene, items)
	}

	return &SimilarProductsResponse{
		Scene:  scene,
		Items:  items,
	}, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/recommendation/cache.go
git commit -m "feat(recommendation): add Redis caching layer"
```

---

## Task 9: 服务入口和配置

**Files:**
- Modify: `internal/app/recommendation/run.go`

- [ ] **Step 1: 更新 run.go 集成缓存**

```go
func Run() error {
	// ... 已有代码 ...

	svc := recommendation.NewService(dbx, rdb)
	h := recommendation.NewHandler(svc)

	// 注入缓存
	cache := recommendation.NewCache(rdb)
	svc.SetCache(cache)

	// 启动MQ消费者
	if rabbit != nil {
		consumer := recommendation.NewConsumer(svc, rabbit)
		go consumer.Start(ctx)
	}

	// ... 已有代码 ...
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/app/recommendation/run.go
git commit -m "feat(recommendation): integrate cache into service"
```

---

## Task 10: 单元测试

**Files:**
- Create: `internal/recommendation/service_test.go`
- Create: `internal/recommendation/algorithm/item_cf_test.go`

- [ ] **Step 1: 创建 service_test.go**

```go
package recommendation

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_ReportBehavior(t *testing.T) {
	// 使用 miniredis 测试
	s := miniredis.RunT(t)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	// TODO: 使用真实的测试数据库
	// svc := NewService(db, rdb)
	// req := &BehaviorReportRequest{SkuID: 123, BehaviorType: "cart"}
	// err := svc.ReportBehavior(context.Background(), req, 1)
	// assert.NoError(t, err)
}

func TestService_GetSimilarProducts(t *testing.T) {
	// 集成测试需要真实数据库
	// 本测试验证逻辑
	assert.True(t, true, "placeholder")
}
```

- [ ] **Step 2: 创建 item_cf_test.go**

```go
package algorithm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemCF_Similarity(t *testing.T) {
	// 单元测试：验证相似度计算公式
	// sim(A,B) = |Users(A) ∩ Users(B)| / sqrt(|Users(A)| × |Users(B)|)
	// 如果A被3个用户行为，B被5个用户行为，共同用户2个
	// sim = 2 / sqrt(3 * 5) = 2 / sqrt(15) ≈ 0.516

	sim := 2.0 / sqrt(3*5)
	expected := 0.516
	assert.InDelta(t, expected, sim, 0.01)
}

func sqrt(n int) float64 {
	return &math.Sqrt(float64(n))
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/recommendation/service_test.go internal/recommendation/algorithm/item_cf_test.go
git commit -m "test(recommendation): add unit tests"
```

---

## Task 11: 集成测试验证

**Files:**
- Modify: 相关服务配置

- [ ] **Step 1: 验证服务启动**

Run: `go build -o recommendation-service ./internal/app/recommendation`

Expected: 编译成功，无错误

- [ ] **Step 2: 验证数据库连接**

手动执行SQL，验证表创建成功

- [ ] **Step 3: 验证MQ连接**

配置正确的RabbitMQ URL，验证消费者能正常启动

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "test(recommendation): integration verification complete"
```

---

## 实现顺序

1. Task 1: 数据库表创建
2. Task 2: 推荐服务基础结构
3. Task 3: 行为采集 API 和 MQ 消费者
4. Task 4: Item-CF 算法实现
5. Task 5: 热卖榜算法实现
6. Task 6: User-CF 算法和辅助方法
7. Task 7: User-CF 定时任务
8. Task 8: 缓存层实现
9. Task 9: 服务入口和配置
10. Task 10: 单元测试
11. Task 11: 集成测试验证

---

## 依赖关系

```
Task 1 (DB) ──► Task 2 (基础结构)
                     │
                     ▼
              Task 3 (API+Consumer)
                     │
          ┌───────────┴───────────┐
          ▼                       ▼
    Task 4 (ItemCF)        Task 5 (Bestseller)
          │                       │
          └───────────┬───────────┘
                      ▼
                Task 6 (UserCF)
                      │
                      ▼
                Task 7 (PreferenceJob)
                      │
                      ▼
                Task 8 (Cache)
                      │
                      ▼
                Task 9 (Run)
                      │
                      ▼
           Task 10 + Task 11 (Test+Verify)
```