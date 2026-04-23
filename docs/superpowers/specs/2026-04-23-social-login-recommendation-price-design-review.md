# 社交登录、个性化推荐、价格监控 - 技术评审与扩展

> 本文档对原始需求文档进行资深工程师视角的技术评审，识别潜在问题，补充工程实现细节，并扩展关键技术方案。
>
> 审阅者：资深后端工程师 + 算法工程师
>
> 日期：2026-04-23

---

## 一、社交登录评审

### 1.1 架构评审

#### 原始方案问题

| 问题 | 严重程度 | 说明 |
|------|----------|------|
| 缺少Token泄露后的重放防护 | 高 | 微信access_token被盗后可用于获取用户信息，需增加token绑定设备/用户IP |
| 社交账号解绑逻辑存在竞态 | 中 | 并发解绑可能导致用户失去所有登录方式 |
| state参数存储在内存 | 中 | 分布式环境下state无法共享，建议使用Redis存储 |
| 缺少登录失败次数限制 | 高 | 暴力破解social token的风险 |

#### 补充设计：分布式Session管理

```go
// State管理 - 使用Redis替代内存存储
type OAuthState struct {
    State     string    `json:"state"`
    Provider  string    `json:"provider"`
    CodeChallenge string `json:"code_challenge"`
    RedirectURI string  `json:"redirect_uri"`
    UserID    string    `json:"user_id,omitempty"`  // 关联用户时填充
    CreatedAt int64     `json:"created_at"`
}

const OAuthStateExpire = 10 * time.Minute

// Redis Key: oauth:state:{state}
// 支持分布式环境下多实例共享
```

#### 补充设计：账号绑定安全加固

```go
// 解绑前必须验证 - 防止盗号解绑
type UnbindRequest struct {
    Provider       string `json:"provider" validate:"required"`
    VerifyType     string `json:"verify_type"`  // "password" | "sms_code"
    Password       string `json:"password,omitempty"`
    SmsCode        string `json:"sms_code,omitempty"`
}

// 解绑保护规则
rules := []UnbindRule{
    {MinLoginMethods: 1, RequirePassword: true},  // 只剩一种登录方式时必须验证密码
    {MinLoginMethods: 2, RequirePassword: false}, // 多种登录时可直接解绑
}
```

### 1.2 微信登录算法流程优化

#### 原始流程问题

原始流程第8步"查询/创建用户记录"存在TOCTOU(Time-of-check to time-of-use)问题：
- 并发请求时，两个新用户可能使用同一个微信openid创建两条记录
- 需要使用数据库唯一索引+事务保证原子性

#### 优化后的用户创建逻辑

```go
func (s *UserService) SocialLoginOrCreate(ctx context.Context, provider, openid string, userInfo *SocialUserInfo) (*User, bool, error) {
    // 1. 先尝试查询是否已存在
    binding, err := s.GetSocialBinding(ctx, provider, openid)
    if err == nil && binding != nil {
        // 已存在绑定，更新token并登录
        user, err := s.GetByID(ctx, binding.UserID)
        if err != nil {
            return nil, false, err
        }
        return user, false, nil  // is_new_user = false
    }
    
    // 2. 不存在，创建新用户（使用事务保证原子性）
    tx, err := s.db.BeginTxx(ctx, nil)
    if err != nil {
        return nil, false, err
    }
    defer tx.Rollback()
    
    // 3. 再次检查（悲观锁，防止并发创建）
    binding, err = s.GetSocialBindingForUpdate(tx, provider, openid)
    if err == nil && binding != nil {
        tx.Rollback()
        user, _ := s.GetByID(ctx, binding.UserID)
        return user, false, nil
    }
    
    // 4. 创建用户
    userID := uuid.NewString()
    _, err = tx.Exec(`INSERT INTO users (user_id, created_source, created_at) VALUES (?, ?, NOW())`,
        userID, provider)
    if err != nil {
        return nil, false, err
    }
    
    // 5. 创建绑定记录
    _, err = tx.Exec(`INSERT INTO user_social_bindings (user_id, provider, provider_user_id, union_id, nickname, avatar) 
        VALUES (?, ?, ?, ?, ?, ?)`,
        userID, provider, openid, userInfo.UnionID, userInfo.Nickname, userInfo.Avatar)
    if err != nil {
        return nil, false, err  // 唯一索引保证不会重复
    }
    
    if err = tx.Commit(); err != nil {
        return nil, false, err
    }
    
    user, _ := s.GetByID(ctx, userID)
    return user, true, nil  // is_new_user = true
}
```

