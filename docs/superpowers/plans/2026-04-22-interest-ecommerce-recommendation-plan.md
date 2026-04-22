# 兴趣电商用户行为追踪与推荐系统实施计划

## 概述

本文档为兴趣电商推荐系统的实施计划，基于 `2026-04-22-interest-ecommerce-recommendation-design.md` 设计文档。

---

## Phase 1: 行为追踪体系 (Week 1-2)

### 1.1 数据模型设计

**目标**: 定义统一的行为事件模型，支持全链路追踪

**任务清单**:

| # | 任务 | 工作内容 | 交付物 |
|---|-----|---------|-------|
| 1.1.1 | 设计 Protobuf 事件模型 | 参考设计文档 2.1，定义 `UserBehaviorEvent` 结构 | `proto/behavior.proto` |
| 1.1.2 | 设计 ClickHouse 表结构 | 行为数据湖表、分区设计、物化视图 | `sql/behavior_ddl.sql` |
| 1.1.3 | 设计 Redis 数据结构 | 实时特征、曝光过滤、黑名单 | `docs/redis_schema.md` |

**proto/behavior.proto**:
```protobuf
syntax = "proto3";
package behavior;

message UserBehaviorEvent {
  string event_id = 1;
  string user_id = 2;
  string device_id = 3;
  int64 timestamp = 4;
  string platform = 5;
  string event_type = 6;
  string item_id = 7;
  string category_id = 8;
  map<string, string> properties = 9;
  int32 position = 10;
  string shop_id = 11;
  int64 price = 12;
}

message BatchEventRequest {
  repeated UserBehaviorEvent events = 1;
  string session_id = 2;
}
```

### 1.2 埋点 SDK 开发

**目标**: 前端 SDK 自动采集用户行为

**任务清单**:

| # | 任务 | 工作内容 | 交付物 |
|---|-----|---------|-------|
| 1.2.1 | 基础追踪模块 | `BehaviorTracker` 核心类，事件队列、批量发送 | `user-mall/src/sdk/BehaviorTracker.ts` |
| 1.2.2 | 自动曝光采集 | `IntersectionObserver` 实现商品曝光 | 曝光追踪模块 |
| 1.2.3 | 页面停留时长 | `visibilitychange` 监听 | 页面分析模块 |
| 1.2.4 | 搜索/加购追踪 | 电商特有行为封装 | `trackSearch`, `trackConversion` |
| 1.2.5 | React Hooks 封装 | `useTracker`, `useExposure` | `user-mall/src/hooks/useTracker.ts` |

**user-mall/src/hooks/useTracker.ts**:
```typescript
import { useEffect, useRef } from 'react'
import { tracker } from '@/sdk/BehaviorTracker'

export function useTracker() {
  return {
    track: tracker.track.bind(tracker),
    trackExposure: tracker.trackExposure.bind(tracker),
    trackSearch: tracker.trackSearch.bind(tracker),
    trackConversion: tracker.trackConversion.bind(tracker),
  }
}

export function useExposure(
  itemId: string,
  position: number,
  elementRef: React.RefObject<HTMLElement>
) {
  const { trackExposure } = useTracker()

  useEffect(() => {
    if (elementRef.current) {
      trackExposure(itemId, position, elementRef.current)
    }
  }, [itemId, position])
}
```

### 1.3 后端事件采集

**目标**: 高性能事件接收和写入 Kafka

**任务清单**:

| # | 任务 | 工作内容 | 交付物 |
|---|-----|---------|-------|
| 1.3.1 | HTTP 接收服务 | `/api/behavior/track` 批量接收 | `internal/behavior/api` |
| 1.3.2 | Kafka Producer | 事件写入 Kafka，消息压缩 | `internal/behavior/kafka` |
| 1.3.3 | 数据校验 | Protobuf 解析、字段校验 | 校验中间件 |
| 1.3.4 | 降级策略 | Redis 缓冲、指数退避重试 | 容错机制 |

**internal/behavior/api/handler.go**:
```go
type Handler struct {
  producer *kafka.Producer
  redis    *redis.Client
}

func (h *Handler) TrackEvents(c *gin.Context) {
  var req proto.BatchEventRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
  }

  // 异步写入 Kafka
  go h.sendToKafka(&req)

  c.JSON(200, gin.H{"code": 0})
}

func (h *Handler) sendToKafka(req *proto.BatchEventRequest) {
  for _, event := range req.Events {
    data, _ := proto.Marshal(event)
    h.producer.Send(&kafka.Message{
      Topic: "user-behavior-topic",
      Key:   []byte(event.UserId),
      Value: data,
    })
  }
}
```

---

## Phase 2: 用户画像 (Week 3-4)

### 2.1 离线画像计算

**目标**: 基于 Hive/Spark 的批量用户画像

**任务清单**:

