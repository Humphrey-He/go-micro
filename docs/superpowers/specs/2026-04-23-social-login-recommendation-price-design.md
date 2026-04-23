# 社交登录、个性化推荐、价格监控 - 完整需求文档

> 本文档用于前后端对齐后，向后端团队提供可执行落地的技术需求规格。
>
> 涉及功能：社交登录（微信/Google/Apple）、个性化推荐服务、价格监控与提醒

---

## 一、社交登录

### 1.1 功能概述

支持用户通过微信、Google、Apple 三大平台账号快速登录/注册。首次社交登录自动创建绑定账号，支持与已有手机账号关联。

### 1.2 OAuth2.0 PKCE 授权流程

> **为什么用 PKCE**：移动端无法安全存储 Client Secret，PKCE（Proof Key for Code Exchange）通过动态生成的 code_verifier/code_challenge 替代，确保授权码兑换安全性。

#### 微信登录流程

```
┌─────────┐                    ┌──────────────┐                    ┌────────────┐
│  用户   │                    │    前端      │                    │   后端     │
└────┬────┘                    └──────┬───────┘                    └─────┬──────┘
     │  1. 点击微信登录                 │                               │
     │ ──────────────────────────────► │                               │
     │                                 │  2. 生成 code_verifier (随机64字节)   │
     │                                 │  3. 计算 code_challenge = Base64URL(SHA256(code_verifier))
     │                                 │                               │
     │  4. 跳转微信授权页               │                               │
     │  ?appid=wx_APPID               │                               │
     │  &redirect_uri=ENCODED回调地址  │                               │
     │  &response_type=code           │                               │
     │  &scope=snsapi_login           │                               │
     │  &code_challenge=XXX           │                               │
     │  &state=RANDOM                 │                               │
     │ ──────────────────────────────►│                               │
     │                                 │                               │
     │  5. 用户授权后，微信回调至后端    │                               │
     │  /api/v1/auth/social/callback/wechat?code=XXX&state=YYY        │
     │                                 │ ◄──────────────────────────────
     │                                 │                               │
     │                                 │  6. 后端用 code + code_verifier 调用微信API
     │                                 │     获取 access_token + openid
     │                                 │                               │
     │                                 │  7. 用 access_token 获取用户信息 (昵称/头像/unionid)
     │                                 │                               │
     │                                 │  8. 查询/创建用户记录，生成 JWT
     │                                 │                               │
     │  9. 回调前端，带上 JWT           │                               │
     │  /auth/callback?token=XXX&is_new=true                          │
     │ ◄──────────────────────────────│                               │
     │                                 │                               │
```

#### Google 登录流程

```
┌─────────┐                    ┌──────────────┐                    ┌────────────┐
│  用户   │                    │    前端      │                    │   后端     │
└────┬────┘                    └──────┬───────┘                    └─────┬──────┘
     │  1. 点击Google登录               │                               │
     │ ──────────────────────────────► │                               │
     │                                 │  2. Google Identity Services (GIS) 获取 Credential
     │                                 │     (前端使用 google.accounts.id.initialize + renderButton)
     │                                 │                               │
     │  3. 跳转 Google 授权页 (通过GIS) │                               │
     │                                 │                               │
     │  4. 用户授权后，Google 返回 id_token 到前端                       │
     │ ◄────────────────────────────────                               │
     │  返回: { credential: "eyJhbG..." }                            │
     │                                 │                               │
     │  5. 前端将 credential 发送给后端   │                               │
     │  POST /api/v1/auth/social/callback/google                       │
     │  { credential: "eyJhbG..." }                                    │
     │ ──────────────────────────────►│                               │
     │                                 │  6. 后端解析 id_token，验证 signature
     │                                 │     获取 sub/email/name/picture                          │
     │                                 │                               │
     │                                 │  7. 查询/创建用户记录，生成 JWT
     │                                 │                               │
     │  8. 返回 JWT 给前端              │                               │
     │ ◄────────────────────────────────                               │
```

#### Apple 登录流程

```
┌─────────┐                    ┌──────────────┐                    ┌────────────┐
│  用户   │                    │    前端      │                    │   后端     │
└────┬────┘                    └──────┬───────┘                    └─────┬──────┘
     │  1. 点击Apple登录               │                               │
     │ ──────────────────────────────► │                               │
     │                                 │  2. Apple ID SDK (appleid.auth.signIn()) 返回 identityToken
     │                                 │                               │
     │  3. 前端将 identityToken + authorizationCode 发送给后端            │
     │  POST /api/v1/auth/social/callback/apple                        │
     │  { identityToken, authorizationCode, user }                      │
     │ ──────────────────────────────►│                               │
     │                                 │  4. 后端验证 identityToken signature
     │                                 │     (用 Apple 公钥验证 RS256)                         │
     │                                 │                               │
     │                                 │  5. 获取 Apple 用户信息 (sub/email/name)
     │                                 │                               │
     │                                 │  6. 查询/创建用户记录，生成 JWT
     │                                 │                               │
     │  7. 返回 JWT 给前端              │                               │
     │ ◄────────────────────────────────                               │
```