### 1.3 Google登录安全性增强

#### 问题识别

原始方案直接使用前端传来的credential，存在以下风险：
- 前端可伪造credential
- 未验证audience是否匹配本应用

#### 增强验证流程

```go
// Google Token验证必须包含：
// 1. 签名验证 (RS256)
// 2. aud验证 - 必须包含本应用的ClientID
// 3. iss验证 - 必须是 accounts.google.com 或 https://accounts.google.com
// 4. exp验证 - 未过期
// 5. iat验证 - 合理时间范围内

type GoogleTokenClaims struct {
    Sub           string `json:"sub"`
    Email         string `json:"email"`
    EmailVerified bool   `json:"email_verified"`
    Name          string `json:"name"`
    Picture       string `json:"picture"`
    Aud           string `json:"aud"`  // 必须验证
    Iss           string `json:"iss"`  // 必须验证
    Exp           int64  `json:"exp"` // 必须验证
    Iat           int64  `json:"iat"` // 必须验证
}

func VerifyGoogleToken(ctx context.Context, credential, expectedClientID string) (*GoogleTokenClaims, error) {
    // 1. 解析JWT (不验证签名)
    parts := strings.Split(credential, ".")
    if len(parts) != 3 {
        return nil, ErrInvalidToken
    }
    
    // 2. 验证签名 - 使用Google公开密钥
    publicKeys, err := fetchGooglePublicKeys(ctx)
    if err != nil {
        return nil, err
    }
    
    // 3. 验证claims
    if claims.Aud != expectedClientID {
        return nil, ErrInvalidAudience
    }
    if claims.Exp < time.Now().Unix() {
        return nil, ErrTokenExpired
    }
    // ... 其他验证
    
    return claims, nil
}
```

### 1.4 Apple登录JS缺陷检测

原始文档提到使用 `appleid.auth.signIn()` 前端SDK，但存在以下问题：

| 问题 | 影响 | 建议 |
|------|------|------|
| Apple登录JS依赖iOS/macOS Safari | Android/Windows无法使用 | 必须配合服务端验证 |
| identityToken可能不包含email | 首次登录可能无email | 需要与Apple重新协商email scope |
| user字段只在首次登录传递 | 后续登录不传user | 需要自己存储user标识 |

### 1.5 扩展：账号关联的合并策略

原始文档 `action: merge` 定义模糊，需要明确数据合并规则：

```go
// 账号合并策略
type MergeStrategy struct {
    PreferredFields []string  // 优先保留的字段，如 ["phone", "email"]
    MergeRules      map[string]MergeRule
}

type MergeRule struct {
    Source   string  // 保留策略: "older" | "newer" | "source" | "target"
    Field    string  // 字段名
}

// 示例：
// nickname: {Source: "newer"}  // 昵称取新的
// phone: {Source: "source"}    // 手机号保留原账号
// orders: {Source: "merge"}    // 订单合并去重
// price_watches: {Source: "merge"}  // 价格监控保留双方
```

---

## 二、个性化推荐服务评审

### 2.1 算法工程师视角评审

#### User-CF 算法问题

| 问题 | 影响 | 建议 |
|------|------|------|
| 相似度计算未考虑时间衰减因子 | 早期行为权重过高 | 应使用exp(-λ×days_ago)衰减 |
| 冷启动阈值5条过于武断 | 新用户质量不可控 | 应根据行为多样性而非数量 |
| Top-K固定20不合理 | 不同用户活跃度差异大 | 应动态计算K值 |

#### 优化：时间加权的余弦相似度

