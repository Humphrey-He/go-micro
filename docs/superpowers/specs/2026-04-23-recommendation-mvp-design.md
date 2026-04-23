# 个性化推荐服务 MVP 技术设计

> 本文档定义推荐服务MVP的技术实现方案，聚焦数据可靠性和推荐效果。
>
> 版本：v1.0
>
> 日期：2026-04-23

---

## 一、系统概述

### 1.1 核心目标

1. **数据可靠**：用户行为数据稳定落库，真实反映用户意图
2. **推荐可信**：基于高质量数据，提供可信的个性化推荐
3. **快速验证**：MVP快速上线，验证数据链路和算法效果

### 1.2 推荐场景

| 场景 | 位置 | 算法 | 说明 |
|------|------|------|------|
| **看了又看** | 商品详情页 | Item-CF | 基于共同行为计算商品相似度 |
| **首页推荐** | 首页 | User-CF + 热门 | 个性化 + 热卖兜底 |

### 1.3 技术架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           推荐服务架构                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  [前端]                                                                  │
│     │                                                                   │
│     ├──加购/收藏/购买──────► [Gateway] ──────► [RabbitMQ]                │
│     │                                        │                          │
│     │                                        ▼                          │
│     │                              [推荐服务 Consumer]                   │
│     │                                        │                          │
│     │                    ┌───────────────────┼───────────────────┐      │
│     │                    ▼                   ▼                   ▼      │
│     │              [MySQL主库]         [Redis]            [计算任务]    │
│     │              (行为数据)          (缓存)             (离线)        │
│     │                    │                   │                   │      │
│     │                    └───────────────────┼───────────────────┘      │
│     │                                        │                          │
│     └───────────────────────────────────────┼──────────────────────────►│
│         查询推荐                               [MySQL从库-只读]            │
│                                                │                         │
│                                                ▼                         │
│                              [推荐服务 Read API]                          │
│                              "看了又看" / "首页推荐"                       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 二、数据模型

### 2.1 用户行为表（核心表）

```sql
CREATE TABLE user_behavior_logs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id             BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    anonymous_id        VARCHAR(64) COMMENT '匿名用户ID（未登录）',
    sku_id              BIGINT UNSIGNED NOT NULL COMMENT '商品SKU ID',
    behavior_type       ENUM('cart', 'favorite', 'purchase') NOT NULL COMMENT '行为类型',
    source              VARCHAR(32) COMMENT '来源: detail/cart/favorite/recommendation',
    stay_duration       INT COMMENT '停留时长(秒)，仅参考',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- 核心索引
    INDEX idx_user_id_type (user_id, behavior_type, created_at),
    INDEX idx_sku_id (sku_id),
    INDEX idx_anonymous_id (anonymous_id),
    INDEX idx_created_at (created_at),

    -- 去重索引：同一用户对同一商品5分钟内同类行为只记1条
    UNIQUE KEY uk_user_sku_type_5min (user_id, anonymous_id, sku_id, behavior_type, time_bucket)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户行为日志表';

-- 5分钟时间桶计算：FLOOR(UNIX_TIMESTAMP(created_at) / 300)
```

### 2.2 商品相似度表

```sql
CREATE TABLE product_similarity (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id_a        BIGINT UNSIGNED NOT NULL COMMENT '商品A',
    sku_id_b        BIGINT UNSIGNED NOT NULL COMMENT '商品B',
    scene           ENUM('view', 'cart', 'purchase') NOT NULL COMMENT '场景',
    similarity      DECIMAL(10,6) NOT NULL COMMENT '相似度分数 0-1',
    weight          INT DEFAULT 1 COMMENT '共同行为次数',
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (sku_id_a, sku_id_b, scene),
    INDEX idx_sku_a_scene (sku_id_a, scene),
    INDEX idx_similarity (similarity DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品相似度表';
```

### 2.3 用户偏好表

```sql
CREATE TABLE user_category_preference (
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
```

### 2.4 类目热卖表

```sql
CREATE TABLE category_bestsellers (
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
```

---

## 三、行为采集模块

### 3.1 数据采集API

