# 兴趣电商用户行为追踪与推荐系统设计

> 基于大厂成熟实践（字节/抖音电商、淘宝猜你喜欢）结合兴趣电商特性

## 1. 概述

### 1.1 背景

兴趣电商以"发现式购物"为核心，通过算法推荐激发用户潜在消费需求。与传统搜索电商不同，用户往往没有明确购物意图，需要通过精细化的用户行为追踪和实时推荐来实现精准匹配。

### 1.2 设计目标

- **行为采集**：全链路、实时、低侵入式用户行为追踪
- **用户理解**：多维度用户画像构建，实时更新
- **推荐系统**：毫秒级响应，支持千亿级特征规模
- **业务适配**：深度整合"秒杀/优惠券/直播"等兴趣电商核心场景

### 1.3 参考架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         用户端 SDK                                │
│  (Web/App/小程序 行为埋点自动采集)                                 │
└─────────────────────────┬───────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Kafka 消息队列                               │
│  (用户行为事件实时写入，支持百万级 TPS)                            │
└─────────────────────────┬───────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Flink 实时计算层                              │
│  - 实时会话切分 (Session)                                        │
│  - 实时特征更新 (Feature Engineering)                            │
│  - 实时指标计算 (UV/PV/转化率)                                    │
└───────────┬─────────────────────────────┬───────────────────────┘
            ▼                             ▼
┌───────────────────────┐   ┌───────────────────────────────┐
│     Redis Cluster     │   │       ClickHouse / Doris      │
│   (在线特征存储)       │   │      (用户行为数据湖 OLAP)      │
│   毫秒级特征读取       │   │      历史分析 & ABTest         │
└───────────┬───────────┘   └───────────────┬───────────────┘
            │                               │
            ▼                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      推荐服务 (Rank Service)                      │
│  - 召回层 (Match): i2i/u2i/u2u2i 多路召回                        │
│  - 粗排层 (Pre-Rank): 轻量级 ML 模型                            │
│  - 精排层 (Rank): DeepFM/DIEN 等复杂模型                         │
│  - 重排层 (Rerank): 上下文感知 & 业务规则                        │
└─────────────────────────┬───────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                        模型训练平台                               │
│  - 样本拼接 (Label + Feature)                                    │
│  - 离线训练 (TensorFlow/PyTorch)                                 │
│  - 在线学习 (Online Learning)                                    │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. 用户行为追踪体系

### 2.1 行为事件模型

参考字节跳动 `LogAgent` + 淘宝用户行为数据模型，设计统一事件格式：

```protobuf
message UserBehaviorEvent {
  // 基础字段
  string event_id = 1;           // 事件唯一ID (UUID)
  string user_id = 2;            // 用户ID (未登录可为空)
  string device_id = 3;          // 设备ID (匿名态追踪)
  int64 timestamp = 4;           // 事件发生时间 (毫秒)
  string platform = 5;           // 平台: app/web/h5/miniapp
  string os = 6;                 // 操作系统
  string app_version = 7;        // App版本

  // 行为上下文
  string scene = 8;              // 场景: home_feed/search/recommend/live
  string page = 9;               // 页面路由
  string referrer_page = 10;     // 来源页面

  // 行为类型
  string event_type = 11;         // 行为类型 (见2.2)
  string item_id = 12;           // 商品/内容ID
  string item_category = 13;     // 商品类目
  repeated string item_tags = 14; // 商品标签

  // 行为属性
  map<string, string> properties = 15;  // 行为附加属性
  int32 duration = 16;          // 停留时长 (毫秒)
  string search_keyword = 17;    // 搜索关键词

  // 曝光信息
  int32 position = 18;          // 推荐位排序
  bool is_clicked = 19;         // 是否点击
  bool is_converted = 20;       // 是否转化 (加购/下单)

  // 电商特有
  string shop_id = 21;          // 店铺ID
  int64 price = 22;             // 商品价格 (分)
  int32 quantity = 23;          // 购买数量
  string order_id = 24;         // 订单ID
}
```

### 2.2 核心行为类型

