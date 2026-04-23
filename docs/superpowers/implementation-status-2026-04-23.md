# 前端功能实现状态报告

> 更新时间: 2026-04-23
>
> 涵盖功能: 社交登录、个性化推荐、价格监控

---

## 一、服务端口对照表

| 服务 | HTTP 端口 | gRPC 端口 | 说明 |
|------|----------|----------|------|
| **Gateway API** | `:8080` | - | 前端统一入口，所有 API 通过此端口 |
| User Service | `:8083` | `:9083` | 用户服务 |
| Order Service | `:8081` | `:9081` | 订单服务 |
| Price Service | `:8088` | `:9088` | 价格服务 |
| Recommendation Service | `:8085` | - | 推荐服务 |

**前端调用示例:**
```typescript
// 前端请求统一通过 VITE_API_BASE_URL 配置
// 开发环境: http://localhost:8080
```

---

## 二、后端 API 实现状态

### ✅ 已实现且可用

| 功能 | 端点 | 方法 | 状态 |
|------|------|------|------|
| 用户登录 | `/api/v1/auth/login` | POST | ✅ |
| 行为上报 | `/api/v1/rec/report` | POST | ✅ |
| 首页推荐 | `/api/v1/rec/home` | GET | ✅ |
| 相似推荐 | `/api/v1/rec/similar/:sku_id` | GET | ✅ |
| 冷启动推荐 | `/api/v1/rec/cold-start` | GET | ✅ |
| 加价购推荐 | `/api/v1/rec/cart-addon` | POST | ✅ |
| 支付完成推荐 | `/api/v1/rec/pay-complete` | GET | ✅ |
| 价格计算 | `/api/v1/price/calculate` | POST | ✅ |
| 价格历史 | `/api/v1/price/history` | GET | ✅ |
| 商品价格走势 | `/api/v1/products/:sku_id/price-history` | GET | ✅ |

### ✅ 已实现 (刚刚注册到 Gateway)

| 功能 | 端点 | 方法 | 状态 |
|------|------|------|------|
| 微信登录回调 | `/api/v1/auth/social/callback/wechat` | POST | ✅ |
| Google登录回调 | `/api/v1/auth/social/callback/google` | POST | ✅ |
| Apple登录回调 | `/api/v1/auth/social/callback/apple` | POST | ✅ |
| 社交绑定查询 | `/api/v1/auth/social/bindings` | GET | ✅ |
| 解绑社交账号 | `/api/v1/auth/social/unbind` | POST | ✅ |
| 关联手机号 | `/api/v1/auth/social/associate` | POST | ✅ |
| 设置价格监控 | `/api/v1/price-watch` | POST | ✅ |
| 取消价格监控 | `/api/v1/price-watch/:sku_id` | DELETE | ✅ |
| 获取监控列表 | `/api/v1/price-watch/list` | GET | ✅ |
| 更新监控设置 | `/api/v1/price-watch/:sku_id` | PUT | ✅ |
| 价格提醒通知 | `/api/v1/notifications/price-watch` | GET | ✅ |

### ❌ 未实现

| 功能 | 端点 | 说明 |
|------|------|------|
| 发送短信验证码 | `/api/v1/auth/sms/send` | 前端已实现，等待后端 |
| 短信验证码登录 | `/api/v1/auth/sms/login` | 前端已实现，等待后端 |
| 用户注册 | `/api/v1/auth/register` | 前端已实现，等待后端 |
| 登出 | `/api/v1/auth/logout` | 前端已实现，等待后端 |

---

## 三、社交登录前端实现状态

### ✅ 已完成功能

| 功能模块 | 文件位置 | 状态 |
|---------|---------|------|
| 社交登录 API | `src/api/auth.ts` | ✅ |
| PKCE 工具函数 | `src/utils/socialLogin.ts` | ✅ |
| 社交登录按钮组件 | `src/components/SocialLoginButtons/index.tsx` | ✅ |
| 登录页集成 | `src/pages/Auth/Login.tsx` | ✅ |
| OAuth 回调页面 | `src/pages/Auth/SocialCallback.tsx` | ✅ |
| 账号关联页面 | `src/pages/Auth/SocialBind.tsx` | ✅ |
| Auth Store 扩展 | `src/stores/authStore.ts` | ✅ |
| ProtectedRoute 组件 | `src/components/ProtectedRoute.tsx` | ✅ |

### ⚠️ 注意事项

- 微信/Google/Apple 实际登录需要各自平台的 App ID 配置到 `.env`
- 后端社交登录回调接口已注册但需要真实的第三方平台配置才能测试

---

## 四、个性化推荐前端实现状态

### ✅ 已完成功能

| 功能模块 | 文件位置 | 状态 |
|---------|---------|------|
| 推荐 API 模块 | `src/api/recommendation.ts` | ✅ |
| 推荐流组件 | `src/components/RecommendationFeed/index.tsx` | ✅ |
| 推荐商品项组件 | `src/components/RecommendationFeed/RecommendationItem.tsx` | ✅ |
| 行为上报 Hook | `src/hooks/useBehaviorReport.ts` | ✅ |
| 商品详情页集成 | `src/pages/ProductDetail/index.tsx` | ✅ |