```
POST /api/v1/rec/report

Headers:
    Authorization: Bearer {token}  // 可选，未登录用户不传

Request:
{
    "sku_id": 12345,
    "behavior_type": "purchase",  // cart | favorite | purchase
    "source": "detail",           // detail | cart | favorite | recommendation
    "stay_duration": 30,          // 停留时长(秒)，可选
    "anonymous_id": "uuid-xxx"     // 未登录用户匿名ID
}

Response:
{
    "code": 0,
    "message": "success"
}
```

### 3.2 采集可靠性保障

```
┌─────────────────────────────────────────────────────────────────┐
│                     数据采集可靠性设计                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  [前端]                                                           │
│     │                                                            │
│     │  1. 发送行为数据                                            │
│     ▼                                                            │
│  [Gateway]                                                        │
│     │                                                            │
│     │  2. 快速写入RabbitMQ（持久化消息）                           │
│     ▼                                                            │
│  [RabbitMQ]  durable=true, delivery_mode=2                        │
│     │                                                            │
│     │  3. Consumer消费                                            │
│     ▼                                                            │
│  [推荐服务]                                                        │
│     │                                                            │
│     │  4. 落库（MySQL主库）                                       │
│     │  5. ACK确认                                                 │
│     │                                                            │
│     │  [失败重试]                                                 │
│     │  - 重试3次，间隔 1s → 5s → 15s                              │
│     │  - 超过重试次数 → 写入死信队列 → 告警                        │
│     │                                                            │
│  [MySQL]                                                          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.3 去重策略

同一用户对同一商品5分钟内同类行为只记录1条：

```sql
-- 使用时间桶实现去重
INSERT INTO user_behavior_logs
    (user_id, anonymous_id, sku_id, behavior_type, source, time_bucket, created_at)
VALUES
    (?, ?, ?, ?, ?, FLOOR(UNIX_TIMESTAMP(NOW()) / 300), NOW())
ON DUPLICATE KEY UPDATE id=id;  -- 忽略重复
```

### 3.4 行为权重

| 行为类型 | 权重 | 说明 |
|----------|------|------|
| cart | 3 | 加购是强购买意图 |
| favorite | 5 | 收藏是明确兴趣 |
| purchase | 10 | 购买是最终转化 |

---

## 四、推荐算法

### 4.1 Item-CF（看了又看）

#### 算法原理

两个商品相似 = 同时对它们产生行为的用户数量 / sqrt(对商品A的用户数 × 对商品B的用户数)

```
sim(A, B) = |Users(A) ∩ Users(B)| / sqrt(|Users(A)| × |Users(B)|)
```

#### 相似度计算任务

```go
// 定时任务：每日凌晨计算商品相似度
type ItemCFJob struct {
    db      *sqlx.DB
    redis   *redis.Client
}

func (j *ItemCFJob) Run() error {
    // 1. 统计每个商品的行为用户数
    itemUsers := j.countItemUsers()

    // 2. 统计商品对的共同用户数（仅计算权重>=2的）
    itemPairs := j.countCoOccurrences()

    // 3. 计算相似度并写入表
    batchSize := 1000
    for _, pair := range itemPairs {
        sim := float64(pair.CoUsers) / math.Sqrt(float64(itemUsers[pair.ItemA])*float64(itemUsers[pair.ItemB]))
        if sim > 0.01 {  // 阈值过滤
            j.saveSimilarity(pair.ItemA, pair.ItemB, sim, pair.Weight)
        }
    }

    // 4. 更新缓存
    j.refreshCache()
    return nil
}
```

#### 查询推荐

```go
func (s *RecommendationService) GetSimilarProducts(skuID string, scene string, limit int) ([]RecItem, error) {
    // 1. 查询相似商品（从从库读取）
    similarities, err := s.getSimilarFromDB(skuID, scene, limit)
    if err != nil {
        return nil, err
    }

    // 2. 过滤已下架/无库存商品
    items, err := s.filterAvailable(similarities)
    if err != nil {
        return nil, err
    }

    // 3. 补全商品信息
    return s.enrichItems(items), nil
}
```

### 4.2 User-CF（首页推荐）

#### 算法原理

```
预测用户U对商品I的兴趣 = Σ(与U相似用户V对I的行为 × 相似度U,V)
```

#### 冷启动策略

```
┌─────────────────────────────────────────────────────────────┐
│                    用户冷启动流程                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  [新用户] ──有类目偏好？───是───► [该类目热卖]                │
│      │                                │                      │
│      │没有                            │                     │
│      ▼                                ▼                     │
│  [全站热卖 + 类目引导入口] ◄─────────────────┘               │
│                                                              │
│  [有行为用户] ──行为数>=10？───是───► [User-CF个性化]         │
│                    │                                         │
│                    │不够                                     │
│                    ▼                                         │
│              [类目热卖 + 热销兜底]                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

