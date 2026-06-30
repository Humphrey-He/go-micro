# 推荐服务缺失功能修复计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复推荐服务的关键缺失功能，使服务能够正常运行并提供完整的推荐能力

**Context:** 根据实现对比报告（docs/recommendation-implementation-review.md），推荐服务的核心算法已完成，但缺少运行时基础设施和部分业务逻辑

**Priority:** P0（立即修复）+ P1（近期完成）

---

## Task 1: 实现 Redis 缓存层 (P0)

**Priority:** P0 - 影响性能，每次请求都查数据库

**Files:**
- Modify: `internal/recommendation/cache.go`
- Modify: `internal/recommendation/service.go`

**Current State:**
- Cache 结构体存在但所有方法返回空/false
- Service 中有缓存调用但永远未命中

**Requirements:**
- 实现 GetSimilarProducts/SetSimilarProducts
- 实现 GetHomeRec/SetHomeRec
- 实现 InvalidateSimilarProducts
- 缓存 TTL 设置为 15 分钟
- Service 中集成缓存调用（先查缓存，未命中查DB，结果写回缓存）

- [ ] **Step 1: 实现 cache.go 中的所有方法**

实现以下方法：
- `GetSimilarProducts(ctx, skuID, scene) ([]RecItem, bool)`
- `SetSimilarProducts(ctx, skuID, scene, items)`
- `GetHomeRec(ctx, userID, page, pageSize) (*HomeRecResponse, bool)`
- `SetHomeRec(ctx, userID, page, pageSize, resp)`
- `InvalidateSimilarProducts(ctx, skuID)`

使用 Redis 的 Get/Set 操作，JSON 序列化数据，TTL 15分钟

- [ ] **Step 2: 在 service.go 中集成缓存**

修改 `GetSimilarProducts` 方法：
1. 先调用 cache.GetSimilarProducts
2. 命中则直接返回
3. 未命中则查数据库
4. 查询结果调用 cache.SetSimilarProducts 写回

修改 `GetHomeRecommendations` 方法：
1. 先调用 cache.GetHomeRec
2. 命中则直接返回
3. 未命中则计算推荐
4. 结果调用 cache.SetHomeRec 写回

- [ ] **Step 3: 在 service.go 中添加 cache 字段和 setter**

```go
type Service struct {
    db    *sqlx.DB
    redis *redis.Client
    cache *Cache
    // ... 其他字段
}

func (s *Service) SetCache(cache *Cache) {
    s.cache = cache
}
```

- [ ] **Step 4: 测试缓存功能**

编写简单测试验证：
- 第一次查询未命中，查数据库
- 第二次查询命中缓存
- TTL 过期后重新查数据库

- [ ] **Step 5: Commit**

```bash
git add internal/recommendation/cache.go internal/recommendation/service.go
git commit -m "feat(recommendation): implement Redis caching layer"
```

---

## Task 2: 添加定时任务调度器 (P0)

**Priority:** P0 - 离线算法无法运行，推荐结果不会更新

**Files:**
- Create: `internal/recommendation/scheduler.go`
- Modify: `internal/app/recommendation/run.go`

**Current State:**
- Job 代码完整（item_cf_job.go, bestseller_job.go, preference_job.go, association_job.go）
- 但没有调度器启动这些 Job

**Requirements:**
- 使用 cron 调度器
- Item-CF Job: 每日 02:00
- Bestseller Job: 每日 03:00
- Preference Job: 每日 04:00
- Association Job: 每日 01:00
- 在服务启动时启动调度器

- [ ] **Step 1: 创建 scheduler.go**