### 1.3 数据模型

#### 用户账号绑定表 (user_social_bindings)

```sql
CREATE TABLE user_social_bindings (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    provider        ENUM('wechat', 'google', 'apple') NOT NULL COMMENT '登录平台',
    provider_user_id VARCHAR(128) NOT NULL COMMENT '第三方平台用户ID',
    access_token    TEXT COMMENT '第三方 access_token (微信专有)',
    refresh_token   TEXT COMMENT '第三方 refresh_token (微信专有，7天有效)',
    token_expires_at DATETIME COMMENT 'access_token 过期时间',
    union_id        VARCHAR(128) COMMENT '微信 UnionID (跨应用识别)',
    nickname        VARCHAR(64) COMMENT '第三方昵称(首次同步)',
    avatar          VARCHAR(512) COMMENT '第三方头像URL',
    bind_status     TINYINT DEFAULT 1 COMMENT '绑定状态: 1正常, 0解绑',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_provider_user (provider, provider_user_id),
    KEY idx_user_id (user_id),
    KEY idx_union_id (union_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户社交账号绑定表';
```

#### 用户主表扩展字段

```sql
-- 建议在 user 表增加以下字段
ALTER TABLE users ADD COLUMN:
    social_provider   ENUM('none', 'wechat', 'google', 'apple') DEFAULT 'none' COMMENT '主要社交登录平台',
    social_uid        VARCHAR(128) COMMENT '主要社交平台用户ID',
    is_phone_bound   TINYINT DEFAULT 0 COMMENT '是否已绑定手机号',
    created_source   ENUM('phone', 'wechat', 'google', 'apple') COMMENT '注册来源';
```

### 1.4 API 接口定义

#### 1.4.1 微信回调接口

```
POST /api/v1/auth/social/callback/wechat

Request:
{
    "code": "微信授权码",
    "code_verifier": "PKCE code_verifier",
    "state": "随机state防止CSRF"
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "token": "JWT Token",
        "expires_in": 86400,
        "is_new_user": true,
        "user_info": {
            "user_id": 12345,
            "nickname": "微信用户",
            "avatar": "https://...",
            "phone_bound": false
        }
    }
}

Error Code:
- 40001: 无效的 code 或 code_verifier
- 40002: 该微信账号已被其他用户绑定
- 50001: 微信服务调用失败
```

#### 1.4.2 Google 回调接口

```
POST /api/v1/auth/social/callback/google

Request:
{
    "credential": "Google id_token (JWT)"
}

Response: 同1.4.1

Error Code:
- 40001: 无效的 credential 或 signature 验证失败
- 40002: 该 Google 账号已被其他用户绑定
```

#### 1.4.3 Apple 回调接口

```
POST /api/v1/auth/social/callback/apple

Request:
{
    "identity_token": "Apple identityToken JWT",
    "authorization_code": "Apple authorizationCode",
    "user": "Apple 传递的 user (可选，首次登录时传递)"
}

Response: 同1.4.1

Error Code:
- 40001: identity_token 验证失败
- 40002: 该 Apple 账号已被其他用户绑定
```

#### 1.4.4 账号绑定查询

```
GET /api/v1/auth/social/bindings

Headers:
    Authorization: Bearer {token}

Response:
{
    "code": 0,
    "data": {
        "bindings": [
            { "provider": "wechat", "bound": true, "nickname": "微信用户", "avatar": "..." },
            { "provider": "google", "bound": false },
            { "provider": "apple", "bound": false }
        ]
    }
}
```

#### 1.4.5 解绑社交账号

```
POST /api/v1/auth/social/unbind

Headers:
    Authorization: Bearer {token}

Request:
{
    "provider": "wechat"  // wechat | google | apple
}

Response:
{
    "code": 0,
    "message": "解绑成功"
}

Error Code:
- 40001: 无法解绑，至少保留一种登录方式
- 40002: 原账号密码未设置，请先设置密码后再解绑
```

#### 1.4.6 账号关联（绑定已有手机账号）

```
POST /api/v1/auth/social/associate

Headers:
    Authorization: Bearer {token}  // 当前社交登录的token

Request:
{
    "phone": "13800138000",
    "code": "短信验证码",
    "action": "bind"  // bind: 绑定到已有账号 | merge: 合并账号(保留手机账号数据)
}

Response:
{
    "code": 0,
    "message": "绑定成功",
    "data": {
        "token": "新的JWT Token",
        "user_id": 12345
    }
}
```