#### 类目偏好计算

```go
// 用户隐式偏好计算：根据行为历史计算类目权重
func (s *UserPreferenceService) ComputeImplicitPreference(userID int64) error {
    // 1. 获取用户最近30天行为
    behaviors, _ := s.getUserBehaviors(userID, 30)

    // 2. 按类目聚合
    categoryScores := make(map[int64]float64)
    for _, b := range behaviors {
        weight := behaviorWeight[b.Type]
        categoryID := s.getCategoryID(b.SkuID)
        categoryScores[categoryID] += weight
    }

    // 3. 归一化并写入
    total := sum(categoryScores)
    for catID, score := range categoryScores {
        normalizedScore := score / total
        if normalizedScore > 0.05 {  // 过滤噪音
            s.savePreference(userID, catID, normalizedScore, "implicit")
        }
    }
    return nil
}
```

### 4.3 类目热卖榜

```go
// 计算类目热卖榜（每日定时任务）
func (j *BestsellerJob) ComputeCategoryBestsellers() error {
    periods := []string{"7d", "30d"}
    for _, period := range periods {
        cutoff := time.Now().AddDate(0, 0, -parsePeriodDays(period))

        // 1. 按类目分组，计算销量分数
        // 销量分数 = Σ(purchase×10 + cart×3 + favorite×5) × 时间衰减
        rows, err := j.db.Query(`
            SELECT p.category_id, b.sku_id, SUM(
                CASE b.behavior_type
                    WHEN 'purchase' THEN 10
                    WHEN 'cart' THEN 3
                    WHEN 'favorite' THEN 5
                END * UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(b.created_at)
            ) as score
            FROM user_behavior_logs b
            JOIN products p ON b.sku_id = p.sku_id
            WHERE b.created_at > ?
            GROUP BY p.category_id, b.sku_id
        `, cutoff)
        // ...
    }
    return nil
}
```

---

## 五、API接口

### 5.1 看了又看

```
GET /api/v1/rec/similar/{sku_id}

Query:
    scene = cart       // cart: 加购场景 | purchase: 购买场景
    limit = 10

Response:
{
    "code": 0,
    "data": {
        "scene": "cart",
        "items": [
            {
                "sku_id": 12346,
                "name": "iPhone 15 专用手机壳",
                "price": 99.00,
                "image": "https://...",
                "similarity": 0.85
            }
        ]
    }
}
```

### 5.2 首页推荐

```
GET /api/v1/rec/home

Query:
    page = 1
    page_size = 20

Response:
{
    "code": 0,
    "data": {
        "items": [
            {
                "sku_id": 12345,
                "name": "商品名称",
                "price": 99.00,
                "original_price": 199.00,
                "image": "https://...",
                "reason": "为你推荐"
            }
        ],
        "page": 1,
        "page_size": 20,
        "total": 100,
        "source": "personalized"  // personalized | category | global
    }
}
```

### 5.3 冷启动接口

```
GET /api/v1/rec/cold-start

Response:
{
    "code": 0,
    "data": {
        "hot_items": [...],        // 全站热卖
        "category_prefs": [        // 类目选择
            { "category_id": 1, "name": "手机", "image": "..." },
            { "category_id": 2, "name": "服装", "image": "..." }
        ]
    }
}
```

### 5.4 类目偏好设置

```
POST /api/v1/rec/preference

Request:
{
    "category_ids": [1, 2, 3]  // 用户选择的类目
}

Response:
{
    "code": 0,
    "message": "设置成功"
}
```

---

## 六、服务配置

### 6.1 推荐服务结构