```go
package recommendation

import (
    "context"
    "log"
    
    "github.com/robfig/cron/v3"
    "go-micro/internal/recommendation/job"
)

type Scheduler struct {
    cron *cron.Cron
    db   *sqlx.DB
}

func NewScheduler(db *sqlx.DB) *Scheduler {
    return &Scheduler{
        cron: cron.New(),
        db:   db,
    }
}

func (s *Scheduler) Start(ctx context.Context) error {
    // Association Job - 每日 01:00
    s.cron.AddFunc("0 1 * * *", func() {
        log.Println("[Scheduler] Running AssociationJob...")
        job.NewAssociationJob(s.db).Run()
    })
    
    // Item-CF Job - 每日 02:00
    s.cron.AddFunc("0 2 * * *", func() {
        log.Println("[Scheduler] Running ItemCFJob...")
        job.NewItemCFJob(s.db).Run()
    })
    
    // Bestseller Job - 每日 03:00
    s.cron.AddFunc("0 3 * * *", func() {
        log.Println("[Scheduler] Running BestsellerJob...")
        job.NewBestsellerJob(s.db).Run()
    })
    
    // Preference Job - 每日 04:00
    s.cron.AddFunc("0 4 * * *", func() {
        log.Println("[Scheduler] Running PreferenceJob...")
        job.NewPreferenceJob(s.db).Run()
    })
    
    s.cron.Start()
    log.Println("[Scheduler] Started with 4 jobs")
    
    <-ctx.Done()
    s.cron.Stop()
    log.Println("[Scheduler] Stopped")
    return nil
}
```

- [ ] **Step 2: 修改 Job 构造函数接受 db 参数**

修改所有 Job 的 New 函数，接受 `*sqlx.DB` 参数而不是内部创建连接：
- `NewItemCFJob(db *sqlx.DB)`
- `NewBestsellerJob(db *sqlx.DB)`
- `NewPreferenceJob(db *sqlx.DB)`
- `NewAssociationJob(db *sqlx.DB)`

- [ ] **Step 3: 在 run.go 中启动调度器**

```go
func Run() error {
    // ... 已有代码 ...
    
    // 启动定时任务调度器
    scheduler := recommendation.NewScheduler(dbx)
    go scheduler.Start(ctx)
    
    // ... 已有代码 ...
}
```

- [ ] **Step 4: 添加 cron 依赖**

```bash
go get github.com/robfig/cron/v3
```

- [ ] **Step 5: Commit**

```bash
git add internal/recommendation/scheduler.go internal/app/recommendation/run.go internal/recommendation/job/
git commit -m "feat(recommendation): add cron scheduler for offline jobs"
```

---

## Task 3: 实现用户偏好计算 (P1)

**Priority:** P1 - 影响 User-CF 冷启动效果

**Files:**
- Modify: `internal/recommendation/service.go`
- Modify: `internal/recommendation/job/preference_job.go`

**Current State:**
- `SetUserPreference` 方法为空实现
- `preference_job.go` 中的 `computeUserPreference` 使用假数据

**Requirements:**
- 实现显式偏好设置（用户手动选择类目）
- 实现隐式偏好计算（从用户行为推导）
- 行为权重：purchase=10, favorite=5, cart=3
- 归一化后过滤 < 0.05 的噪音

- [ ] **Step 1: 完善 SetUserPreference 方法**

```go
func (s *Service) SetUserPreference(ctx context.Context, userID int64, categoryIDs []int64) error {
    if userID <= 0 || len(categoryIDs) == 0 {
        return nil
    }
    
    tx, err := s.db.Beginx()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // 清除旧的显式偏好
    _, err = tx.Exec(`DELETE FROM user_category_preference WHERE user_id = ? AND source = 'explicit'`, userID)
    if err != nil {
        return err
    }
    
    // 插入新偏好
    for _, catID := range categoryIDs {
        _, err := tx.Exec(`
            INSERT INTO user_category_preference (user_id, category_id, weight, source)
            VALUES (?, ?, 1.0, 'explicit')
        `, userID, catID)
        if err != nil {
            return err
        }
    }
    
    if err := tx.Commit(); err != nil {
        return err
    }
    
    // 清除缓存
    cacheKey := fmt.Sprintf("rec:user_pref:%d", userID)
    s.redis.Del(ctx, cacheKey)
    
    return nil
}
```