| 行为类型 | event_type | 触发时机 | 推荐权重 |
|---------|------------|---------|---------|
| 商品曝光 | `item_exposure` | 商品在推荐位展示>100ms | 更新置信度 |
| 商品点击 | `item_click` | 点击商品卡片 | 正样本+1 |
| 商品详情页 | `item_detail` | 进入商品详情页 | 深度行为 |
| 收藏 | `item_collect` | 点击收藏按钮 | 强正信号 |
| 加购 | `item_cart` | 加入购物车 | 强正信号 |
| 下单 | `item_order` | 提交订单 | 转化信号 |
| 支付 | `item_pay` | 完成支付 | 强转化 |
| 搜索 | `search` | 发起搜索 | 意图信号 |
| 取消/退款 | `item_refund` | 取消/退款 | 负信号 |
| 点赞 | `item_like` | 点赞商品 | 强正信号 |
| 分享 | `item_share` | 分享商品 | 强正信号 |
| 直播互动 | `live_interact` | 评论/送礼/关注 | 深度行为 |
| 短视频互动 | `video_interact` | 完播/点赞/评论 | 内容理解 |

### 2.3 兴趣电商特有行为

```protobuf
// 直播场景扩展
message LiveBehavior {
  string live_id = 1;           // 直播间ID
  string anchor_id = 2;         // 主播ID
  int64 watch_duration = 3;    // 观看时长
  bool is_followed = 4;         // 是否关注
  string gift_type = 5;          // 送礼类型
  string comment_content = 6;    // 评论内容 (脱敏)
}

// 短视频场景扩展
message VideoBehavior {
  string video_id = 1;          // 视频ID
  int64 play_duration = 2;      // 播放时长
  int64 video_duration = 3;     // 视频总时长
  bool is_full_play = 4;        // 是否完整播放
  string interaction_type = 5;  // 互动类型
}
```

### 2.4 采集 SDK 设计

参考字节 `DataKit` 架构：

```typescript
// user-mall/src/sdk/BehaviorTracker.ts

interface TrackConfig {
  appId: string;
  serverUrl: string;
  sessionTimeout: number;   // 会话超时(ms)
  sampleRate: number;      // 采样率 0-1
  enableDebug: boolean;
}

class BehaviorTracker {
  private config: TrackConfig;
  private eventQueue: UserBehaviorEvent[] = [];
  private sessionId: string;
  private pageMark: string;  // 页面标识

  constructor(config: TrackConfig) {
    this.config = config;
    this.sessionId = this.generateSessionId();
    this.setupAutoTrack();   // 自动采集页面行为
  }

  // 基础行为追踪
  track(eventType: string, properties: Record<string, any> = {}) {
    if (!this.shouldTrack()) return;

    const event = this.buildEvent(eventType, properties);
    this.eventQueue.push(event);
    this.flushIfNeeded();
  }

  // 商品曝光 (需配合 IntersectionObserver)
  trackExposure(itemId: string, position: number, element: HTMLElement) {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting && entry.intersectionRatio > 0.5) {
          this.track('item_exposure', {
            item_id: itemId,
            position: position,
            exposure_time: Date.now(),
            viewport_ratio: entry.intersectionRatio
          });
        }
      });
    }, { threshold: 0.5 });

    observer.observe(element);
  }

  // 页面停留时长
  trackPageStay(pageName: string, referrer?: string) {
    const startTime = Date.now();

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'hidden') {
        const duration = Date.now() - startTime;
        this.track('page_stay', {
          page: pageName,
          referrer: referrer,
          duration: duration
        });
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
  }

  // 搜索行为
  trackSearch(keyword: string, resultCount: number) {
    this.track('search', {
      search_keyword: keyword,
      result_count: resultCount,
      is_empty: resultCount === 0
    });
  }

  // 加购/收藏/下单
  trackConversion(eventType: 'cart' | 'collect' | 'order' | 'pay', item: Product) {
    this.track(`item_${eventType}`, {
      item_id: item.sku_id,
      item_category: item.category_id,
      shop_id: item.shop_id,
      price: item.price,
      tags: item.tags.join(',')
    });
  }

  // 批量发送 (节流)
  private flush() {
    if (this.eventQueue.length === 0) return;

    const batch = this.eventQueue.splice(0, 100);
    this.sendToServer(batch);
  }

  private sendToServer(events: UserBehaviorEvent[]) {
    navigator.sendBeacon?.(this.config.serverUrl, JSON.stringify({
      events,
      session_id: this.sessionId
    }));
  }
}

export const tracker = new BehaviorTracker({
  appId: 'user-mall',
  serverUrl: '/api/behavior/track',
  sessionTimeout: 1800000,  // 30分钟
  sampleRate: 1.0,
  enableDebug: false
});
```

---

## 3. 用户画像体系