| # | 任务 | 工作内容 | 交付物 |
|---|-----|---------|-------|
| 2.1.1 | 画像表设计 | MySQL 用户画像表 | `sql/user_profile_ddl.sql` |
| 2.1.2 | 统计特征计算 | 累计订单/金额/类目偏好 | `jobs/spark/user_stats_job` |
| 2.1.3 | 偏好类目计算 | Top-K 类目标签 | `jobs/spark/category_preference_job` |
| 2.1.4 | ETL 调度 | 每日增量更新 | `jobs/scheduler/profile_update` |

### 2.2 实时特征计算

**目标**: Flink 实时更新用户统计特征

**任务清单**:

| # | 任务 | 工作内容 | 交付物 |
|---|-----|---------|-------|
| 2.2.1 | Flink 任务开发 | 会话窗口聚合 | `jobs/flink/user_realtime_stats` |
| 2.2.2 | Redis Sink | 特征写入 Redis | Redis 写入器 |
| 2.2.3 | 画像服务 API | 特征读取接口 | `internal/feature/service` |

**internal/feature/service/user_feature.go**:
```go
type UserFeatureService struct {
  redis *redis.Client
}

type UserFeatures struct {
  ClickCount5m  int   `json:"click_count_5m"`
  CartCount5m  int   `json:"cart_count_5m"`
  OrderCount5m int   `json:"order_count_5m"`
  LastActiveTime int64 `json:"last_active_time"`
}

func (s *UserFeatureService) GetUserFeatures(userID string) (*UserFeatures, error) {
  key := fmt.Sprintf("rt:user:%s", userID)
  data, err := s.redis.HgetAll(key).Result()
  if err != nil {
    return nil, err
  }

  return &UserFeatures{
    ClickCount5m:   parseInt(data["click_count_5m"]),
    CartCount5m:    parseInt(data["cart_count_5m"]),
    OrderCount5m:   parseInt(data["order_count_5m"]),
    LastActiveTime: parseInt64(data["last_active_time"]),
  }, nil
}
```

---

## Phase 3: 推荐系统 (Week 5-8)

### 3.1 召回服务

**目标**: 多路召回，支持 1000+ 候选集

**任务清单**:

| # | 任务 | 工作内容 | 交付物 |
|---|-----|---------|-------|
| 3.1.1 | 协同过滤召回 | i2i/u2i 召回 | `internal/recommend/recall/cf_recall` |
| 3.1.2 | 商品相似召回 | 向量召回 (Milvus) | `internal/recommend/recall/embedding_recall` |
| 3.1.3 | 实时行为召回 | 基于最近交互 | `internal/recommend/recall/realtime_recall` |
| 3.1.4 | 热门/新品召回 | Hot/New 通道 | `internal/recommend/recall/hot_new_recall` |
| 3.1.5 | 多路召回合并 | 分数归一化、权重分配 | `internal/recommend/recall/multi_recall` |

### 3.2 排序模型

**目标**: 精排模型预测转化概率

**任务清单**:

| # | 任务 | 工作内容 | 交付物 |
|---|-----|---------|-------|
| 3.2.1 | 特征工程 | 用户/商品/交叉特征 | `internal/recommend/feature` |
| 3.2.2 | 粗排模型 | LightGBM, 特征简化 | `model/lgbm_rank.py` |
| 3.2.3 | 精排模型 | DeepFM/DIEN | `model/deepfm_rank.py` |
| 3.2.4 | 模型服务 | TorchScript/TFServing | `internal/recommend/model_service` |

### 3.3 重排服务

**目标**: 多样性保证 + 业务规则

**任务清单**:

| # | 任务 | 工作内容 | 交付物 |
|---|-----|---------|-------|
| 3.3.1 | MMR 多样性 | 类目打散 | `internal/recommend/rerank/mmr` |
| 3.3.2 | 位置偏差校正 | Debias 策略 | `internal/recommend/rerank/debias` |
| 3.3.3 | 业务规则 | 黑名单/强插 | `internal/recommend/rerank/business_rules` |

**internal/recommend/rerank/reranker.go**:
```go
type Reranker struct {
  blacklist    map[string]bool
  boostRules   []BoostRule
}

func (r *Reranker) Rerank(items []RerankItem, ctx *Context) []RerankItem {
  // 1. 黑名单过滤
  items = r.filterBlacklist(items)

  // 2. MMR 多样性
  items = r.applyMMR(items, ctx)

  // 3. 业务规则
  items = r.applyBoostRules(items, ctx)

  return items
}
```

### 3.4 推荐 API

**目标**: 对外提供推荐接口

**任务清单**:

| # | 任务 | 工作内容 | 交付物 |
|---|-----|---------|-------|
| 3.4.1 | 首页推荐 | `/api/recommend/home` | `internal/recommend/api/home` |
| 3.4.2 | 购物车推荐 | `/api/recommend/cart` | `internal/recommend/api/cart` |
| 3.4.3 | 商品详情页推荐 | `/api/recommend/item/:id` | `internal/recommend/api/item` |
| 3.4.4 | 推荐结果缓存 | Redis 缓存 5 分钟 | 缓存层 |

---

