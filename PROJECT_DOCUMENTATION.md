# Go-Micro 微服务项目说明文档

> 本文档基于代码实现整理，涵盖项目架构、服务职责、核心链路、技术选型及部署方案。

---

## 一、项目概述

### 1.1 业务场景
**电商下单履约系统**：模拟真实电商订单创建、库存预占、异步履约、支付补偿的完整链路。

### 1.2 架构模式
```
HTTP (对外) + gRPC (内部) + RabbitMQ (异步) + MySQL + Redis
```

### 1.3 项目目标
- 完整演示订单创建主链路（创建 → 预占 → 履约 → 完成/取消）
- 展示服务拆分、异步通信、最终一致性与补偿机制
- 提供可运行的微服务样例，含完整 K8s 部署配置

---

## 二、服务清单与职责

| 服务 | 端口 (HTTP/gRPC) | 职责 | 关键能力 |
|------|------------------|------|----------|
| `gateway-api` | 8080/— | 统一入口、鉴权、路由聚合 | 聚合 user/order/inventory/task/payment 等服务 |
| `order-service` | 8081/9081 | 订单生命周期、Outbox 事件 | 状态机、幂等、库存预占调用 |
| `inventory-service` | 8082/9082 | 库存预占/释放/确认 | 缓存保护、reserved_id 追踪 |
| `user-service` | 8083/9083 | 用户信息、认证 | JWT 解析、缓存查询 |
| `task-service` | 8084/9084 | 履约任务编排、超时补偿 | Worker 体系、重试/死信 |
| `payment-service` | 8085/9085 | 支付状态机、超时补偿 | CREATED→SUCCESS/FAILED/TIMEOUT/REFUNDED |
| `activity-service` | 8087/9087 | 活动、秒杀、发券 | Redis 锁 + 事务 + 幂等 |
| `refund-service` | 8086/9086 | 退款处理、MQ 异步 | 退款发起、消费重试 |
| `price-service` | 8088/9088 | 价格计算、历史查询 | gRPC 读接口 |

### 服务依赖关系

```
Client → gateway-api (HTTP)
           ├── user-service (gRPC)
           ├── order-service (gRPC)
           │       └── inventory-service (gRPC)
           ├── inventory-service (gRPC)
           ├── task-service (gRPC)
           ├── payment-service (HTTP 直接调用 via 本地 service 注入)
           ├── refund-service (gRPC)
           ├── activity-service (gRPC)
           └── price-service (gRPC)

order-service → MySQL (orders, order_items, order_events, order_outbox)
              → Redis (缓存)
              → RabbitMQ (发布 order.created)

task-service → MySQL (tasks, task_retry, task_deadletter)
            → RabbitMQ (消费 order.created)
            → order-service (gRPC 回调)
            → inventory-service (gRPC 释放)
```

---

## 三、核心链路

### 3.1 下单履约主链路

```
1. Client → gateway-api: POST /api/v1/orders
2. gateway-api → order-service (gRPC): CreateOrder
3. order-service → inventory-service (gRPC): Reserve
   - 预占成功返回 reserved_id
4. order-service: 更新订单状态 CREATED → RESERVED
5. order-service: 写入 Outbox (order_reserved，状态 PENDING)
6. order-service → RabbitMQ: 发布 order.created (由 Outbox 轮询发送)
7. task-service 消费 order.created 事件
8. task-service → order-service (gRPC): UpdateOrderStatus RESERVED → PROCESSING
9. task-service 创建 15 分钟后超时取消任务 (TIMEOUT_CANCEL)
10. 若超时任务触发时订单未完成：
    - 调用 CancelOrder
    - 调用 ReleaseByOrder 幂等释放库存
11. Client 可通过 GET /api/v1/order-views/{order_no} 查看聚合状态
```

### 3.2 支付状态机

```
CREATED → SUCCESS (支付成功)
CREATED → FAILED (支付失败 → 补偿: 取消订单 + 释放库存)
CREATED → TIMEOUT (支付超时 → 补偿: 取消订单 + 释放库存)
SUCCESS → REFUNDED (退款)
```

### 3.3 库存模型 (TCC 思路)

```
Reserve (预占): available↓ reserved↑, 返回 reserved_id
Confirm (确认): reserved↓ actual↓ (实际扣减)
Release (释放): available↑ reserved↓ (归还)
ReleaseByOrder: 按 order_id 幂等释放所有预占记录
```

---

## 四、技术栈

| 层级 | 技术 | 用途 |
|------|------|------|
| Web 框架 | Gin | HTTP 对外接口 |
| RPC 框架 | gRPC + Protobuf | 服务间同步调用 |
| ORM | GORM / sqlx | 数据库访问 |
| 消息队列 | RabbitMQ | 异步解耦、事件驱动 |
| 缓存 | Redis | 热点数据、分布式锁、幂等 Key |
| 配置 | Viper | 配置加载 |
| 日志 | Zap | 结构化日志 |
| 指标 | Prometheus client_golang | 指标采集 |
| 链路追踪 | OpenTelemetry (预留) | 分布式追踪 |
| 容器化 | Docker + Kubernetes | 部署与编排 |