### 1.5 安全设计

| 安全项 | 方案 |
|--------|------|
| CSRF 防护 | OAuth 流程使用 state 参数，前端生成随机字符串，后端校验一致性 |
| Token 泄露 | 微信 access_token 加密存储，敏感操作要求密码验证 |
| 账号绑定风险 | 解绑前要求验证密码或短信验证码，防止盗号解绑 |
| 登录凭证传输 | 全程 HTTPS，code/code_verifier 不在 URL 中传输 |
| 第三方签名验证 | Google/Apple token 必须验证数字签名，不可仅解析 |

### 1.6 前端实现要点

```
前端文件: user-mall/src/api/auth.ts
新增接口:
  - socialCallback(provider, callbackParams)  // 社交登录回调
  - getSocialBindings()                       // 获取绑定状态
  - unbindSocial(provider)                     // 解绑
  - associatePhone(phone, code, action)        // 关联手机号

前端页面: user-mall/src/pages/Login/index.tsx
  - 微信登录按钮: 调用微信SDK (JSSDK) 发起授权
  - Google登录按钮: 使用 Google Identity Services
  - Apple登录按钮: 使用 Sign in with Apple JS (appleid.auth.signIn())

前端存储: authStore.ts
  - 社交登录后，保存 token 和用户信息到 zustand persist
```

---

## 二、个性化推荐服务

### 2.1 功能概述

后端独立推荐服务，基于用户行为数据（浏览/收藏/购买）和商品特征，提供实时个性化推荐。

### 2.2 推荐场景与算法对应

| 推荐场景 | 推荐算法 | 数据来源 | 排序策略 |
|----------|----------|----------|----------|
| 首页"为你推荐" | CF协同过滤 + 热门 | 用户行为日志 | CTR预估排序 |
| 商品详情"看了又看" | Item-CF 物品相似度 | 商品特征向量 | 相似度分数 |
| 商品详情"买了还买" | Item-CF 购买关联 | 购买历史 | 共同购买次数 |
| 购物车"加价购" | 关联规则 + 热门 | 商品共现矩阵 | 支持度×提升度 |
| 支付完成"猜你喜欢" | 混合推荐 | 品类偏好 + 热销 | 个性化分数 |

### 2.3 推荐算法详解

#### 2.3.1 协同过滤（User-CF）

**原理**：找到与目标用户兴趣相似的用户群，推荐这些用户喜欢但目标用户未浏览的商品。

**计算步骤**：
```
1. 构建用户-商品交互矩阵 R (m×n)
   R[u][i] = 行为分数 = 浏览:1, 加购:3, 收藏:5, 购买:10

2. 计算用户相似度 (余弦相似度)
   sim(u, v) = (R[u] · R[v]) / (||R[u]|| × ||R[v]||)

3. 找到Top-K相似用户 S(u)

4. 预测用户对商品的兴趣分
   score(u, i) = Σ(sim(u, v) × R[v][i]) / Σ|sim(u, v)|
   (v ∈ S(u) 且 v 对 i 有交互)

5. 返回Top-N推荐结果
```

**参数配置**：
- K (相似用户数): 20
- N (推荐数量): 20
- 相似度计算: 考虑时间衰减（近期行为权重更高）
- 冷启动: 新用户(<5条行为)使用热门推荐

#### 2.3.2 物品相似度（Item-CF）

**原理**：基于商品共同被用户交互的记录，计算商品间的相似度。

**计算步骤**：
```
1. 构建商品-用户交互倒排表
   item -> [user1, user2, ...]

2. 对每对商品计算共同用户数
   co_users(i, j) = |Users(i) ∩ Users(j)|

3. 计算物品相似度
   sim(i, j) = co_users(i, j) / sqrt(|Users(i)| × |Users(j)|)

4. 给定商品 i，返回 Top-M 相似商品
```

**参数配置**：
- M (相似商品数): 10
- 相似度类型: 余弦相似度
- 热门降权: 过于热门的商品相似度分数适当降低（避免推荐全是爆款）

#### 2.3.3 关联规则（加价购推荐）

**原理**：挖掘"商品A→商品B"的购买关联规则。

**评估指标**：
```
支持度: support(A→B) = count(A∪B) / total_transactions
置信度: confidence(A→B) = count(A∪B) / count(A)
提升度: lift(A→B) = confidence(A→B) / support(B)

推荐条件:
  - support(A→B) > 0.01 (至少1%用户购买过)
  - confidence(A→B) > 0.3 (购买A的用户中30%买了B)
  - lift(A→B) > 1.5 (A和B存在正相关)
```