- [ ] **Step 2: 修复 preference_job.go 中的隐式偏好计算**

修改 `computeUserPreference` 方法，从 products 表 JOIN 获取真实类目 ID：

```go
func (j *PreferenceJob) computeUserPreference(userID int64) error {
    // 获取用户最近30天行为，JOIN products 表获取类目
    type behaviorWithCategory struct {
        CategoryID int64  `db:"category_id"`
        Type       string `db:"behavior_type"`
    }
    
    behaviors := []behaviorWithCategory{}
    err := j.db.Select(&behaviors, `
        SELECT p.category_id, b.behavior_type
        FROM user_behavior_logs b
        JOIN products p ON b.sku_id = p.sku_id
        WHERE b.user_id = ? AND b.created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
    `, userID)
    if err != nil {
        return err
    }
    
    if len(behaviors) == 0 {
        return nil
    }
    
    // 按类目聚合
    catScores := make(map[int64]float64)
    weightMap := map[string]float64{"purchase": 10, "cart": 3, "favorite": 5}
    
    for _, b := range behaviors {
        weight := weightMap[b.Type]
        catScores[b.CategoryID] += weight
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
    tx, err := j.db.Beginx()
    if err != nil {
        return err
    }
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

- [ ] **Step 3: Commit**

```bash
git add internal/recommendation/service.go internal/recommendation/job/preference_job.go
git commit -m "feat(recommendation): implement user preference computation"
```

---

## Task 4: 添加时间衰减到热卖榜 (P1)

**Priority:** P1 - 老商品权重过高，新品难上榜

**Files:**
- Modify: `internal/recommendation/algorithm/bestseller.go`

**Current State:**
- 热卖榜计算缺少时间衰减因子
- 导致老商品权重过高

**Requirements:**
- 添加指数衰减：`e^(-λ×天数)`
- λ = 0.1（衰减系数）
- 在 SQL 中计算：`EXP(-0.1 * (UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(created_at)) / 86400)`

- [ ] **Step 1: 修改 ComputeCategoryBestsellers 添加时间衰减**

```go
func (b *Bestseller) ComputeCategoryBestsellers() error {
    query := fmt.Sprintf(`
        SELECT
            COALESCE(p.category_id, 0) as category_id,
            b.sku_id,
            SUM(
                CASE b.behavior_type
                    WHEN 'purchase' THEN %d
                    WHEN 'cart' THEN %d
                    WHEN 'favorite' THEN %d
                END * EXP(-0.1 * (UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(b.created_at)) / 86400)
            ) as score
        FROM user_behavior_logs b
        LEFT JOIN products p ON b.sku_id = p.sku_id
        WHERE b.created_at > DATE_SUB(NOW(), INTERVAL %d DAY)
            AND b.user_id > 0
        GROUP BY COALESCE(p.category_id, 0), b.sku_id
    `, WeightPurchase, WeightCart, WeightFavorite, DefaultPeriodDays)
    
    // ... 其余逻辑不变
}
```

- [ ] **Step 2: 修改 ComputeGlobalBestsellers 添加时间衰减**

```go
func (b *Bestseller) ComputeGlobalBestsellers() error {
    query := fmt.Sprintf(`
        SELECT
            b.sku_id,
            SUM(
                CASE b.behavior_type
                    WHEN 'purchase' THEN %d
                    WHEN 'cart' THEN %d
                    WHEN 'favorite' THEN %d
                END * EXP(-0.1 * (UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(b.created_at)) / 86400)
            ) as score
        FROM user_behavior_logs b
        WHERE b.created_at > DATE_SUB(NOW(), INTERVAL %d DAY)
            AND b.user_id > 0
        GROUP BY b.sku_id
        ORDER BY score DESC
        LIMIT ?
    `, WeightPurchase, WeightCart, WeightFavorite, DefaultPeriodDays)
    
    // ... 其余逻辑不变
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/recommendation/algorithm/bestseller.go
git commit -m "feat(recommendation): add time decay to bestseller algorithm"
```

---

## Task 5: 实现商品信息补全 (P1)

**Priority:** P1 - 推荐结果缺少商品详情

**Files:**
- Modify: `internal/recommendation/service.go`

**Current State:**
- `enrichItems` 方法返回占位符数据
- 推荐结果缺少真实商品信息

**Requirements:**
- 从 products 表查询商品信息
- 批量查询优化性能
- 缓存商品信息（可选）

- [ ] **Step 1: 实现 enrichItems 方法**

```go
func (s *Service) enrichItems(ctx context.Context, items []RecItem) []RecItem {
    if len(items) == 0 {
        return items
    }
    
    // 收集所有 SKU ID
    skuIDs := make([]int64, 0, len(items))
    for _, item := range items {
        skuIDs = append(skuIDs, item.SkuID)
    }
    
    // 批量查询商品信息
    type productInfo struct {
        SkuID int64   `db:"sku_id"`
        Name  string  `db:"name"`
        Price int64   `db:"price"`
        Image string  `db:"image"`
    }
    
    query, args, err := sqlx.In(`
        SELECT sku_id, name, price, image
        FROM products
        WHERE sku_id IN (?)
    `, skuIDs)
    if err != nil {
        return items
    }
    
    query = s.db.Rebind(query)
    rows := []productInfo{}
    if err := s.db.Select(&rows, query, args...); err != nil {
        return items
    }
    
    // 构建 SKU ID -> 商品信息映射
    productMap := make(map[int64]productInfo)
    for _, p := range rows {
        productMap[p.SkuID] = p
    }
    
    // 补全商品信息
    for i := range items {
        if p, ok := productMap[items[i].SkuID]; ok {
            items[i].Name = p.Name
            items[i].Price = float64(p.Price) / 100.0 // 分转元
            items[i].Image = p.Image
        } else {
            // 商品不存在或已下架，使用占位符
            items[i].Name = fmt.Sprintf("商品%d", items[i].SkuID)
            items[i].Price = 0
            items[i].Image = ""
        }
    }
    
    return items
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/recommendation/service.go
git commit -m "feat(recommendation): implement product info enrichment"
```

---

## 实现顺序

按优先级和依赖关系：

1. **Task 1: Redis 缓存层** (P0) - 独立，优先修复
2. **Task 2: 定时任务调度器** (P0) - 独立，优先修复
3. **Task 3: 用户偏好计算** (P1) - 依赖 Task 2（调度器）
4. **Task 4: 时间衰减** (P1) - 独立，算法优化
5. **Task 5: 商品信息补全** (P1) - 独立，数据补全

---

## 依赖关系

```
Task 1 (Cache) ──► 独立
Task 2 (Scheduler) ──► Task 3 (Preference)
Task 4 (Time Decay) ──► 独立
Task 5 (Enrich Items) ──► 独立
```

---

## 验收标准

**Task 1 完成标准:**
- [ ] Cache 所有方法实现完整
- [ ] Service 中缓存调用正常
- [ ] 第二次查询命中缓存

**Task 2 完成标准:**
- [ ] Scheduler 启动成功
- [ ] 4个 Job 注册到 cron
- [ ] 日志显示调度器运行

**Task 3 完成标准:**
- [ ] 显式偏好设置成功写入数据库
- [ ] 隐式偏好从真实类目计算
- [ ] 归一化和过滤逻辑正确

**Task 4 完成标准:**
- [ ] SQL 包含时间衰减公式
- [ ] 新商品权重高于老商品

**Task 5 完成标准:**
- [ ] 推荐结果包含真实商品信息
- [ ] 批量查询性能良好
- [ ] 不存在的商品使用占位符