---

## 五、接口入口

### 5.1 HTTP API (gateway-api)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/orders` | 创建订单 |
| GET | `/api/v1/orders/{id}` | 查询订单详情 |
| GET | `/api/v1/order-views/{order_no}` | 聚合视图 (订单+任务+库存状态) |
| GET | `/api/v1/inventory/{sku_id}` | 查询库存 |
| POST | `/api/v1/payments` | 创建支付 |
| GET | `/api/v1/payments/{id}` | 查询支付 |
| POST | `/api/v1/payments/{id}/success` | 标记支付成功 |
| POST | `/api/v1/payments/{id}/failed` | 标记支付失败 |
| POST | `/api/v1/payments/{id}/timeout` | 标记支付超时 |

### 5.2 gRPC (内部服务)

**OrderService** (`proto/order.proto`):
- `CreateOrder` / `GetOrder` / `GetOrderByBizNo`
- `UpdateOrderStatus` / `CancelOrder`

**InventoryService** (`proto/inventory.proto`):
- `Reserve` / `Release` / `ReleaseByOrder` / `Confirm` / `GetReservation`

**TaskService** (`proto/task.proto`):
- `GetTaskByOrder`

**PaymentService** (`proto/payment.proto`):
- `CreatePayment` / `GetPayment`
- `MarkSuccess` / `MarkFailed` / `MarkTimeout` / `Refund`

**RefundService** (`proto/refund.proto`):
- `Initiate` / `Status` / `Rollback`

---

## 六、数据库表结构

### order-service
- `order`: order_id, biz_no, user_id, status, total_amount, idempotent_key
- `order_item`: order_id, sku_id, quantity, price
- `order_event`: order_id, event, detail
- `order_outbox`: event_type, payload, status (PENDING/SENT)

### inventory-service
- `inventory`: sku_id, available, reserved
- `inventory_reserved`: reserved_id, order_id, sku_id, quantity, status

### task-service
- `task`: task_id, biz_no, order_id, type, status, retry_count
- `task_retry`: task_id, next_retry_at, retry_count
- `task_deadletter`: task_id, reason

### payment-service
- `payments`: payment_id, order_id, amount, status, request_id

---

## 七、统一错误码

| code | message | 含义 |
|-----:|---------|------|
| 0 | OK | 成功 |
| 40001 | invalid request | 参数错误/校验失败 |
| 40101 | missing authorization | 缺少认证信息 |
| 40401 | not found | 资源不存在 |
| 40901 | conflict | 冲突 (如库存不足) |
| 40902 | invalid state | 状态机流转错误 |
| 50001 | internal error | 内部错误 |
| 50201 | upstream unavailable | 下游服务不可用 |

---

## 八、消息与事件

### 8.1 RabbitMQ 配置
- 主交换机: `order.events`
- 主队列: `order.created`
- 死信交换机: `order.events.dlx`
- 死信队列: `order.created.dlq`
- 消费模式: 手动 ACK

### 8.2 Outbox 模式
```
订单写入事务内 → 同时写入 order_outbox (PENDING)
后台轮询 → 发布消息到 MQ → 更新 outbox (SENT)
保证数据与事件一致性
```

---

## 九、缓存保护策略

| 策略 | 实现 |
|------|------|
| 空值缓存防穿透 | 查询结果为空时缓存空值，TTL 短 |
| TTL 抖动防雪崩 | 缓存 TTL ± 随机抖动 |
| 热点变更主动删除 | 库存变更时主动 delete key |
| 互斥锁防击穿 | Redis SETNX 做分布式锁 |

---

## 十、部署方案

### 10.1 本地启动

```bash
# 1. 初始化数据库
mysql -u root -p < deploy/sql/schema.sql

# 2. 启动依赖 (docker-compose)
docker-compose up -d mysql redis rabbitmq

# 3. 生成 swagger 文档
make swagger-swag

# 4. 启动各服务
# 分别在各 cmd 目录下运行
go run cmd/gateway-api/main.go
go run cmd/order-service/main.go
go run cmd/inventory-service/main.go
go run cmd/user-service/main.go
go run cmd/task-service/main.go
go run cmd/payment-service/main.go
go run cmd/refund-service/main.go
go run cmd/activity-service/main.go
go run cmd/price-service/main.go
```

### 10.2 Kubernetes 部署