```go
// 时间加权相似度计算
func UserSimilarityWithTimeDecay(r Matrix, u, v int, decayRate float64) float64 {
    now := time.Now().Unix()
    
    // 计算带时间衰减的向量
    weightedU := make([]float64, len(r[0]))
    weightedV := make([]float64, len(r[0]))
    
    for i := 0; i < len(r[0]); i++ {
        // 时间衰减因子
        decayU := math.Exp(-decayRate * float64(now - r.Time[u][i]) / 86400)
        decayV := math.Exp(-decayRate * float64(now - r.Time[v][i]) / 86400)
        
        weightedU[i] = r.Value[u][i] * decayU
        weightedV[i] = r.Value[v][i] * decayV
    }
    
    // 余弦相似度
    dot := dotProduct(weightedU, weightedV)
    normU := vectorNorm(weightedU)
    normV := vectorNorm(weightedV)
    
    if normU == 0 || normV == 0 {
        return 0
    }
    return dot / (normU * normV)
}
```

#### 冷启动策略优化

原始方案新用户使用"热门+类目偏好选择入口"，但存在以下问题：
- 热门推荐不考虑用户已有偏好
- 纯热门推荐无法体现个性化

**优化方案：基于类目热卖的轻量级协同过滤**

```go
// 新用户冷启动推荐策略
func ColdStartRecommendation(userID string, categoryPreferences []int64, limit int) ([]RecItem, error) {
    // 1. 获取用户显式选择的类目偏好
    if len(categoryPreferences) > 0 {
        // 1a. 类目热卖：销量 × 类目匹配度
        items, err := getCategoryBestsellers(categoryPreferences, limit)
        if err != nil {
            return nil, err
        }
        
        // 1b. 基于同龄人热卖（如果能获取用户属性）
        if ageGroup := inferAgeGroup(userID); ageGroup != "" {
            ageItems, _ := getAgeGroupBestsellers(ageGroup, categoryPreferences, limit/2)
            items = mergeWithDeduplication(items, ageItems)
        }
        return items, nil
    }
    
    // 2. 无偏好时使用全局热卖 + 探索策略
    // 80% 热门 + 20% 随机（探索未知类目）
    hotItems, _ := getGlobalBestsellers(int(float64(limit) * 0.8))
    exploreItems, _ := getRandomItems(limit - len(hotItems))
    
    return append(hotItems, exploreItems...), nil
}
```

#### Item-CF 性能优化

商品数量大时，实时计算相似度开销巨大。建议：

```go
// 离线预计算 + 实时查询
type SimilarityService struct {
    // 预先计算的相似商品缓存
    // Key: sku_id, Value: TopM相似商品列表
    similarityCache *redis.Cache
    
    // 实时增量更新（处理新上架商品）
    incrementalUpdateChan chan ItemInteraction
}

// 离线任务：每日全量更新 + 实时增量更新
// 全量更新使用Spark/MapReduce
// 增量更新使用滑动窗口计算
```

#### 扩展：深度学习推荐模型（Phase 6）

当前文档Phase 1-5都是传统算法，建议增加Phase 6：

| 阶段 | 内容 | 算法 | 预期收益 |
|------|------|------|----------|
| Phase 6 | 深度推荐模型 | DeepFM + DIN | CTR预估+15% |

```python
# 推荐模型架构
# Input: 用户行为序列 + 商品特征 + 用户画像
# Model: DeepFM (FM + Deep Neural Network)
# Output: 点击率预估分数

model_input = {
    'user_id': sparse_column('user_id'),
    'behavior_seq': sequence_of_item_ids(max_len=50),
    'item_features': dense_features(['category', 'brand', 'price']),
    'user_context': dense_features(['age_group', 'gender', 'city_tier'])
}

# DIN (Deep Interest Network) 用于捕获候选商品与历史行为的注意力权重
# 特别适用于"看了又看"等场景
```

### 2.2 工程实现评审

#### 高并发下的推荐服务设计