## Phase 4: 前端集成 (Week 7-8)

### 4.1 推荐组件开发

**目标**: 封装推荐 API 为可复用组件

**任务清单**:

| # | 任务 | 工作内容 | 交付物 |
|---|-----|---------|-------|
| 4.1.1 | 推荐列表组件 | `<RecommendFeed>` | `user-mall/src/components/RecommendFeed` |
| 4.1.2 | 猜你喜欢组件 | `<YouMayAlsoLike>` | `user-mall/src/components/YouMayAlsoLike` |
| 4.1.3 | 商品卡片曝光追踪 | 懒加载 + IntersectionObserver | `user-mall/src/components/ProductCard` |
| 4.1.4 | 埋点 Hook | `useProductTracker` | `user-mall/src/hooks/useProductTracker` |

**user-mall/src/components/RecommendFeed/index.tsx**:
```typescript
import { useEffect, useState } from 'react'
import { getRecommend } from '@/api/recommend'
import ProductCard from '@/components/ProductCard'
import { useTracker } from '@/hooks/useTracker'

interface Props {
  scene: 'home' | 'cart' | 'detail'
  itemId?: string
  title?: string
}

export function RecommendFeed({ scene, itemId, title }: Props) {
  const [products, setProducts] = useState([])
  const [loading, setLoading] = useState(true)
  const tracker = useTracker()

  useEffect(() => {
    loadRecommend()
  }, [scene, itemId])

  const loadRecommend = async () => {
    try {
      const res = await getRecommend({ scene, item_id: itemId })
      setProducts(res.items)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  if (loading) return <Skeleton />

  return (
    <div className="recommend-feed">
      {title && <div className="feed-title">{title}</div>}
      <div className="feed-list columns-2 gap-2 px-2">
        {products.map((product, index) => (
          <div key={product.sku_id} className="mb-2 break-inside-avoid">
            <ProductCard
              product={product}
              viewMode="waterfall"
              onExposure={() => tracker.trackExposure(product.sku_id, index)}
              onClick={() => tracker.trackClick(product.sku_id)}
            />
          </div>
        ))}
      </div>
    </div>
  )
}
```

### 4.2 页面集成

**目标**: 各页面接入推荐

| 页面 | 推荐场景 | 接入方式 |
|-----|---------|---------|
| 首页 | 猜你喜欢 Feed | `<RecommendFeed scene="home" />` |
| 商品详情页 |看了又看 | `<RecommendFeed scene="detail" itemId={id} />` |
| 购物车 | 追加推荐 | `<RecommendFeed scene="cart" />` |
| 支付成功页 | 为您推荐 | `<RecommendFeed scene="order" />` |

---

## Phase 5: 测试与优化 (Week 9-10)

### 5.1 功能测试

| # | 测试项 | 验证标准 |
|---|-------|---------|
| 1 | 行为埋点 | 事件正确写入 Kafka |
| 2 | 画像更新 | 实时特征 5 分钟内更新 |
| 3 | 推荐接口 | P99 < 100ms |
| 4 | 前端曝光 | 商品可见 50% 触发曝光 |

### 5.2 性能测试

| # | 指标 | 目标 |
|---|-----|------|
| 1 | 推荐接口 QPS | > 5000 |
| 2 | 推荐接口延迟 P99 | < 100ms |
| 3 | 特征读取延迟 P99 | < 5ms |
| 4 | 系统可用性 | > 99.9% |

### 5.3 ABTest 设计

```sql
-- ABTest 实验配置表
CREATE TABLE abtest_experiments (
  experiment_id VARCHAR(64) PRIMARY KEY,
  experiment_name VARCHAR(128),
  traffic_ratio DECIMAL(5,4),  -- 流量比例 0.0000-1.0000
  param_json JSON,              -- 实验参数
  status TINYINT,              -- 0草稿 1运行 2结束
  start_time DATETIME,
  end_time DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 用户实验分流
CREATE TABLE abtest_user_bucket (
  user_id BIGINT PRIMARY KEY,
  bucket_id INT,               -- 分桶号 0-99
  updated_at DATETIME
);
```

---

## Milestone 里程碑

| 阶段 | 完成日期 | 交付内容 |
|-----|---------|---------|
| M1 | Week 2 | 行为追踪 SDK + 事件采集 |
| M2 | Week 4 | 用户画像 + 特征服务 |
| M3 | Week 6 | 召回 + 粗排服务 |
| M4 | Week 8 | 精排 + 重排 + 推荐 API |
| M5 | Week 10 | 前端集成 + ABTest 上线 |

---

## 风险与对策

| 风险 | 影响 | 对策 |
|-----|------|-----|
| 数据延迟 | 推荐不准确 | 增加实时特征权重，离线特征兜底 |
| 冷启动 | 新用户无数据 | 热门/新品召回，类目平均偏好 |
| 计算瓶颈 | 推荐延迟高 | 缓存优化，模型蒸馏降级 |
| 隐私合规 | 数据合规风险 | 匿名化处理，用户授权确认 |