### 3.1 画像结构设计

参考淘宝猜你喜欢 `TMM` (Three-Stage Million Million) 画像系统：

```sql
-- 用户基础画像
CREATE TABLE user_profile (
  user_id          BIGINT PRIMARY KEY,
  device_id        VARCHAR(64),           -- 匿名设备ID

  -- 基础属性
  gender          TINYINT,                -- 性别 0未知 1男 2女
  age_range       TINYINT,                -- 年龄段
  city_level      TINYINT,                -- 城市等级
  register_time    DATETIME,              -- 注册时间
  member_level    TINYINT,                -- 会员等级

  -- 统计特征 (实时更新)
  total_orders    INT,                    -- 累计订单数
  total_amount    BIGINT,                 -- 累计消费金额(分)
  avg_order_amount BIGINT,                -- 平均订单金额
  cart_count      INT,                    -- 购物车商品数
  favorite_count  INT,                    -- 收藏数

  -- 偏好类目 (Top 20, JSON数组)
  prefer_categories   JSON,               -- ["C001:0.85", "C002:0.72", ...]
  prefer_brands       JSON,               -- 偏好品牌
  prefer_price_ranges JSON,               -- 偏好价格带

  -- 行为特征
  active_days_30   INT,                    -- 近30天活跃天数
  last_active_time DATETIME,              -- 最后活跃时间
  last_browse_time DATETIME,             -- 最后浏览时间
  last_order_time  DATETIME,              -- 最后下单时间

  -- 兴趣电商特有
  live_watch_time  BIGINT,                -- 累计直播观看时长(秒)
  video_likes     INT,                    -- 短视频点赞数
  follow_anchors  JSON,                   -- 关注的主播列表

  -- 实时特征 (每5分钟更新)
  realtime_features JSON,                 -- 实时特征快照

  -- 向量特征 (用于向量检索)
  user_vector_id   BIGINT,                -- Milvus向量ID

  updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- 用户行为序列 (用于序列模型)
CREATE TABLE user_behavior_sequence (
  id              BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id         BIGINT NOT NULL,
  event_type      VARCHAR(32) NOT NULL,   -- browse/click/cart/order
  item_id         VARCHAR(64) NOT NULL,
  category_id     VARCHAR(32),
  shop_id         VARCHAR(64),
  price           INT,
  timestamp       DATETIME NOT NULL,

  INDEX idx_user_time (user_id, timestamp DESC),
  INDEX idx_user_behavior (user_id, event_type)
) ENGINE=InnoDB;
```

### 3.2 实时特征计算 (Flink)

```scala
// 用户实时统计特征计算
object UserRealtimeStats {

  def processUserBehavior(): Table = {
    env.fromKafkaSource[UserBehaviorEvent]("user-behavior-topic")
      // 按用户ID分组
      .keyBy(_.user_id)
      // 滑动窗口计算
      .window(SlidingEventTimeWindows.of(Time.minutes(5), Time.minutes(1)))
      .aggregate(new UserStatsAggregator)
  }
}

class UserStatsAggregator extends AggregateFunction[UserBehaviorEvent, UserStats, UserStats] {

  override def createAccumulator(): UserStats = UserStats()

  override def add(value: UserBehaviorEvent, accumulator: UserStats): UserStats = {
    value.event_type match {
      case "item_click" =>
        accumulator.click_count += 1
      case "item_cart" =>
        accumulator.cart_count += 1
      case "item_order" | "item_pay" =>
        accumulator.order_count += 1
        accumulator.order_amount += value.price * value.quantity
    }
    accumulator.last_active_time = value.timestamp
    accumulator
  }

  override def getResult(accumulator: UserStats): UserStats = accumulator

  override def merge(a: UserStats, b: UserStats): UserStats = {
    a.merge(b)
  }
}

// 5分钟窗口结果写入 Redis
class RedisSink extends SinkFunction[UserStats] {
  override def invoke(value: UserStats): Unit = {
    val key = s"user:stats:${value.user_id}"
    val hash = Map(
      "click_count_5m" -> value.click_count,
      "cart_count_5m" -> value.cart_count,
      "order_count_5m" -> value.order_count,
      "last_active_time" -> value.last_active_time
    )
    redis.hset(key, hash)
    redis.expire(key, 600)  // 10分钟过期
  }
}
```

---

## 4. 推荐系统架构

### 4.1 多阶段推荐流程

参考抖音电商推荐架构：