```go
// 推荐服务架构 - 适配现有go-zero框架
type RecommendationService struct {
    // 本地缓存：应对热点商品的高并发
    localCache *bigcache.BigCache
    
    // 分布式缓存：推荐结果
    redisCache *redis.Client
    
    // 消息队列：行为数据异步处理
    mqClient *rabbitmq.Client
    
    // 相似度计算：离线计算结果
    similarityStore *sqlx.DB
}

// 推荐结果缓存策略
// 1. 首页推荐：TTL=15分钟，按用户分桶
// 2. 相似推荐：TTL=1小时，商品维度
// 3. 加价购：TTL=30分钟，购物车维度

func (s *RecommendationService) GetHomeRecommendations(ctx context.Context, userID string, page, pageSize int) ([]RecItem, error) {
    // 1. 检查本地缓存（热点数据）
    cacheKey := fmt.Sprintf("rec:home:%s:%d", userID, page)
    if items, ok := s.localCache.Get(cacheKey); ok {
        return items, nil
    }
    
    // 2. 检查Redis缓存
    if items, err := s.redisCache.Get(ctx, cacheKey).Result(); err == nil {
        var recItems []RecItem
        json.Unmarshal([]byte(items), &recItems)
        // 回填本地缓存
        s.localCache.Set(cacheKey, recItems)
        return recItems, nil
    }
    
    // 3. 缓存未命中，实时计算
    items, err := s.computeHomeRecommendations(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // 4. 写入缓存
    s.redisCache.Set(ctx, cacheKey, items, 15*time.Minute)
    s.localCache.Set(cacheKey, items)
    
    return items, nil
}
```

#### 行为数据上报可靠性

原始文档"异步处理，不影响主业务流程"存在数据丢失风险：

```go
// 问题：前端直接调用API，如果服务端重启会导致数据丢失
// 优化：前端先写入本地队列，重试机制

type BehaviorReporter struct {
    localQueue *PersistentQueue  // IndexedDB本地持久化队列
    batchSize  int
    retryTimes int
}

// 行为上报流程优化
// 1. 前端先将行为写入本地队列
// 2. 批量上报到服务端
// 3. 服务端写入Kafka（不丢失）
// 4. 离线任务消费Kafka

// 服务端保证
func (s *BehaviorService) Report(ctx context.Context, req *ReportRequest) error {
    // 1. 先写入Kafka（高可靠）
    err := s.kafkaClient.Send(ctx, "user_behavior_topic", req)
    if err != nil {
        // Kafka不可用时写入本地文件降级
        s.writeToLocalFile(req)
        return err
    }
    
    // 2. 异步处理，不阻塞返回
    go s.processBehavior(req)
    
    return nil  // 快速返回
}
```

### 2.3 扩展：实时推荐引擎

原始文档推荐架构是"离线计算+缓存"的混合模式，建议增加实时计算能力：

```go
// 实时推荐场景
// 场景：用户刚收藏了商品A，立即推荐相似商品
// 离线计算无法做到实时，需要实时计算补充

type RealTimeRecommender struct {
    // 商品相似度（内存缓存，全量加载）
    itemSimilarity map[int64][]ItemScore
    
    // 用户实时行为（Redis Stream）
    userStream *redis.Client
}

// 处理用户实时行为，更新推荐结果
func (s *RealTimeRecommender) OnUserAction(userID int64, itemID int64, action string) {
    switch action {
    case "favorite":
        // 用户收藏了商品，立即推荐相似商品
        similarItems := s.itemSimilarity[itemID][:10]
        // 推送到用户的消息队列
        s.pushToUserStream(userID, "rec:similar", similarItems)
    case "purchase":
        // 用户购买后，推荐"买了还买"
        boughtAgain := s.getBoughtTogether(itemID)
        s.pushToUserStream(userID, "rec:bought_together", boughtAgain)
    }
}
```

---

## 三、价格监控评审

### 3.1 架构评审

#### 原始方案问题

| 问题 | 严重程度 | 说明 |
|------|----------|------|
| 定时检查间隔9:00-21:00可能漏掉夜间降价 | 中 | 部分平台夜间也会降价 |
| 降价判断条件B存在逻辑缺陷 | 高 | "之前提醒过"的判断依赖last_notify_at字段未考虑时区 |
| 未考虑商品下架/缺货情况 | 中 | 下架商品的价格监控应自动暂停 |
| 缺少防抖机制 | 高 | 价格波动可能导致频繁通知 |

#### 优化：价格变化防抖

