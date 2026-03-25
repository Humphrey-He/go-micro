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

## 快速开始
1. 初始化数据库：执行 `deploy/sql/schema.sql`
2. 启动依赖：MySQL、Redis、RabbitMQ
3. 启动服务：
   - gateway-api
   - order-service
   - inventory-service
   - user-service
   - task-service

## Swagger 文档
手动生成：
```bash
swag init -g cmd/gateway-api/main.go -o ./docs/swagger
```
访问：`http://localhost:8080/swagger/index.html`

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

## 配置项（核心）
- `MYSQL_DSN`
- `REDIS_ADDR`
- `MQ_URL`
- `ORDER_GRPC_ADDR` / `INVENTORY_GRPC_ADDR`