#### 2.3.4 热门推荐

**热度分数计算**：
```
hot_score(i, t) =
  sales_score(i, t) × 0.4 +      // 近期销量
  view_score(i, t) × 0.3 +       // 近期浏览
  favorites_score(i, t) × 0.3     // 近期收藏

时间衰减: score = hot_score × exp(-λ × days_ago)
λ = 0.1 (7天后权重降为50%)
```

### 2.4 数据模型

#### 2.4.1 用户行为日志表

```sql
CREATE TABLE user_behavior_logs (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    user_id_alias   VARCHAR(64) COMMENT '用户唯一标识(未登录用户匿名ID)',
    sku_id          BIGINT UNSIGNED NOT NULL COMMENT '商品SKU ID',
    behavior_type   ENUM('view', 'favorite', 'cart', 'purchase') NOT NULL COMMENT '行为类型',
    stay_duration   INT COMMENT '停留时长(秒)',
    source          VARCHAR(32) COMMENT '来源: home/recommend/search/cart',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_sku_id (sku_id),
    INDEX idx_created_at (created_at),
    INDEX idx_user_behavior (user_id, behavior_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户行为日志表';
```

#### 2.4.2 商品向量特征表

```sql
CREATE TABLE product_feature_vectors (
    sku_id          BIGINT UNSIGNED PRIMARY KEY,
    category_ids    JSON COMMENT '类目ID列表 [1,2,3]',
    tag_ids         JSON COMMENT '标签ID列表',
    price_range     VARCHAR(32) COMMENT '价格段: 0-50/50-100/...',
    brand_id        BIGINT COMMENT '品牌ID',
    feature_vector  JSON COMMENT '商品特征向量 [0.1, 0.3, 0.5, ...]',
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_category (category_ids),
    INDEX idx_brand (brand_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品特征向量表';
```

#### 2.4.3 推荐结果缓存表

```sql
CREATE TABLE recommendation_cache (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    scene           ENUM('home', 'detail_sim', 'detail_buy', 'cart_add', 'pay_complete') NOT NULL,
    sku_ids         JSON COMMENT '推荐商品ID列表 [123,456,789]',
    scores          JSON COMMENT '对应推荐分数',
    expire_at       DATETIME NOT NULL COMMENT '缓存过期时间',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_scene (user_id, scene),
    INDEX idx_expire_at (expire_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='推荐结果缓存表';
```

#### 2.4.4 商品相似度表

```sql
CREATE TABLE product_similarity (
    sku_id_a        BIGINT UNSIGNED NOT NULL,
    sku_id_b        BIGINT UNSIGNED NOT NULL,
    scene           ENUM('view', 'purchase') NOT NULL COMMENT '相似度场景',
    similarity      DECIMAL(10,6) NOT NULL COMMENT '相似度分数',
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (sku_id_a, sku_id_b, scene),
    INDEX idx_sku_a (sku_id_a),
    INDEX idx_sku_b (sku_id_b)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品相似度表';
```

#### 2.4.5 关联规则表

```sql
CREATE TABLE association_rules (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id_a        BIGINT UNSIGNED NOT NULL COMMENT '前提商品',
    sku_id_b        BIGINT UNSIGNED NOT NULL COMMENT '关联商品',
    support         DECIMAL(10,6) NOT NULL COMMENT '支持度',
    confidence      DECIMAL(10,6) NOT NULL COMMENT '置信度',
    lift            DECIMAL(10,6) NOT NULL COMMENT '提升度',
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_rule (sku_id_a, sku_id_b),
    INDEX idx_sku_a (sku_id_a)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='关联规则表';
```

### 2.5 推荐服务架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        推荐服务架构                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────────┐   │
│  │ 用户行为 │───►│ 行为上报API  │───►│  Kafka / 消息队列     │   │
│  │  前端   │    │ /api/v1/rec/ │    │  user_behavior_topic │   │
│  └──────────┘    │   report     │    └──────────┬───────────┘   │
│                  └──────────────┘             │               │
│                                               ▼               │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────────┐   │
│  │ 推荐API  │◄───│ 推荐引擎     │◄───│  离线计算平台         │   │
│  │ /api/v1 │    │ (实时)       │    │  (协同过滤/关联规则)  │   │
│  │ /rec/.. │    └──────────────┘    └──────────────────────┘   │
│  └────┬─────┘             │                                    │
│       │                  │                                    │
│       │          ┌────────▼────────┐                           │
│       └─────────►│   Redis 缓存    │                           │
│                  │ 推荐结果缓存    │                           │
│                  │ 相似度缓存      │                           │
│                  └────────────────┘                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