```go
// 价格变化防抖策略
type PriceWatchChecker struct {
    // 关键商品监控（用户设置了目标价）
    priorityWatches chan WatchCheck
    
    // 普通商品监控
    normalWatches chan WatchCheck
    
    // 价格变化确认窗口（防止抖动）
    priceConfirmWindow time.Duration
    priceConfirmCache  *redis.Cache  // Key: sku_id, Value: last_confirmed_price
}

const PriceChangeConfirmationWindow = 2 * time.Hour
const MinPriceDropPercent = 5.0  // 最小降价幅度5%

// 降价确认逻辑
func (c *PriceWatchChecker) checkPriceDrop(watch *UserPriceWatch, currentPrice decimal.Decimal) (bool, decimal.Decimal) {
    cacheKey := fmt.Sprintf("price:confirmed:%d", watch.SkuID)
    
    // 获取上次确认价格
    lastConfirmed, err := c.priceConfirmCache.Get(ctx, cacheKey).Result()
    if err != nil {
        // 无确认记录，以监控创建时价格为准
        lastConfirmed = watch.CurrentPrice
    }
    
    // 计算降价幅度
    dropPercent := lastConfirmed.Sub(currentPrice).Div(lastConfirmed).Mul(decimal.NewFromFloat(100))
    
    // 防抖：价格变化需要在确认窗口内持续才确认
    if dropPercent.GreaterThan(decimal.NewFromFloat(MinPriceDropPercent)) {
        // 价格确实下降了，发送通知
        c.priceConfirmCache.Set(ctx, cacheKey, currentPrice.String(), PriceChangeConfirmationWindow)
        return true, currentPrice
    }
    
    return false, currentPrice
}
```

#### 扩展：价格监控状态机

```go
// 监控状态机
type WatchStatus string

const (
    StatusActive  WatchStatus = "active"   // 正常监控中
    StatusPaused  WatchStatus = "paused"  // 暂停（商品下架/缺货）
    StatusTriggered WatchStatus = "triggered" // 已触发提醒
    StatusExpired WatchStatus = "expired" // 已过期
    StatusCancelled WatchStatus = "cancelled" // 用户取消
)

// 状态转换
type StateTransition struct {
    From   WatchStatus
    To     WatchStatus
    Trigger string
    Action  func(*UserPriceWatch) error
}

transitions := []StateTransition{
    {StatusActive, StatusPaused, "product_unavailable", pauseWatch},
    {StatusActive, StatusTriggered, "price_dropped", sendNotification},
    {StatusPaused, StatusActive, "product_available", resumeWatch},
    {StatusTriggered, StatusActive, "user_ack", resetWatch},  // 用户查看后重置
    {StatusActive, StatusExpired, "expire_date_reached", expireWatch},
}

// 状态机实现
func (w *UserPriceWatch) Transition(trigger string) error {
    for _, t := range transitions {
        if w.Status == t.From && t.Trigger == trigger {
            if err := t.Action(w); err != nil {
                return err
            }
            w.Status = t.To
            return nil
        }
    }
    return ErrInvalidTransition
}
```

### 3.2 通知服务集成

项目已有 `notification service`，价格提醒应复用：

```go
// 集成到现有notification service
func (s *PriceWatchService) SendPriceDropNotification(userID string, watch *UserPriceWatch, oldPrice, newPrice decimal.Decimal) error {
    discountAmount := oldPrice.Sub(newPrice)
    discountRate := discountAmount.Div(oldPrice).Mul(decimal.NewFromFloat(100))
    
    // 复用notification service
    return s.notificationService.CreatePriceWatchNotification(
        context.Background(),
        userID,
        &PriceWatchNotification{
            Type:    "price_watch",
            Title:   "商品降价啦！",
            Content: fmt.Sprintf("您关注的【%s】从 ¥%s 降到了 ¥%s，点击查看",
                watch.ProductName, oldPrice.String(), newPrice.String()),
            Metadata: map[string]interface{}{
                "sku_id":        watch.SkuID,
                "old_price":     oldPrice,
                "new_price":     newPrice,
                "discount_rate": discountRate,
                "action":        "view_product",
                "deeplink":      fmt.Sprintf("mall://product/%d", watch.SkuID),
            },
        },
    )
}
```

### 3.3 定时任务优化

原始文档定时任务设计合理，但缺少任务监控：