每个服务在 `deploy/k8s/` 下有独立 YAML:
- `gateway-api.yaml` (无 gRPC 端口)
- `order-service.yaml` (HTTP 8081 + gRPC 9081)
- `inventory-service.yaml` (HTTP 8082 + gRPC 9082)
- `user-service.yaml` (HTTP 8083 + gRPC 9083)
- `task-service.yaml` (HTTP 8084 + gRPC 9084)
- `payment-service.yaml` (HTTP 8085 + gRPC 9085)
- `refund-service.yaml` (HTTP 8086 + gRPC 9086)
- `activity-service.yaml` (HTTP 8087 + gRPC 9087)
- `price-service.yaml` (HTTP 8088 + gRPC 9088)

核心环境变量:
- `MYSQL_DSN`: MySQL 连接串
- `REDIS_ADDR`: Redis 地址
- `MQ_URL`: RabbitMQ 连接串
- `*_GRPC_TARGET`: 各服务 gRPC 地址 (如 `order-service:9081`)

### 10.3 可观测性入口

| 端点 | 用途 |
|------|------|
| `/healthz` | 存活探针 |
| `/readyz` | 就绪探针 |
| `/metrics` | Prometheus 指标 |

---

## 十一、关键实现细节

### 11.1 幂等控制
- 订单创建: `request_id` (幂等 Key) + `idempotent_key` 唯一索引
- 库存预占: `order_id` + `reserved_id` 追踪
- 支付: `request_id` 唯一索引

### 11.2 状态流转规则

**订单状态机**:
```
CREATED  → RESERVED (库存预占成功)
CREATED  → CANCELED (取消)
RESERVED → PROCESSING (履约中)
RESERVED → FAILED (失败)
RESERVED → CANCELED (取消)
PROCESSING → SUCCESS (完成)
PROCESSING → FAILED (失败)
```

**取消规则**: 仅 `CREATED` / `RESERVED` 状态允许取消 (幂等返回成功)

### 11.3 熔断与超时
- 服务间 RPC 超时: 3s (可配置)
- 熔断器: 下游异常时快速失败
- 重试: 上限 3 次

---

## 十二、项目结构

```
go-micro/
├── cmd/                    # 各服务入口
│   ├── gateway-api/
│   ├── order-service/
│   ├── inventory-service/
│   ├── user-service/
│   ├── task-service/
│   ├── payment-service/
│   ├── refund-service/
│   ├── activity-service/
│   └── price-service/
├── internal/               # 业务逻辑
│   ├── gateway/           # HTTP 聚合、跨服务编排
│   ├── order/             # 订单、Outbox
│   ├── inventory/         # 库存预占/释放
│   ├── user/              # 用户、认证
│   ├── task/              # 履约任务、重试/死信
│   ├── payment/           # 支付状态机
│   ├── refund/            # 退款
│   ├── activity/          # 活动、秒杀
│   └── price/             # 价格
├── proto/                 # gRPC IDL 定义
├── pkg/                   # 公共库 (cache, config, db, resilience, httpx)
├── deploy/
│   ├── k8s/               # K8s 部署 YAML
│   └── sql/               # 数据库 schema
├── docs/                  # 文档
│   ├── swagger/           # Swagger UI 资源
│   └── *.md              # 设计文档
├── api/                   # goctl API 定义 (gateway.api)
├── Makefile
└── docker-compose.yml     # 本地开发依赖
```

---

## 十三、测试验证

### 完整流程测试
```bash
# 1. 登录获取 token
POST /api/v1/auth/login
# 2. 查询库存
GET /api/v1/inventory/SKU001
# 3. 创建订单
POST /api/v1/orders
# 4. 查询订单
GET /api/v1/orders/{order_id}
# 5. 查询聚合视图
GET /api/v1/order-views/{biz_no}
# 6. 创建支付
POST /api/v1/payments
# 7. 标记支付成功/失败/超时
POST /api/v1/payments/{payment_id}/success
```

详见: [docs/APIFOX_TEST_GUIDE.md](docs/APIFOX_TEST_GUIDE.md)

### 单元测试
```bash
go test ./... -v -cover -race
```

---

## 十四、当前实现状态

| 模块 | 状态 | 说明 |
|------|------|------|
| gateway-api | ✅ 完整 | 鉴权、聚合、跨服务编排 |
| order-service | ✅ 完整 | 订单、Outbox、状态机 |
| inventory-service | ✅ 完整 | 预占/释放/确认、缓存保护 |
| user-service | ✅ 完整 | 用户、认证、缓存 |
| task-service | ✅ 完整 | Worker、重试、死信、超时补偿 |
| payment-service | ✅ 完整 | 支付状态机、超时扫描、补偿 |
| refund-service | ✅ 完整 | MQ 异步退款、重试 |
| activity-service | ✅ 完整 | 秒杀 Redis 锁 + 幂等 |
| price-service | ✅ 完整 | 价格查询、gRPC |
| K8s 部署 | ✅ 完整 | 所有 9 个服务 YAML |

---

*文档更新时间: 2026/04/14*
