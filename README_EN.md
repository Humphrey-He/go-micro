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