```go
// 定时任务监控指标
type PriceWatchMetrics struct {
    TotalWatches          prometheus.Gauge
    ActiveWatches        prometheus.Gauge
    TriggeredToday       prometheus.Counter
    NotificationSent     prometheus.Counter
    NotificationFailed   prometheus.Counter
    CheckDuration        prometheus.Histogram
    ProductsChecked      prometheus.Counter
}

// 任务执行监控
func (j *PriceCheckJob) Run() {
    start := time.Now()
    
    count, err := j.execute()
    if err != nil {
        metrics.NotificationFailed.Add(float64(count))
    } else {
        metrics.NotificationSent.Add(float64(count))
    }
    
    metrics.CheckDuration.Observe(time.Since(start).Seconds())
}
```

### 3.4 扩展：价格预测（算法增强）

基于历史价格走势，预测未来价格走势：

```go
// 价格预测服务
type PricePredictor struct {
    model *arima.Model  // ARIMA时序预测
}

// 预测未来7天价格走势
func (p *PricePredictor) Predict(skuID int64, periods int) (*PricePrediction, error) {
    // 1. 获取历史价格数据（90天）
    history, err := p.getPriceHistory(skuID, 90)
    if err != nil {
        return nil, err
    }
    
    // 2. 拟合ARIMA模型
    model, err := arima.Fit(history)
    if err != nil {
        return nil, err
    }
    
    // 3. 预测未来价格
    forecast, confInt := model.Predict(periods)
    
    // 4. 分析趋势
    trend := analyzeTrend(forecast)
    
    return &PricePrediction{
        SkuID:         skuID,
        Forecast:      forecast,
        ConfidenceLow: confInt[0],
        ConfidenceHigh: confInt[1],
        Trend:         trend,  // "rising" | "falling" | "stable"
        RecommendedAction: recommendAction(trend),  // "buy_now" | "wait" | "watch"
    }, nil
}

// 推荐购买时机
func recommendAction(trend string) string {
    switch trend {
    case "falling":
        return "wait"  // 价格还会降
    case "rising":
        return "buy_now"  // 价格要涨了
    default:
        return "watch"  // 价格稳定
    }
}
```

---

## 四、扩展章节：技术风险与应对

### 4.1 高可用设计

#### 社交登录降级

```go
// 微信登录不可用时的降级策略
type SocialLoginFallback struct {
    primaryProvider string
    fallbackProvider string
}

var FallbackStrategy = map[string]string{
    "wechat": "google",  // 微信不可用时降级到Google
    "google": "apple",   // Google不可用时降级到Apple
    "apple": "phone",    // Apple不可用时降级到手机号
}

// 降级触发条件
// 1. 微信API调用超时 > 3秒
// 2. 微信API返回系统错误
// 3. 微信服务不可用（健康检查失败）
```

#### 推荐服务降级

```go
// 推荐服务降级策略
type RecommendationFallback struct{}

func (f *RecommendationFallback) GetHomeRec(userID string) []RecItem {
    // 降级顺序：
    // 1. 用户历史行为推荐（最个性）
    // 2. 类目热卖（次个性）
    // 3. 全局热卖（无个性）
    
    // 尝试用户历史行为
    if items, err := f.getHistoryBasedRec(userID); err == nil && len(items) > 0 {
        return items
    }
    
    // 尝试类目热卖
    if items, err := f.getCategoryBestseller(userID); err == nil && len(items) > 0 {
        return items
    }
    
    // 最终降级到全局热卖
    return f.getGlobalBestseller()
}
```

### 4.2 数据安全与合规

#### 敏感数据处理

```go
// 社交登录敏感数据处理
type SensitiveDataHandler struct {
    encryptKey []byte  // 从KMS获取
}

// 加密存储微信access_token
func (h *SensitiveDataHandler) EncryptAccessToken(token string) (string, error) {
    encrypted, err := aes.Encrypt(token, h.encryptKey)
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(encrypted), nil
}

// 解密时做安全检查
func (h *SensitiveDataHandler) DecryptAccessToken(encrypted string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(encrypted)
    if err != nil {
        return "", ErrInvalidFormat
    }
    
    decrypted, err := aes.Decrypt(data, h.encryptKey)
    if err != nil {
        return "", ErrDecryptFailed
    }
    
    // 解密后检查token格式
    if len(decrypted) < 10 {
        return "", ErrInvalidToken
    }
    
    return decrypted, nil
}

// 日志脱敏
func (s *SocialLoginService) LogLogin(provider, openid string) {
    // openid脱敏：只显示前3位和后3位
    maskedOpenid := maskString(openid, 3, 3, "*")
    log.Printf("Social login: provider=%s, openid=%s", provider, maskedOpenid)
}
```

