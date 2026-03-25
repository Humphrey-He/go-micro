# Go Microservices Project (English README)

## Background
This project simulates a real-world e-commerce order fulfillment flow with a Go microservice architecture. It targets both interview preparation and practical engineering skills: service decomposition, RPC, messaging, data consistency, cache resilience, observability, and operations.

## Tech Stack
- Language: Go
- Web: Gin
- RPC: gRPC
- Database: MySQL
- Cache: Redis
- Messaging: RabbitMQ (with DLQ)
- Config: Environment variables + unified config loader
- Logging: Zap
- API Docs: Swagger (swag + gin-swagger)

## Highlights
- **Clear service boundaries**: gateway, order, inventory, user, task orchestration.
- **gRPC internal calls**: strong contracts, low latency, easier governance.
- **Outbox consistency**: order write + outbox insert in one transaction; async publishing ensures eventual consistency.
- **Reliable consumption**: manual ACK + DLQ for failed messages.
- **Cache resilience**: null-cache for penetration, TTL jitter to mitigate avalanche.
- **Unified error codes** across services.
- **Testability**: unit tests for order, inventory, and cache.

## Aggregated View Status Mapping
The aggregated endpoint `GET /api/v1/order-views/{order_no}` returns a primary `view_status` and detailed states. `view_status` is the unified status for callers, and conflicts are resolved by priority rules below:

Priority rules (high to low):
1. `order_status == CANCELED` and `task_type == TIMEOUT_CANCEL` -> `view_status = TIMEOUT`
2. `order_status == CANCELED` -> `view_status = CANCELED`
3. `task_status == DEAD` -> `view_status = DEAD`
4. `task_status == FAILED` -> `view_status = FAILED`
5. `order_status == SUCCESS` -> `view_status = SUCCESS`
6. `task_status == RUNNING` -> `view_status = PROCESSING`
7. `order_status == RESERVED` and `task_status in (PENDING, NOT_FOUND)` -> `view_status = PENDING`

Response example:
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

Example mapping table:
```
order_status   task_status   reservation_status   view_status
RESERVED       PENDING       RESERVED            PENDING
PROCESSING     RUNNING       RESERVED            PROCESSING
SUCCESS        SUCCESS       CONFIRMED           SUCCESS
CANCELED       FAILED        RELEASED            CANCELED
CANCELED       DEAD          RELEASED            TIMEOUT/DEAD
```

## Quick Start
1. Initialize DB: run `deploy/sql/schema.sql`
2. Start dependencies: MySQL, Redis, RabbitMQ
3. Start services:
   - gateway-api
   - order-service
   - inventory-service
   - user-service
   - task-service

## Swagger Docs
Generate manually:
```bash
swag init -g cmd/gateway-api/main.go -o ./docs/swagger
```
Open: `http://localhost:8080/swagger/index.html`

## Repository Structure
```
cmd/            service entrypoints
internal/       core business logic
pkg/            shared libraries
proto/          gRPC definitions and generated code
deploy/         database/deploy scripts
docs/           documentation
```

## Dependencies
- MySQL 8
- Redis
- RabbitMQ

## Key Configs
- `MYSQL_DSN`
- `REDIS_ADDR`
- `MQ_URL`
- `ORDER_GRPC_ADDR` / `INVENTORY_GRPC_ADDR`