```
┌─────────────────────────────────────────────────────────────────┐
│                        推荐请求入口                              │
│              (首页Feed/商品详情页/购物车/搜索结果)                 │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     召回层 (Match) ~1000 items                   │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │  协同过滤    │ │  深度网络   │ │  商品相似   │ │  实时召回   │ │
│  │  (i2i/u2i) │ │  (Graph)   │ │  (i2i)     │ │  (Hot/New) │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │  搜索召回   │ │  直播召回   │ │  短视频召回 │ │  地域/LBS   │ │
│  │            │ │            │ │            │ │            │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     粗排层 (Pre-Rank) ~300 items                │
│           LightGBM 模型, 特征量少, 时延<5ms                       │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     精排层 (Rank) ~100 items                    │
│         DeepFM/DIEN 模型, 深度特征, 时延<20ms                    │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     重排层 (Rerank) ~20 items                   │
│  - 多样性保证 (MMR)                                             │
│  - 上下文感知 (Context-aware)                                    │
│  - 业务规则 (黑名单/强插/必返)                                   │
│  - 打散策略 (类目/品牌/价格带)                                   │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     推荐结果返回                                 │
│              (商品列表 + 推荐理由 + 解释性文本)                    │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 召回通道设计

```go
// internal/recommend/recall/multi_recall.go

type RecallResult struct {
  Items   []RecallItem
  Channel string
  Score   float64
}

type RecallItem struct {
  ItemID       string
  Score       float64
  Reason      string  // 召回原因 (用于推荐理由)
  Channel     string  // 召回通道
}

// 多路召回服务
type MultiRecall struct {
  // u2i 召回 (用户 -> 商品)
  userCFRecall     *UserCFRecall
  userEmbeddingRecall *UserEmbeddingRecall

  // i2i 召回 (商品 -> 商品)
  itemCFRecall     *ItemCFRecall
  itemGraphRecall  *ItemGraphRecall

  // 实时行为召回
  realtimeRecall   *RealtimeRecall

  // 热门/新品召回
  hotRecall        *HotRecall
  newRecall        *NewRecall

  // 搜索召回
  searchRecall     *SearchRecall

  // 直播/内容召回
  liveRecall       *LiveRecall
}

func (m *MultiRecall) Recall(ctx *RecallRequest) ([]RecallResult, error) {
  var wg sync.WaitGroup
  ch := make(chan []RecallResult, 8)

  // 并行执行各召回通道
  wg.Add(8)

  go func() { defer wg.Done(); ch <- m.userCFRecall.Recall(ctx) }()
  go func() { defer wg.Done(); ch <- m.userEmbeddingRecall.Recall(ctx) }()
  go func() { defer wg.Done(); ch <- m.itemCFRecall.Recall(ctx) }()
  go func() { defer wg.Done(); ch <- m.itemGraphRecall.Recall(ctx) }()
  go func() { defer wg.Done(); ch <- m.realtimeRecall.Recall(ctx) }()
  go func() { defer wg.Done(); ch <- m.hotRecall.Recall(ctx) }()
  go func() { defer wg.Done(); ch <- m.newRecall.Recall(ctx) }()
  go func() { defer wg.Done(); ch <- m.searchRecall.Recall(ctx) }()

  wg.Wait()
  close(ch)

  // 合并结果
  return m.mergeResults(ch), nil
}

// 实时行为召回 (基于用户最近交互商品)
type RealtimeRecall struct {
  redis *redis.Client
}

func (r *RealtimeRecall) Recall(ctx *RecallRequest) []RecallResult {
  // 获取用户最近交互的N个商品
  recentItems, _ := r.redis.Lrange(fmt.Sprintf("user:recent:%s", ctx.UserID), 0, 9).Result()

  var results []RecallResult
  for _, itemID := range recentItems {
    // 查找相似商品
    similarItems, _ := r.getSimilarItems(itemID, 20)
    for _, sim := range similarItems {
      results = append(results, RecallResult{
        Items:   []RecallItem{sim},
        Channel: "realtime",
        Score:   sim.Score * 0.9,  // 实时行为加权
      })
    }
  }

  return results
}
```

### 4.3 精排模型设计

参考抖音/快手精排模型 `SIM` (Supplement Index Model) + `DIEN`:

```python
# internal/recommend/model/rank_model.py

import torch
import torch.nn as nn