### 4.3 性能与扩展性

#### 推荐系统扩展性设计

```go
// 水平扩展：用户分桶
type ShardedRecommender struct {
    shardCount   int
    shardClients []*RecommendationService
}

// 根据用户ID哈希分桶
func (s *ShardRecommender) GetShard(userID string) *RecommendationService {
    hash := crc32.ChecksumIEEE([]byte(userID))
    idx := hash % uint32(s.shardCount)
    return s.shardClients[idx]
}

// 数据扩展：历史数据归档
type BehaviorArchiver struct {
    hotThresholdDays = 30  // 30天内数据为热数据
}

// 定期归档冷数据到ClickHouse
func (a *BehaviorArchiver) ArchiveOldData() error {
    cutoff := time.Now().AddDate(0, 0, -a.hotThresholdDays)
    
    // 归档到ClickHouse（列式存储，适合分析查询）
    _, err := a.db.Exec(`
        INSERT INTO user_behavior_archive
        SELECT * FROM user_behavior_logs
        WHERE created_at < ?
    `, cutoff)
    
    // 删除MySQL冷数据
    _, err = a.db.Exec(`
        DELETE FROM user_behavior_logs
        WHERE created_at < ?
    `, cutoff)
    
    return err
}
```

---

## 五、原始文档勘误与补充

### 5.1 数据模型修正

#### user_social_bindings 表

原始文档缺少索引：
```sql
-- 补充索引
KEY idx_provider_union (provider, union_id),  -- 微信UnionID查询
KEY idx_expire_at (token_expires_at)           -- 过期token清理
```

#### user_behavior_logs 表

原始文档 `source` 字段缺少来源枚举定义：
```sql
-- 建议增加枚举约束
behavior_type ENUM('view', 'favorite', 'cart', 'purchase') NOT NULL COMMENT '行为类型',
source ENUM('home', 'detail', 'search', 'cart', 'recommendation', 'unknown') DEFAULT 'unknown' COMMENT '来源',
```

### 5.2 API勘误

#### 1.4.2 Google回调接口

原始文档 Request 字段名有误：
```diff
- { "credential": "Google id_token (JWT)" }
+ { "id_token": "Google id_token (JWT)" }
```

Google OAuth返回的是 `id_token`，不是 `credential`。

#### 2.6.1 行为上报接口

原始文档缺少分页参数和响应说明：
```diff
Response:
{
    "code": 0,
-   "message": "success"
+   "message": "success",
+   "data": {
+       "tracked": true,
+       "deduplicated": false
+   }
}
```

### 5.3 时序图补充

#### 微信登录state参数传递

原始流程图state参数描述不清，补充：

```
前端:
1. 生成随机state (32字节UUID)
2. 存储到Redis: oauth:state:{state} -> {provider, redirect_uri, created_at}
3. 微信回调时携带state参数
4. 后端验证state存在且未过期(10分钟内)
5. 删除已使用的state（一次性）
```

---

## 六、开发优先级调整建议

### 6.1 重新排序

| 阶段 | 内容 | 调整理由 |
|------|------|----------|
| Phase 0 | 社交登录基础（手机号登录+JWT） | 其他功能依赖JWT，需优先完成 |
| Phase 1 | 社交登录接入（选一个平台先做） | 微信/Google/Apple选一个，建议微信 |
| Phase 2 | 价格监控基础（快照+提醒） | 数据依赖商品表，实现相对独立 |
| Phase 3 | 推荐服务基础（行为上报+热门） | 前端改动最小，快速验证 |
| Phase 4 | Item-CF相似推荐 | 依赖行为数据积累 |
| Phase 5 | 社交登录全平台接入 | Phase 1经验复用 |
| Phase 6 | User-CF个性化推荐 | 依赖数据量和用户规模 |
| Phase 7 | 关联规则+加价购 | 业务价值明确 |
| Phase 8 | 深度学习推荐模型 | 资源消耗大，后期投入 |