### ⚠️ 待验证

| 功能模块 | 说明 |
|---------|------|
| 首页推荐集成 | 需要后端 `/rec/home` 接口返回数据后验证 |
| 购物车加价购 | 需要后端 `/rec/cart-addon` 接口返回数据后验证 |
| 支付完成推荐 | 需要后端 `/rec/pay-complete` 接口返回数据后验证 |

---

## 五、价格监控前端实现状态

### ✅ 已完成功能

| 功能模块 | 文件位置 | 状态 |
|---------|---------|------|
| 价格监控 API | `src/api/priceWatch.ts` | ✅ |
| 价格走势图组件 | `src/components/PriceTrendChart/index.tsx` | ✅ |
| 降价提醒按钮 | `src/components/PriceWatchButton/index.tsx` | ✅ |
| 商品详情页集成 | `src/pages/ProductDetail/index.tsx` | ✅ |

### ⚠️ 待验证

| 功能模块 | 说明 |
|---------|------|
| 价格监控列表页 | 需要后端返回真实数据后验证 |
| 消息通知页 | 需要后端通知功能完成后验证 |

---

## 六、用户信息与认证修复状态

### ✅ 已修复问题

| 问题 | 修复方案 | 验证状态 |
|------|---------|---------|
| ProtectedRoute 未集成 | 在 `App.tsx` 中为所有 `requiresAuth: true` 路由包裹 `ProtectedRoute` 组件 | ✅ Build 通过 |
| 行为追踪未集成 | 在 `ProductDetail/index.tsx` 中集成 `useBehaviorReport` hook | ✅ Build 通过 |
| 短信登录未实现 | 在 `Login.tsx` 中实现完整短信登录流程，等待后端 API | ⚠️ 待后端 |
| 登出 API 未调用 | 在 `authStore.ts` 的 `logout` 方法中调用 `logoutApi()` | ⚠️ 待后端实现 |

---

## 七、文件变更清单

### 修改的文件 (2026-04-23 修复)

```
user-mall/src/
├── api/auth.ts                              # 增加 smsLogin API
├── pages/Auth/Login.tsx                     # 实现完整短信登录
├── pages/ProductDetail/index.tsx             # 集成行为追踪
├── stores/authStore.ts                       # logout 调用 API

internal/gateway/
├── handler.go                               # 注册社交登录和价格监控路由
├── run.go                                   # 导入 pricewatch 包

internal/social/
├── handler.go                              # 修改为接受 RouterGroup

internal/pricewatch/
└── handler.go                              # 修改为接受 RouterGroup
```

---

## 八、后端待实现 API

### 高优先级

1. **发送短信验证码** `POST /api/v1/auth/sms/send`
   - 参数: `{ phone: string, type: 'login' | 'register' | 'reset_password' }`
   - 返回: `{ expires_in: number }`

2. **短信验证码登录** `POST /api/v1/auth/sms/login`
   - 参数: `{ phone: string, code: string }`
   - 返回: `LoginResponse` (token + user)

3. **用户注册** `POST /api/v1/auth/register`
   - 参数: `{ phone: string, code: string, password: string, invite_code?: string }`
   - 返回: `LoginResponse`

4. **登出** `POST /api/v1/auth/logout`
   - 需要认证
   - 返回: `{ message: string }`

---

## 九、环境变量配置

### 后端 (.env 或环境变量)

```bash
# Gateway
GATEWAY_ADDR=:8080

# Services
USER_GRPC_TARGET=localhost:9083
ORDER_GRPC_TARGET=localhost:9081
PRICE_GRPC_TARGET=localhost:9088
REDIS_ADDR=localhost:6379

# Recommendation
RECOMMENDATION_URL=http://localhost:8085
```

### 前端 (.env)

```bash
# 社交登录配置
VITE_WECHAT_APP_ID=your_wechat_appid
VITE_GOOGLE_CLIENT_ID=your_google_clientid.apps.googleusercontent.com
VITE_APPLE_CLIENT_ID=com.yourcompany.yourapp

# API 基础地址
VITE_API_BASE_URL=http://localhost:8080
```

---

## 十、联调测试指南

### 测试步骤

1. **启动后端服务**
   ```bash
   # 启动 Gateway (端口 8080)
   go run cmd/gateway-api/main.go

   # 启动 Recommendation Service (端口 8085)
   go run cmd/recommendation-service/main.go
   ```

2. **启动前端开发服务器**
   ```bash
   cd user-mall
   npm run dev
   ```

3. **API 测试**
   ```bash
   # 测试登录
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"test","password":"test"}'

   # 测试价格走势
   curl http://localhost:8080/api/v1/products/12345/price-history?period=30d

   # 测试推荐
   curl http://localhost:8080/api/v1/rec/home
   ```