class InterestECommerceRankModel(nn.Module):
    """
    兴趣电商精排模型
    融合: DeepFM + DIEN(兴趣进化) + 电商场景特征
    """

    def __init__(self, config):
        super().__init__()
        self.config = config

        # 1. DeepFM 部分 (低阶+高阶特征交叉)
        self.fm = FMInteraction()
        self.dnn = DNN(config.embed_dim, config.hidden_units)

        # 2. DIEN 兴趣进化层
        self.interest_extractor = InterestExtractor(
            embed_dim=config.embed_dim,
            hidden_dim=config.hidden_dim
        )
        self.interest_evolution = InterestEvolution(
            hidden_dim=config.hidden_dim,
            attention_dim=config.attention_dim
        )

        # 3. 电商场景特有特征
        self.price_encoder = nn.Linear(1, config.embed_dim)
        self.category_encoder = nn.Embedding(config.category_num, config.embed_dim)
        self.shop_encoder = nn.Embedding(config.shop_num, config.embed_dim)

        # 4. 输出层
        self.output_layer = nn.Sequential(
            nn.Linear(config.embed_dim * 4 + config.hidden_dim, 256),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(256, 64),
            nn.ReLU(),
            nn.Linear(64, 1),
            nn.Sigmoid()
        )

    def forward(self, user_features, item_features, behavior_sequence):
        """
        Args:
            user_features: 用户特征 [batch, user_feat_dim]
            item_features: 候选商品特征 [batch, item_feat_dim]
            behavior_sequence: 用户行为序列 [batch, seq_len, embed_dim]
        """

        # 1. DeepFM
        fm_out = self.fm(user_features, item_features)
        dnn_out = self.dnn(torch.cat([user_features, item_features], dim=1))

        # 2. DIEN 兴趣提取
        interest_extracted = self.interest_extractor(behavior_sequence)
        interest_evolved = self.interest_evolution(
            interest_extracted,
            item_features  # 用于attention target
        )

        # 3. 电商特征编码
        price_feat = self.price_encoder(item_features[:, :1])
        cat_feat = self.category_encoder(item_features[:, 1])
        shop_feat = self.shop_encoder(item_features[:, 2])

        # 4. 融合
        concat_feat = torch.cat([
            fm_out, dnn_out, interest_evolved,
            price_feat, cat_feat, shop_feat
        ], dim=1)

        # 5. 预测
        output = self.output_layer(concat_feat)
        return output.squeeze(-1)


class InterestExtractor(nn.Module):
    """兴趣提取层 - 模拟DIEN"""

    def __init__(self, embed_dim, hidden_dim):
        super().__init__()
        self.gru = nn.GRU(embed_dim, hidden_dim, batch_first=True)

    def forward(self, behavior_sequence):
        # behavior_sequence: [batch, seq_len, embed_dim]
        output, _ = self.gru(behavior_sequence)
        return output  # 返回每个时刻的兴趣表示


class InterestEvolution(nn.Module):
    """兴趣进化层 - 注意力机制"""

    def __init__(self, hidden_dim, attention_dim):
        super().__init__()
        self.attention = nn.Sequential(
            nn.Linear(hidden_dim * 2, attention_dim),
            nn.ReLU(),
            nn.Linear(attention_dim, 1),
            nn.Softmax(dim=1)
        )
        self.gru = nn.GRU(hidden_dim, hidden_dim, batch_first=True)

    def forward(self, interests, target):
        # interests: [batch, seq_len, hidden_dim]
        # target: [batch, embed_dim]

        # 构造attention输入
        target_expand = target.unsqueeze(1).expand(-1, interests.size(1), -1)
        attention_input = torch.cat([interests, target_expand], dim=2)

        # 计算attention权重
        attention_weight = self.attention(attention_input)

        # 加权求和
        weighted_interests = interests * attention_weight
        return weighted_interests.sum(dim=1)
```

### 4.4 重排策略

```go
// internal/recommend/rerank/diversity_rerank.go

type Reranker struct {
  blacklist  *BlacklistService
  businessRules *BusinessRules
}

type RerankRequest struct {
  Items        []RerankItem
  UserID       string
  Scene        string  // home/search/cart
  Context      *RequestContext
}

type RerankItem struct {
  ItemID       string
  Score       float64
  CategoryID  string
  BrandID     string
  PriceRange  string
  ShopID      string
}