数据流转:
1. 前端上报用户行为 (浏览/收藏/加购/购买) 到 /api/v1/rec/report
2. 行为数据写入 Kafka (异步，不阻塞主流程)
3. 离线计算平台消费 Kafka 数据，更新:
   - 用户-商品交互矩阵 (实时更新)
   - 商品相似度 (每日全量+实时增量更新)
   - 关联规则 (每日全量更新)
4. 用户请求推荐时:
   - 先查 Redis 缓存 (推荐结果缓存)
   - 缓存命中直接返回
   - 缓存未命中，实时计算 + 写入缓存 (TTL=15分钟)
```

### 2.6 API 接口定义

#### 2.6.1 行为上报接口

```
POST /api/v1/rec/report

Headers:
    Authorization: Bearer {token}  // 可选，未登录用户用匿名ID

Request:
{
    "sku_id": 12345,
    "behavior_type": "view",  // view | favorite | cart | purchase
    "stay_duration": 30,       // 停留时长(秒)，仅浏览时必填
    "source": "home",          // home | detail | search | cart | recommendation
    "anonymous_id": "uuid-xxx" // 未登录用户匿名ID
}

Response:
{
    "code": 0,
    "message": "success"
}

Note: 异步处理，不影响主业务流程
```

#### 2.6.2 首页推荐接口

```
GET /api/v1/rec/home

Headers:
    Authorization: Bearer {token}

Query:
    page = 1           // 页码
    page_size = 20     // 每页数量

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
                "sales_count": 1000,
                "score": 0.95  // 推荐分数，仅debug模式返回
            }
        ],
        "page": 1,
        "page_size": 20,
        "total": 100
    }
}
```

#### 2.6.3 商品相似推荐接口

```
GET /api/v1/rec/similar/{sku_id}

Headers:
    Authorization: Bearer {token}  // 可选

Query:
    scene = view       // view: 看了又看 | purchase: 买了还买
    limit = 10

Response:
{
    "code": 0,
    "data": {
        "scene": "view",
        "items": [
            {
                "sku_id": 12346,
                "name": "相似商品",
                "price": 89.00,
                "image": "https://...",
                "similarity": 0.85  // 相似度分数
            }
        ]
    }
}
```

#### 2.6.4 购物车加价购推荐

```
GET /api/v1/rec/cart-addon

Headers:
    Authorization: Bearer {token}

Request Body:
{
    "cart_sku_ids": [12345, 12346]  // 购物车中的商品ID列表
}

Response:
{
    "code": 0,
    "data": {
        "items": [
            {
                "sku_id": 12347,
                "name": "加价购商品",
                "price": 29.00,
                "original_price": 59.00,
                "addon_price": 19.00,  // 加价购专享价
                "image": "https://...",
                "match_reason": "常与购物车中商品一起购买"
            }
        ]
    }
}
```

#### 2.6.5 支付完成推荐

```
GET /api/v1/rec/pay-complete

Headers:
    Authorization: Bearer {token}

Query:
    purchased_sku_ids = 12345,12346  // 本次购买的商品ID
    limit = 10

Response:
{
    "code": 0,
    "data": {
        "items": [
            {
                "sku_id": 12347,
                "name": "为你推荐",
                "price": 79.00,
                "image": "https://...",
                "match_reason": "你购买的商品所属品类热门"
            }
        ]
    }
}
```

### 2.7 冷启动策略

| 用户状态 | 推荐策略 |
|----------|----------|
| 新用户 (0条行为) | 热门榜单 + 类目偏好选择入口 |
| 新用户 (1-5条行为) | 热门 + 基于类目热卖的简单协同过滤 |
| 老用户 (>5条行为) | 全量协同过滤 + 热门加权混合 |

```
冷启动接口:
GET /api/v1/rec/cold-start

Response:
{
    "code": 0,
    "data": {
        "hot_items": [...],        // 热门商品
        "category_prefs": [        // 类目偏好选择
            { "category_id": 1, "name": "手机", "image": "..." },
            { "category_id": 2, "name": "服装", "image": "..." }
        ]
    }
}
```

### 2.8 前端实现要点

```
前端文件: user-mall/src/api/recommendation.ts
新增接口:
  - reportBehavior(behavior)          // 上报行为
  - getHomeRecommendations(page)      // 首页推荐
  - getSimilarProducts(skuId, scene)    // 相似推荐
  - getCartAddons(cartSkuIds)          // 加价购
  - getPayCompleteRecommendations()    // 支付完成推荐

前端页面改造:
  - Home/index.tsx: 接入"为你推荐"接口
  - ProductDetail/index.tsx: 接入"看了又看"、"买了还买"
  - Cart/index.tsx: 接入"加价购"
  - PaymentResult/index.tsx: 接入"猜你喜欢"

