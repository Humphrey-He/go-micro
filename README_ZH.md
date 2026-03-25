# Go 微服务实战项目（中文说明）

## 项目背景
本项目以“电商下单履约”为核心业务场景，采用 Go 微服务架构模拟真实团队交付流程。目标是覆盖面试高频点与工程化实践：服务拆分、RPC、消息队列、数据一致性、缓存稳定性、可观测性与部署运维等。

## 项目技术选型
- 语言：Go
- Web：Gin
- RPC：gRPC
- 数据库：MySQL
- 缓存：Redis
- 消息队列：RabbitMQ（支持 DLQ）
- 配置：环境变量 + 统一配置读取
- 日志：Zap
- API 文档：Swagger（swag + gin-swagger）

## 项目亮点
- **微服务拆分清晰**：网关、订单、库存、用户、任务编排各司其职。
- **内部调用 gRPC**：接口强约束、低延迟、易治理。
- **Outbox 一致性**：订单写入与消息写入同事务，异步投递保证最终一致。
- **可靠消息消费**：RabbitMQ 手动 ACK + DLQ 死信队列。
- **缓存稳定性设计**：空值缓存防穿透、TTL 抖动防雪崩。
- **统一错误码**：跨服务一致的错误码与错误语义。
- **可测试性**：核心逻辑单测覆盖（订单、库存、缓存）。
- **超时取消补偿**：订单超时触发取消任务，释放库存并保持幂等。
 - **支付状态机**：支付成功/失败/超时/退款全流程，失败触发补偿。

## 聚合视图状态映射
聚合接口 `GET /api/v1/order-views/{order_no}` 会返回主状态与明细状态。`view_status` 是面向调用方的统一状态，当底层状态冲突时按优先级规则映射：

优先级规则（从高到低）：
1. `order_status == CANCELED` 且 `task_type == TIMEOUT_CANCEL` -> `view_status = TIMEOUT`
2. `order_status == CANCELED` -> `view_status = CANCELED`
3. `task_status == DEAD` -> `view_status = DEAD`
4. `task_status == FAILED` -> `view_status = FAILED`
5. `order_status == SUCCESS` -> `view_status = SUCCESS`
6. `task_status == RUNNING` -> `view_status = PROCESSING`
7. `order_status == RESERVED` 且 `task_status in (PENDING, NOT_FOUND)` -> `view_status = PENDING`

返回结构示例：
```json
{
  "order_no": "BIZ-xxxx",
  "view_status": "CANCELED",
  "order_status": "CANCELED",
  "task_status": "DEAD",
  "reservation_status": "RELEASED",
  "cancel_reason": "timeout"
}
```

示例映射表：
```
order_status   task_status   reservation_status   view_status
RESERVED       PENDING       RESERVED            PENDING
PROCESSING     RUNNING       RESERVED            PROCESSING
SUCCESS        SUCCESS       CONFIRMED           SUCCESS
CANCELED       FAILED        RELEASED            CANCELED
CANCELED       DEAD          RELEASED            TIMEOUT/DEAD
```

## 快速开始
1. 初始化数据库：执行 `deploy/sql/schema.sql`
2. 启动依赖：MySQL、Redis、RabbitMQ
3. 启动服务：
   - gateway-api
   - order-service
   - inventory-service
   - user-service
   - task-service
   - payment-service

## Swagger 文档
手动生成：
```bash
swag init -g cmd/gateway-api/main.go -o ./docs/swagger
```
访问：`http://localhost:8080/swagger/index.html`

## 测试与证据
CI 已配置：`.github/workflows/ci.yml`  
执行项：
- `go test ./... -v -cover`
- `go test ./... -race`

建议本地执行：
```bash
go test ./... -v -cover -race
```
说明：
- `-race` 需要启用 CGO 且系统具备 C 编译器（Windows 需安装 gcc）
- 如果启用 Go Toolchain 下载，请确保网络可达或设置 `GOTOOLCHAIN=local`
覆盖重点：
- 状态映射逻辑：覆盖所有状态分支
- 库存幂等释放：覆盖重复释放与异常场景
- timeout cancel 链路：覆盖完整补偿流程

## 缓存指标与验证
指标：
- `cache_hits_total{cache="..."}`  
- `cache_misses_total{cache="..."}`

PromQL 示例：
```
sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m])))
```

验证方式：
1. 调用订单/库存/用户查询接口，观察命中率变化  
2. 停止 Redis 后确认降级为本地缓存，日志有告警  
3. 访问 `/metrics` 查看命中/未命中计数

## 可观测性与部署演示
详见：[docs/可观测性与部署演示.md](./docs/可观测性与部署演示.md)

## 目录结构
```
cmd/            各服务启动入口
internal/       业务核心逻辑
pkg/            公共库与工具
proto/          gRPC 协议定义与生成代码
deploy/         数据库与部署脚本
docs/           文档与说明
```

## 运行依赖
- MySQL 8
- Redis
- RabbitMQ
 - Elasticsearch/Fluentd（可选，用于日志系统）

## 配置项（核心）
- `MYSQL_DSN`
- `REDIS_ADDR`
- `MQ_URL`
- `ORDER_GRPC_ADDR` / `INVENTORY_GRPC_ADDR`