func (r *Reranker) Rerank(req *RerankRequest) []RerankItem {
  items := req.Items

  // 1. 黑名单过滤
  items = r.filterBlacklist(items, req.UserID)

  // 2. 业务规则过滤 (已售罄/下架等)
  items = r.filterBusinessRules(items)

  // 3. 多样性打散 (MMR)
  items = r.applyMMR(items, req.Context)

  // 4. 类目/品牌打散
  items = r.applyCategoryDiversity(items)

  // 5. 强插策略 (新用户必返爆款/老用户个性化)
  items = r.applyBoosting(items, req.UserID)

  // 6. 位置偏差校正 (position bias)
  items = r.debiasPosition(items)

  return items
}

// MMR (Maximal Marginal Relevance) 多样性
func (r *Reranker) applyMMR(items []RerankItem, ctx *RequestContext) []RerankItem {
  const lambda = 0.5  // 多样性权重
  result := make([]RerankItem, 0, len(items))

  for len(result) < 20 && len(items) > 0 {
    // 计算每个候选的 MMR 分数
    bestIdx := 0
    bestScore := -1.0

    for i, item := range items {
      relevance := item.Score
      diversity := r.calcDiversity(item, result)
      mmrScore := lambda * relevance + (1 - lambda) * diversity

      if mmrScore > bestScore {
        bestScore = mmrScore
        bestIdx = i
      }
    }

    result = append(result, items[bestIdx])
    items = append(items[:bestIdx], items[bestIdx+1:]...)
  }

  return result
}

// 类目打散
func (r *Reranker) applyCategoryDiversity(items []RerankItem) []RerankItem {
  const maxSameCategory = 3  // 最多连续3个同类目

  result := make([]RerankItem, 0, len(items))
  categoryCount := make(map[string]int)
  lastCategory := ""
  consecutiveCount := 0

  for _, item := range items {
    if item.CategoryID == lastCategory {
      consecutiveCount++
      if consecutiveCount > maxSameCategory {
        // 找下一个不同类目的商品
        continue
      }
    } else {
      consecutiveCount = 1
      lastCategory = item.CategoryID
    }

    result = append(result, item)
    categoryCount[item.CategoryID]++
  }

  // 填充不足20个的情况
  // ...

  return result
}
```

---

## 5. 兴趣电商核心场景推荐

### 5.1 首页Feed推荐

参考抖音电商 "极致沉浸" 体验：

```go
// 首页推荐策略
type HomeFeedStrategy struct {
  userProfile    *UserProfileService
  contentBalance *ContentBalanceStrategy
}

func (s *HomeFeedStrategy) GenerateRecommend(ctx *Context) (*FeedResponse, error) {
  // 1. 获取用户实时状态
  realtimeState, _ := s.userProfile.GetRealtimeState(ctx.UserID)

  // 2. 动态调整各类型内容比例
  // 带货视频 : 纯内容视频 : 品牌专场 = 根据用户活跃度动态调整
  contentMix := s.calculateContentMix(realtimeState)

  // 3. 执行召回
  candidates, _ := s.multiRecall.Recall(&RecallRequest{
    UserID:       ctx.UserID,
    Scene:        "home_feed",
    Count:        1000,
    UserFeatures: realtimeState,
  })

  // 4. 排序
  ranked, _ := s.rankModel.Rank(candidates, realtimeState)

  // 5. 重排
  reranked, _ := s.reranker.Rerank(&RerankRequest{
    Items:   ranked,
    UserID:  ctx.UserID,
    Scene:   "home_feed",
    Context: ctx,
  })

  // 6. 组装推荐理由
  reasons := s.generateRecommendReasons(reranked, realtimeState)

  return &FeedResponse{
    Items:   reranked[:20],
    Reasons: reasons,
    TraceID: generateTraceID(),
  }, nil
}