前端组件:
  - components/RecommendationFeed.tsx  // 通用的推荐流组件
  - components/ProductCard.tsx           // 复用现有商品卡片
```

---

## 三、价格监控与提醒

### 3.1 功能概述

用户可对任意商品设置"降价提醒"，查看商品历史价格走势，降价时收到App内通知。

### 3.2 功能范围

| 功能 | 说明 |
|------|------|
| 设置降价提醒 | 选择商品，设置目标价格（或不设置），开启提醒 |
| 降价历史记录 | 用户查看自己所有设置过的监控记录及价格变化 |
| 商品价格走势 | 展示商品30天/90天/1年历史价格 |
| App内通知 | 降价时写入通知中心，用户打开App可见 |
| 提醒开关 | 用户可开启/关闭价格提醒通知 |

### 3.3 数据模型

#### 3.3.1 商品价格快照表

```sql
CREATE TABLE product_price_snapshots (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id          BIGINT UNSIGNED NOT NULL,
    price           DECIMAL(10,2) NOT NULL COMMENT '记录时的价格',
    original_price  DECIMAL(10,2) COMMENT '记录时的原价',
    snapshot_date   DATE NOT NULL COMMENT '快照日期',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_sku_date (sku_id, snapshot_date),
    INDEX idx_sku_id (sku_id),
    INDEX idx_snapshot_date (snapshot_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品价格快照表';

-- 每日凌晨定时任务生成前一天的快照
-- 保留期限: 1年 (365条记录/商品)
```

#### 3.3.2 用户价格监控表

```sql
CREATE TABLE user_price_watches (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL,
    sku_id          BIGINT UNSIGNED NOT NULL,
    target_price    DECIMAL(10,2) COMMENT '目标价格，NULL表示任意降幅都提醒',
    current_price   DECIMAL(10,2) COMMENT '设置时的商品价格',
    notify_enabled  TINYINT DEFAULT 1 COMMENT '提醒开关: 1开启, 0关闭',
    last_notify_at  DATETIME COMMENT '上次提醒时间',
    notify_count    INT DEFAULT 0 COMMENT '累计提醒次数',
    status          ENUM('active', 'stopped', 'expired') DEFAULT 'active' COMMENT '监控状态',
    expire_at       DATE COMMENT '监控过期日期，NULL表示不过期',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_sku (user_id, sku_id),
    INDEX idx_user_id (user_id),
    INDEX idx_sku_id (sku_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户价格监控表';
```

#### 3.3.3 价格提醒通知表

```sql
CREATE TABLE price_watch_notifications (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL,
    watch_id        BIGINT UNSIGNED NOT NULL,
    sku_id          BIGINT UNSIGNED NOT NULL,
    old_price       DECIMAL(10,2) NOT NULL COMMENT '降价前价格',
    new_price       DECIMAL(10,2) NOT NULL COMMENT '降价后价格',
    discount_amount DECIMAL(10,2) NOT NULL COMMENT '降价金额',
    discount_rate   DECIMAL(5,2) COMMENT '降价幅度(%)',
    message_id      VARCHAR(64) COMMENT '消息中心MessageID',
    is_read         TINYINT DEFAULT 0 COMMENT '是否已读',
    read_at         DATETIME COMMENT '阅读时间',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_watch_id (watch_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='价格提醒通知表';
```

### 3.4 业务流程

#### 3.4.1 设置价格监控

```
用户操作: 商品详情页 → 点击"降价提醒" → 设置目标价格(可选) → 确认

前端: POST /api/v1/price-watch
Request: { sku_id, target_price }

后端:
1. 查询当前商品价格
2. 写入 user_price_watches 表
3. 记录初始价格 (current_price)
4. 返回成功
```

#### 3.4.2 定时检查降价（后端定时任务）

```
定时任务: 每日 09:00, 15:00, 21:00 执行

执行逻辑:
1. 查询所有开启提醒的 active 监控记录
2. 批量查询这些商品当前的实时价格
3. 对比:
   - 条件A: 有目标价格 && 当前价格 <= 目标价格
   - 条件B: 无目标价格 && 之前提醒过 && 当前价格 < 上次通知时的价格
   - 条件C: 无目标价格 && 未提醒过 && 当前价格 < 设置时的价格 && 降幅>=10%
4. 满足任一条件:
   - 写入 price_watch_notifications
   - 更新 last_notify_at
   - 通知数+1
5. 发送App内通知 (写入消息中心)
```

#### 3.4.3 价格走势查询

```
前端请求: GET /api/v1/products/{sku_id}/price-history?period=30d

后端:
1. 根据 period 参数计算日期范围
   30d: 最近30天 (每日快照)
   90d: 最近90天 (每周汇总)
   1y: 最近1年 (每月汇总)
2. 查询 product_price_snapshots
3. 返回价格点列表

Response:
{
    "code": 0,
    "data": {
        "sku_id": 12345,
        "current_price": 89.00,
        "lowest_price": 79.00,
        "lowest_date": "2026-03-15",
        "highest_price": 199.00,
        "highest_date": "2026-01-01",
        "price_points": [
            { "date": "2026-04-01", "price": 99.00 },
            { "date": "2026-04-02", "price": 89.00 },
            ...
        ]
    }
}
```

### 3.5 消息通知设计

```
通知触发后，写入消息中心:

消息类型: price_watch
消息内容:
{
    "type": "price_watch",
    "title": "商品降价啦！",
    "content": "您关注的【商品名称】从 ¥199 降到了 ¥89，点击查看",
    "sku_id": 12345,
    "old_price": 199.00,
    "new_price": 89.00,
    "action": "查看商品",
    "deeplink": "mall://product/12345"
}

通知展示: 消息中心单独Tab "价格提醒"
```

### 3.6 API 接口定义

#### 3.6.1 设置价格监控

```
POST /api/v1/price-watch

Headers:
    Authorization: Bearer {token}

Request:
{
    "sku_id": 12345,
    "target_price": 79.00  // 可选，NULL表示任意降幅提醒
}

Response:
{
    "code": 0,
    "message": "监控设置成功",
    "data": {
        "watch_id": 10001,
        "current_price": 89.00,
        "target_price": 79.00
    }
}

Error Code:
- 40001: 该商品已在监控中
- 40002: 目标价格不能高于当前价格
```

#### 3.6.2 取消价格监控

```
DELETE /api/v1/price-watch/{sku_id}

Headers:
    Authorization: Bearer {token}

Response:
{
    "code": 0,
    "message": "监控已取消"
}
```

#### 3.6.3 获取我的价格监控列表

```
GET /api/v1/price-watch/list

Headers:
    Authorization: Bearer {token}

Query:
    page = 1
    page_size = 20
    status = active  // active | all，可选

Response:
{
    "code": 0,
    "data": {
        "items": [
            {
                "watch_id": 10001,
                "sku_id": 12345,
                "product_name": "商品名称",
                "image": "https://...",
                "current_price": 89.00,
                "original_price": 199.00,
                "target_price": 79.00,
                "notify_enabled": true,
                "lowest_price": 79.00,
                "lowest_price_date": "2026-03-15",
                "created_at": "2026-04-01",
                "price_trend": "down"  // down | stable | up
            }
        ],
        "page": 1,
        "page_size": 20,
        "total": 5
    }
}
```

#### 3.6.4 更新监控设置

```
PUT /api/v1/price-watch/{sku_id}

Headers:
    Authorization: Bearer {token}

Request:
{
    "target_price": 69.00,     // 可选，更新目标价
    "notify_enabled": true      // 可选，开关提醒
}

Response:
{
    "code": 0,
    "message": "更新成功"
}
```

#### 3.6.5 获取商品价格走势

```
GET /api/v1/products/{sku_id}/price-history

Query:
    period = 30d  // 30d | 90d | 1y

Response:
{
    "code": 0,
    "data": {
        "sku_id": 12345,
        "current_price": 89.00,
        "lowest_price": 79.00,
        "lowest_date": "2026-03-15",
        "highest_price": 199.00,
        "highest_date": "2026-01-01",
        "average_price": 125.00,
        "price_points": [
            { "date": "2026-04-01", "price": 99.00 },
            { "date": "2026-04-02", "price": 89.00 }
        ]
    }
}
```

#### 3.6.6 价格提醒通知列表

```
GET /api/v1/notifications/price-watch

Headers:
    Authorization: Bearer {token}

Query:
    page = 1
    page_size = 20

Response:
{
    "code": 0,
    "data": {
        "items": [
            {
                "notification_id": 5001,
                "sku_id": 12345,
                "product_name": "商品名称",
                "image": "https://...",
                "old_price": 199.00,
                "new_price": 89.00,
                "discount_amount": 110.00,
                "discount_rate": 55.3,
                "is_read": false,
                "created_at": "2026-04-02 10:00:00"
            }
        ],
        "unread_count": 3,
        "page": 1,
        "page_size": 20,
        "total": 10
    }
}
```

### 3.7 前端实现要点

```
前端文件: user-mall/src/api/priceWatch.ts
新增接口:
  - setPriceWatch(skuId, targetPrice)       // 设置监控
  - cancelPriceWatch(skuId)                 // 取消监控
  - getPriceWatchList(params)               // 监控列表
  - updatePriceWatch(skuId, params)         // 更新设置
  - getPriceHistory(skuId, period)          // 价格走势
  - getPriceWatchNotifications(params)      // 提醒通知

前端页面改造:
  - ProductDetail/index.tsx:
    - 增加"降价提醒"按钮
    - 弹窗选择目标价格
    - 显示当前商品价格走势（可折叠）

  - PriceWatchList/index.tsx (新页面):
    - 我的价格监控列表
    - 监控状态切换
    - 删除监控

  - Messages/index.tsx:
    - 增加"价格提醒"Tab

前端组件:
  - components/PriceTrendChart.tsx  // 价格走势图 (可用ECharts)
  - components/PriceWatchButton.tsx // 降价提醒按钮
```

### 3.8 后端定时任务

| 任务 | 频率 | 说明 |
|------|------|------|
| 生成价格快照 | 每日 00:05 | 为所有商品生成前一天的快照 |
| 检查降价并发送通知 | 每日 09:00, 15:00, 21:00 | 三次检查，提高触达 |
| 清理过期监控 | 每日 00:30 | 自动清理超期监控记录 |
| 清理过期通知 | 每周一 02:00 | 清理30天前的已读通知 |

---

## 四、数据采集要求

### 4.1 用户浏览行为采集

前端必须上报的行为数据：

| 行为 | 触发时机 | 必填字段 | 选填字段 |
|------|----------|----------|----------|
| view | 商品详情页停留>2秒 | sku_id, source | stay_duration |
| favorite | 点击收藏/取消收藏 | sku_id, action(add/remove) | - |
| cart | 加入购物车/移除 | sku_id, action(add/remove), quantity | - |
| purchase | 支付成功 | sku_id, order_id, quantity | - |

```
上报时机:
- view: 进入商品详情页2秒后上报，离开页面上报stay_duration
- favorite: 实时上报
- cart: 实时上报
- purchase: 支付成功回调中上报

数据清洗:
- 同一商品5分钟内多次view只记1次
- 快速划过(<2秒)不记view
- 取消收藏/购物车移除不上报stay_duration
```

### 4.2 数据质量要求

| 指标 | 要求 |
|------|------|
| 数据延迟 | 行为发生到入库<5秒 |
| 数据完整性 | 有效行为占比>95% |
| 数据准确性 | 用户ID匹配准确率>99.9% |

---

## 五、非功能性要求

### 5.1 性能要求

| 场景 | 指标 |
|------|------|
| 推荐接口 P99 | <200ms |
| 价格走势接口 P99 | <150ms |
| 行为上报接口 P99 | <100ms（异步，不阻塞） |

### 5.2 可用性要求

| 服务 | 可用性 |
|------|--------|
| 社交登录 | 99.9% |
| 推荐服务 | 99.5%（降级：返回热门） |
| 价格监控 | 99.5%（降级：不发通知但不报错） |

### 5.3 安全要求

| 项目 | 要求 |
|------|------|
| OAuth state | 一次性，5分钟内有效 |
| Token存储 | 微信access_token加密存储 |
| 敏感操作 | 解绑/删除需二次验证 |
| 数据脱敏 | 日志中用户ID脱敏 |

---

## 六、依赖关系与优先级

### 6.1 功能依赖

```
社交登录:
  - 前端: 无特殊依赖
  - 后端: 微信开放平台申请AppID + Google Console + Apple Developer

推荐服务:
  - 前端: 行为上报API
  - 后端: 独立推荐服务（需新建），依赖商品基础数据

价格监控:
  - 前端: 价格走势API、商品详情页改造
  - 后端: 价格快照采集任务、通知服务
```

### 6.2 推荐开发优先级

| 阶段 | 内容 | 预计工作 |
|------|------|----------|
| Phase 1 | 基础推荐（热门 + 浏览历史） | 1周 |
| Phase 2 | Item-CF 相似推荐 | 1周 |
| Phase 3 | User-CF 个性化推荐 | 2周 |
| Phase 4 | 关联规则加价购 | 1周 |
| Phase 5 | 实时推荐 + A/B测试 | 2周 |

---

## 七、附录

### 7.1 第三方平台申请资料

| 平台 | 需要准备 |
|------|----------|
| 微信 | 微信开放平台账号 + AppID + AppSecret + 域名白名单 |
| Google | Google Cloud Console 项目 + OAuth 2.0 Client ID |
| Apple | Apple Developer 账号 + Services ID + Private Key |

### 7.2 参考文档

- 微信OAuth2.0文档: https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/Wechat_webpage_Authorization.html
- Google Identity Services: https://developers.google.com/identity/gsi/web
- Sign in with Apple: https://developer.apple.com/documentation/sign_in_with_apple