```
internal/
  app/
    recommendation/          # 推荐服务入口
      run.go
  recommendation/
    handler.go               # HTTP handlers
    service.go               # 业务逻辑
    consumer.go              # MQ消费者
    algorithm/
      item_cf.go             # Item-CF算法
      user_cf.go             # User-CF算法
      bestseller.go           # 热卖计算
    model.go                 # 数据模型
    cache.go                 # Redis缓存
```

### 6.2 数据库配置

```yaml
# 推荐服务配置
Recommendation:
  DB:
    Master: "root:password@tcp(localhost:3306)/recommendation?charset=utf8mb4"
    Slave: "root:password@tcp(localhost:3307)/recommendation?charset=utf8mb4"  # 读写分离

  Redis:
    Addr: "localhost:6379"
    DB: 1

  RabbitMQ:
    URL: "amqp://guest:guest@localhost:5672/"
    Queue: "user_behavior_topic"
    PrefetchCount: 10

  Algorithm:
    ItemCF:
      MinCoUsers: 2          # 最小共同用户数
      TopN: 50               # 每个商品保留Top50相似商品
    UserCF:
      TopK: 20               # 相似用户数
      MinBehaviors: 10       # 最小行为数才用UserCF
    Bestseller:
      PeriodDays: 30         # 统计周期
      TopN: 100             # 每个类目保留Top100
```

---

## 七、定时任务

| 任务 | 频率 | 说明 |
|------|------|------|
| 计算商品相似度 | 每日 02:00 | Item-CF离线计算 |
| 计算类目热卖 | 每日 03:00 | 更新热卖榜 |
| 计算用户偏好 | 每日 04:00 | 批量更新用户偏好 |
| 清理过期数据 | 每周一 02:00 | 清理90天外行为日志 |
| 补全商品信息 | 每小时 | 从商品服务同步商品信息到缓存 |

---

## 八、监控指标

### 8.1 业务指标

| 指标 | 说明 | 报警阈值 |
|------|------|----------|
| behavior_report_qps | 行为上报QPS | < 100 |
| mq_consume_lag | 消息消费延迟 | > 1000 |
| rec_click_rate | 推荐点击率 | < 1% |
| rec_coverage | 推荐覆盖率 | < 30% |

### 8.2 技术指标

| 指标 | 说明 | 报警阈值 |
|------|------|----------|
| api_latency_p99 | API延迟P99 | > 200ms |
| mq_dlq_count | 死信队列消息数 | > 10 |
| db_connection_errors | 数据库连接错误 | > 0 |
| cache_hit_rate | 缓存命中率 | < 80% |

---

## 九、MVP交付范围

### 9.1 必须交付

| 模块 | 功能 | 验收标准 |
|------|------|----------|
| 行为采集 | cart/favorite/purchase上报 | 数据稳定落库，无丢失 |
| Item-CF | 商品相似度计算 | 离线任务正常，结果合理 |
| 看了又看API | 返回相似商品 | P99 < 100ms |
| 首页推荐API | 个性化+热卖兜底 | 冷启动用户能看到热卖 |
| 类目偏好 | 用户选择类目 | 设置后下次生效 |

### 9.2 暂不交付

- 首页User-CF个性化（数据量不足时）
- 推荐效果评估系统
- A/B测试框架
- 实时推荐

---

## 十、数据可靠性保证

### 10.1 写入可靠性

1. **MQ持久化**：durable=true, delivery_mode=2
2. **消费者ACK**：手动ACK，处理成功后确认
3. **失败重试**：指数退避 1s → 5s → 15s
4. **死信队列**：超过重试次数进入DLQ，触发告警

### 10.2 数据质量

1. **去重**：5分钟内同用户同商品同行为只记1条
2. **行为权重**：不同行为类型有不同权重
3. **时间衰减**：近期行为权重更高
4. **数据校验**：behavior_type、sku_id合法性校验

### 10.3 异常处理

| 异常场景 | 处理方式 |
|----------|----------|
| RabbitMQ不可用 | 前端降级为同步调用，失败重试 |
| MySQL主库不可用 | 写入本地文件，恢复后补录 |
| Redis不可用 | 降级为直接查DB |
| 计算任务失败 | 发送告警，保留上次计算结果 |