// 内容配比计算
func (s *HomeFeedStrategy) calculateContentMix(state *RealtimeState) ContentMix {
  // 高活跃用户 -> 更多发现性内容
  if state.activeDays30 > 20 {
    return ContentMix{
      EcommerceRatio:   0.3,
      ContentRatio:    0.4,
      BrandLiveRatio:  0.3,
    }
  }

  // 低活跃用户 -> 更多确定性内容 (已购/收藏相似)
  if state.activeDays30 < 5 {
    return ContentMix{
      EcommerceRatio:   0.7,
      ContentRatio:    0.1,
      BrandLiveRatio:  0.2,
    }
  }

  // 默认配比
  return ContentMix{
    EcommerceRatio:   0.5,
    ContentRatio:    0.3,
    BrandLiveRatio:  0.2,
  }
}
```

### 5.2 购物车追加推荐

参考淘宝 "猜你喜欢" 购物车页：

```go
// 购物车追加推荐
func (s *CartRecommend) GetRecommend(ctx *Context) ([]RecommendItem, error) {
  cartItems, _ := s.cartService.GetUserCart(ctx.UserID)

  // 1. 分析购物车商品特征
  cartFeatures := s.analyzeCartFeatures(cartItems)

  // 2. 协同召回 (买了又买)
  cfItems, _ := s.collaborativeRecall.Recall(&RecallRequest{
    Items:    cartItems,
    Count:    200,
    Strategy: "bought_also_bought",
  })

  // 3. 搭配召回 (买了A可能需要B)
  bundleItems, _ := s.bundleRecall.Recall(&RecallRequest{
    Items:    cartItems,
    Count:    100,
    Strategy: "frequently_bought_together",
  })

  // 4. 过滤 (已购/已在购物车)
  cfItems = s.filterAlreadyOwned(cfItems, cartItems)
  bundleItems = s.filterAlreadyOwned(bundleItems, cartItems)

  // 5. 合并排序
  allItems := append(cfItems, bundleItems...)
  ranked, _ := s.rankModel.Rank(allItems, cartFeatures)

  // 6. 返回TopN + 推荐理由
  return s.buildRecommendWithReasons(ranked[:10], cartItems), nil
}

// 推荐理由生成
func (s *CartRecommend) buildRecommendWithReasons(items []RankedItem, cartItems []CartItem) []RecommendItem {
  reasons := map[string]string{
    "bought_also_bought":    "购买该商品的用户也买了",
    "frequently_bought":    "与购物车中商品搭配购买更优惠",
    "看了又看":              "根据您的浏览记录推荐",
    "同类商品":              "同类型商品对比",
  }

  result := make([]RecommendItem, len(items))
  for i, item := range items {
    result[i] = RecommendItem{
      Item:   item,
      Reason: reasons[item.RecallChannel],
    }
  }
  return result
}
```

### 5.3 秒杀/直播场景推荐

```go
// 秒杀场景推荐
type SeckillRecommend struct {
  seckillService *SeckillService
  flashSaleRecall *FlashSaleRecall
}

func (s *SeckillRecommend) GetRecommend(ctx *Context) ([]RecommendItem, error) {
  // 1. 获取正在进行/即将开始的秒杀场次
  flashSales, _ := s.seckillService.GetActiveFlashSales(ctx.UserID)

  // 2. 用户偏好匹配
  userPrefs, _ := s.userProfile.GetUserPrefs(ctx.UserID)

  // 3. 库存紧张度感知
  for _, fs := range flashSales {
    stockRatio := fs.RemainingStock / fs.TotalStock
    if stockRatio < 0.1 {
      // 库存紧张，优先推荐
      fs.Priority = 1.5
    }
  }

  // 4. 匹配用户偏好类目
  matchedSales := s.matchUserPreferences(flashSales, userPrefs)

  // 5. 排序 (考虑: 匹配度 * 库存紧张度 * 时间紧迫度)
  scored := s.scoreFlashSales(matchedSales)

  return scored, nil
}
```

---

## 6. 数据存储设计

### 6.1 存储架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        在线存储 (Redis)                          │
│  - 用户实时特征 (5分钟窗口)                                       │
│  - 曝光/点击黑名单                                               │
│  - 在线推荐结果缓存                                              │
│  - Session 会话信息                                              │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        离线存储 (Hive)                           │
│  - 原始行为数据 (Parquet 格式)                                    │
│  - 特征工程中间结果                                              │
│  - 模型训练样本                                                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      分析型存储 (ClickHouse)                      │
│  - 用户行为明细 OLAP 查询                                        │
│  - ABTest 指标分析                                              │
│  - 实时/离线指标对比                                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      向量存储 (Milvus)                           │
│  - 商品向量索引 (i2i 向量检索)                                    │
│  - 用户向量索引 (u2i 向量检索)                                    │
└─────────────────────────────────────────────────────────────────┘
```

### 6.2 Redis 数据结构