### 6.2 技术债务预防

| 风险点 | 预防措施 | 验收标准 |
|--------|----------|----------|
| 推荐算法冷启动 | Phase 1先做热门推荐 | 新用户首周留存>40% |
| 社交登录账号冲突 | DB唯一索引+事务 | 并发测试无数据不一致 |
| 价格监控通知轰炸 | 防抖+频率限制 | 同一商品通知间隔>24h |
| 行为数据丢失 | 本地队列+Kafka持久化 | 99.9%数据不丢失 |
| 推荐服务雪崩 | 熔断+降级 | 降级RT<50ms |

---

## 七、测试策略

### 7.1 社交登录测试用例

```go
// 单元测试：微信登录
func TestWechatLogin_NewUser(t *testing.T) {
    // 1. 新用户首次登录
    // 预期：创建用户 + 创建binding + 返回JWT
    
    // 2. 老用户再次登录
    // 预期：更新token + 返回JWT
    
    // 3. 同一微信绑定两个账号（并发）
    // 预期：只有一个成功，另一个返回错误
    
    // 4. code已使用
    // 预期：返回错误
    
    // 5. code过期（>10分钟）
    // 预期：返回错误
}

// 集成测试：完整OAuth流程
func TestWechatOAuthFlow(t *testing.T) {
    // 1. 获取授权码
    // 2. 交换access_token
    // 3. 获取用户信息
    // 4. 登录/注册
    // 5. 验证JWT有效
}
```

### 7.2 推荐算法测试用例

```go
// 算法效果评估
type RecEvaluator struct {
    kValues    []int  // @NDCG@k
    metrics    []Metric
}

func (e *RecEvaluator) Evaluate(testSet []UserBehavior, recommendations map[int64][]RecItem) map[string]float64 {
    results := make(map[string]float64)
    
    for _, k := range e.kValues {
        ndcg := e.calcNDCG(testSet, recommendations, k)
        results[fmt.Sprintf("NDCG@%d", k)] = ndcg
        
        precision := e.calcPrecision(testSet, recommendations, k)
        results[fmt.Sprintf("Precision@%d", k)] = precision
        
        recall := e.calcRecall(testSet, recommendations, k)
        results[fmt.Sprintf("Recall@%d", k)] = recall
    }
    
    return results
}

// 离线评估指标
// - NDCG@k: 排序质量
// - Precision@k: 准确率
// - Recall@k: 召回率
// - Coverage: 覆盖率（推荐商品占总商品比例）
// - Diversity: 多样性（同类商品比例）
```

### 7.3 价格监控测试用例

```go
// 边界条件测试
func TestPriceWatch_CheckPriceDrop(t *testing.T) {
    cases := []struct {
        name        string
        currentP    decimal.Decimal
        targetP     decimal.Decimal
        lastP       decimal.Decimal
        expectNotify bool
    }{
        {"低于目标价", "80", "90", "100", true},
        {"未达目标价", "95", "90", "100", false},
        {"小幅降价不通知", "99", nil, "100", false},
        {"降幅超10%通知", "89", nil, "100", true},
        {"价格反而上涨", "110", nil, "100", false},
    }
    
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            result := checker.checkPriceDrop(c.currentP, c.targetP, c.lastP)
            assert.Equal(t, c.expectNotify, result.notify)
        })
    }
}
```

---

## 八、总结

本文档从资深后端工程师和算法工程师的角度，对原始需求文档进行了全面评审：

### 主要发现

1. **社交登录**：流程基本正确，但缺少分布式环境下的状态管理、账号绑定安全加固
2. **推荐服务**：算法设计合理，但缺少时间衰减、实时计算能力、性能优化
3. **价格监控**：基础功能完善，但缺少防抖机制、状态机管理、预测能力

### 核心建议

1. 社交登录优先实现微信，积累经验后扩展到其他平台
2. 推荐服务分阶段实施，先热门推荐验证数据链路，再逐步增加算法复杂度
3. 价格监控可复用现有notification service，注意防抖和降级策略

### 后续工作

1. 详细技术设计文档（每模块独立设计）
2. 数据库迁移脚本review
3. API接口协议定稿
4. 测试计划制定