```go
// 用户实时特征
Key: "rt:user:{user_id}"
Type: Hash
TTL: 600s (10分钟)
Fields:
  - click_count_5m: int
  - cart_count_5m: int
  - order_count_5m: int
  - last_active_time: timestamp
  - last_browse_items: JSON (最近浏览商品ID列表)

// 用户曝光过滤 (防重复曝光)
Key: "exposure:{user_id}"
Type: SortedSet
Score: timestamp
Member: item_id
TTL: 86400s (24小时)

// 实时推荐缓存
Key: "rec:home:{user_id}"
Type: String (JSON)
TTL: 300s (5分钟)

// 购物车追加推荐缓存
Key: "rec:cart:{user_id}"
Type: String (JSON)
TTL: 600s (10分钟)

// 热门商品 (实时更新)
Key: "hot:items:{date}"
Type: SortedSet
Score: 热度分数
Member: item_id
```

---

## 7. 服务部署架构

### 7.1 推荐服务微服务拆分

```
┌─────────────────────────────────────────────────────────────────┐
│                      API Gateway                                 │
│              (统一入口/鉴权/限流/路由)                            │
└───────────────────────────┬────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    推荐服务 (Go)                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │ 召回服务    │ │ 排序服务    │ │ 重排服务    │               │
│  │ (Match)    │ │ (Rank)      │ │ (Rerank)   │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
│                                                                 │
│  P99 < 50ms, QPS > 10000                                        │
└───────────────────────────┬────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      特征服务 (Go)                                │
│  (在线特征读取/特征计算/特征组装)                                 │
│                                                                 │
│  P99 < 5ms                                                       │
└───────────────────────────┬────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      模型服务 (Python)                           │
│  (TensorFlow Serving / TorchScript)                             │
│                                                                 │
│  P99 < 20ms                                                       │
└─────────────────────────────────────────────────────────────────┘
```

### 7.2 高可用设计

```yaml
# 推荐服务 Kubernetes 部署
apiVersion: apps/v1
kind: Deployment
metadata:
  name: recommend-service
spec:
  replicas: 4
  selector:
    matchLabels:
      app: recommend-service
  template:
    spec:
      containers:
      - name: recommend
        image: registry/recommend-service:v1.0
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
        ports:
        - containerPort: 8080
        env:
        - name: REDIS_HOST
          valueFrom:
            configMapKeyRef:
              name: recommend-config
              key: redis.host
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 3

---
# 特征服务 (多副本 + Redis 本地缓存)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: feature-service
spec:
  replicas: 8
  # HPA 自动扩缩容
  autoscaler:
    minReplicas: 4
    maxReplicas: 20
    targetCPUUtilizationPercentage: 70
```

---

## 8. 实现计划

### 8.1 Phase 1: 基础行为追踪 (2周)

| 任务 | 负责 | 里程碑 |
|-----|------|-------|
| 行为事件数据模型设计 | 后端 | Day 1-3 |
| 埋点SDK开发 | 前端 | Day 4-7 |
| Kafka事件采集 | 后端 | Day 8-10 |
| ClickHouse行为数据湖 | 后端 | Day 11-14 |

### 8.2 Phase 2: 用户画像 (2周)

| 任务 | 负责 | 里程碑 |
|-----|------|-------|
| 离线画像计算 | 后端 | Day 15-18 |
| Flink实时特征 | 后端 | Day 19-22 |
| Redis特征存储 | 后端 | Day 23-25 |
| 画像服务API | 后端 | Day 26-28 |

### 8.3 Phase 3: 推荐系统 (4周)

| 任务 | 负责 | 里程碑 |
|-----|------|-------|
| 召回通道开发 | 后端 | Day 29-35 |
| 粗排模型 | 算法 | Day 36-42 |
| 精排模型 | 算法 | Day 43-49 |
| 重排策略 | 后端 | Day 50-56 |
| 集成测试 | 全员 | Day 57-60 |

---

## 9. 附录

### 9.1 参考资料

- 字节跳动推荐系统架构: TikTok Discovery System
- 淘宝猜你喜欢: SIM (Supplement Index Model)
- 快手推荐架构: Kwai Recommender System
- 美团推荐实践: Meituan Food Delivery Recommendation

### 9.2 术语表

| 术语 | 说明 |
|-----|------|
| i2i | Item-to-Item 商品相似度召回 |
| u2i | User-to-Item 用户-商品匹配召回 |
| MMR | Maximal Marginal Relevance 多样性算法 |
| DIEN | Deep Interest Evolution Network 阿里妈妈 |
| DeepFM | 深度因子分解机 模型结构 |
|position bias | 位置偏差 推荐系统经典问题 |