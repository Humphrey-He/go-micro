# Go 微服务实战项目（中文说明）


**[English](./README.md) | [中文](./README.md)**


# 项目介绍

本项目是基于 Go 语言开发的电商下单履约微服务实战项目，严格遵循企业级开发流程与工程化规范，模拟电商场景中从用户下单、库存扣减、订单履约到任务调度的全链路业务流程。项目核心目标是覆盖后端面试高频考点与工程化实践要点，帮助开发者掌握微服务架构设计、分布式系统问题解决、高可用架构搭建等核心能力，可直接作为个人面试核心项目，快速提升求职竞争力。

项目地址：[https://github.com/Humphrey-He/go-micro](https://github.com/Humphrey-He/go-micro)

# 项目背景与目标

## 项目背景

随着电商行业的快速发展，高并发、高可用、数据一致性已成为后端开发的核心诉求。本项目以“电商下单履约”为核心业务场景，拆解真实业务中的核心流程，采用微服务架构模式，解决分布式场景下的服务通信、数据一致性、缓存稳定性、可观测性等工程痛点，模拟真实团队交付流程，打造可直接落地、可扩展的实战项目。

## 核心目标

- 覆盖微服务开发全流程：从服务拆分、接口设计到部署运维，完整还原企业级开发场景；

- 攻克面试高频考点：聚焦服务拆分、RPC通信、消息队列应用、数据一致性、缓存稳定性等核心面试要点；

- 落地工程化实践：规范代码结构、完善单测覆盖、实现可观测性，提升项目可维护性与可扩展性；

- 打造面试级实战项目：项目亮点突出、技术栈主流，可直接作为个人核心项目用于求职面试。

# 目录结构

项目目录严格遵循 Go 工程化规范，结构清晰、职责分明，便于开发维护与扩展，具体如下：

```bash
cmd/          # 各微服务启动入口（网关、订单、库存等服务）
internal/     # 业务核心逻辑（各服务的业务实现、数据处理）
pkg/          # 公共库及工具包（通用工具、中间件、公共逻辑）
proto/        # gRPC 协议定义与自动生成代码（服务间RPC通信依赖）
deploy/       # 部署相关资源文件（数据库脚本、部署配置）
docs/         # 项目文档（接口文档、说明文档、Swagger资源）
.github/      # CI/CD 配置（GitHub Actions 工作流）
.gitignore    # Git 忽略文件配置
go.mod        # Go 模块依赖配置
go.sum        # Go 依赖版本锁定文件
README.md     # 项目英文说明文档
README_ZH.md  # 项目中文说明文档（本文档）
README_EN.md  # 项目英文补充说明文档
```

# 技术选型

项目采用主流、成熟的技术栈，贴合企业实际开发场景，兼顾性能、可靠性与可维护性，具体选型如下：

- **开发语言**：Go（简洁高效、原生支持高并发，适配微服务开发场景）

- **Web框架**：Gin（轻量高性能，用于网关服务API开发）

- **RPC通信**：gRPC（接口强约束、低延迟、易治理，用于服务间内部通信）

- **数据库**：MySQL 8（关系型数据库，存储核心业务数据，保障数据一致性）

- **缓存**：Redis（高性能缓存，缓解数据库压力，提升接口响应速度）

- **消息队列**：RabbitMQ（支持死信队列DLQ（Dead-Letter Queue），实现可靠消息投递与消费）

- **配置管理**：环境变量 + 统一配置加载（简洁灵活，适配不同部署环境）

- **日志系统**：Zap（高性能结构化日志，便于问题排查与日志分析）

- **API文档**：Swagger（基于swag + gin-swagger，自动生成接口文档，便于接口调试）

- **CI/CD**：GitHub Actions（实现自动化构建、测试，提升开发交付效率）

# 项目核心亮点

本项目区别于普通练手项目，聚焦企业级工程化实践与面试高频痛点，核心亮点如下：

1. **微服务拆分清晰合理**：严格按照业务职责拆分服务，网关、订单、库存、用户、任务编排五大服务各司其职，低耦合、高内聚，符合微服务设计原则，可直接参考用于企业级项目拆分。

2. **内部通信标准化**：采用 gRPC 实现服务间通信，定义标准化 proto 协议，实现接口强约束，降低服务间耦合，同时保证通信低延迟、易治理，贴合企业微服务通信实际场景。

3. **数据一致性保障**：采用 Outbox 模式，实现订单写入与消息写入同事务，通过异步投递保障分布式场景下的数据最终一致性，解决电商场景中“下单成功但库存未扣减”“消息丢失”等核心痛点。

4. **可靠消息消费机制**：基于 RabbitMQ 实现手动 ACK 确认机制，搭配 DLQ（Dead-Letter Queue）死信队列，处理消息消费失败场景，避免消息丢失、重复消费，保障消息消费的可靠性与稳定性。

5. **缓存稳定性设计**：针对缓存穿透、缓存雪崩等高频问题，实现空值缓存、TTL 抖动优化等方案，提升缓存系统稳定性，避免缓存异常导致的服务雪崩，适配高并发业务场景需求。

6. **统一错误码规范**：定义跨服务统一的错误码与错误语义，简化问题排查流程，提升服务间协作效率，符合企业级项目错误处理规范。

7. **可测试性强**：核心逻辑（订单、库存、缓存）实现单测覆盖，规范测试流程，提升代码质量，同时便于后续功能迭代与问题排查，体现工程化开发思维。

8. **工程化规范完善**：遵循 Go 工程化规范，目录结构清晰、代码风格统一，集成 CI/CD 流程、接口文档自动化生成，贴合企业实际开发交付标准。

# 核心业务逻辑

## 聚合视图状态映射

项目提供聚合接口 `GET /api/v1/order-views/{order_no}`，用于返回订单的主状态与明细状态，其中 `view_status` 为面向调用方的统一状态，当底层状态（订单状态、任务状态）发生冲突时，按以下优先级规则映射（从高到低）：

1. 当 `order_status == CANCELED` 且 `task_type == TIMEOUT_CANCEL` 时，`view_status = TIMEOUT`（订单超时取消）；

2. 当 `order_status == CANCELED` 时，`view_status = CANCELED`（订单主动取消）；

3. 当 `task_status == DEAD` 时，`view_status = DEAD`（任务执行失败且无法重试）；

4. 当 `task_status == FAILED` 时，`view_status = FAILED`（任务执行失败，可重试）；

5. 当 `order_status == SUCCESS` 时，`view_status = SUCCESS`（订单履约成功）；

6. 当 `task_status == RUNNING` 时，`view_status = PROCESSING`（订单履约中）；

7. 当 `order_status == RESERVED` 且`task_status in (PENDING, NOT_FOUND)` 时，`view_status = PENDING`（订单待履约）。

### 返回结构示例

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

### 状态映射示例表

```bash
order_status  task_status  reservation_status  view_status
RESERVED      PENDING      RESERVED            PENDING
PROCESSING    RUNNING      RESERVED            PROCESSING
SUCCESS       SUCCESS      CONFIRMED           SUCCESS
CANCELED      FAILED       RELEASED            CANCELED
CANCELED      DEAD         RELEASED            TIMEOUT/DEAD
```

# 快速开始

## 运行依赖

请确保本地或服务器已安装以下依赖，且版本符合要求：

- MySQL 8（核心业务数据存储）

- Redis（缓存服务）

- RabbitMQ（消息队列，需支持DLQ死信队列）

- Go 1.18+（项目开发语言）

## 部署步骤

1. **克隆项目**`git clone https://github.com/Humphrey-He/go-micro.git
cd go-micro`注意：克隆时若出现“link dead”报错，请检查GitHub链接有效性，并确保网络连接稳定。

2. **初始化数据库**执行部署脚本初始化数据库表结构：`mysql -u 用户名 -p 密码 < deploy/sql/schema.sql`

3. **配置核心参数**设置环境变量，配置核心依赖地址（也可根据需求修改配置加载逻辑）：

    - `MYSQL_DSN`：MySQL 连接地址（格式：user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local）

    - `REDIS_ADDR`：Redis 连接地址（格式：host:port）

    - `MQ_URL`：RabbitMQ 连接地址（格式：amqp://user:password@host:port/）

    - `ORDER_GRPC_ADDR`：订单服务 gRPC 地址

    - `INVENTORY_GRPC_ADDR`：库存服务 gRPC 地址

4. **启动依赖服务**启动 MySQL、Redis、RabbitMQ，确保服务正常运行，可通过对应客户端验证连接有效性。

5. **启动微服务**依次启动以下服务（可在不同终端执行）：`# 启动网关服务
go run cmd/gateway-api/main.go

# 启动订单服务
go run cmd/order-service/main.go

# 启动库存服务
go run cmd/inventory-service/main.go

# 启动用户服务
go run cmd/user-service/main.go

# 启动任务服务
go run cmd/task-service/main.go`

# Swagger 接口文档

## 生成文档

手动生成 Swagger 文档（需提前安装 swag 工具：`go install github.com/swaggo/swag/cmd/swag@latest`）：

```bash
swag init -g cmd/gateway-api/main.go -o ./docs/swagger
```

## 访问文档

服务启动后，访问以下地址即可查看并调试接口：

```bash
http://localhost:8080/swagger/index.html
```

注意：访问时若出现“invalid link”报错，请确认网关服务已正常启动，并检查访问地址正确性（核实 localhost:8080 是否正确，以及服务端口是否被修改）。

# 项目维护与贡献

## 项目维护

本项目当前由 Humphrey-He 维护，最新提交记录可查看项目 GitHub 提交历史，主要更新方向包括：完善业务逻辑、优化工程化实践、修复已知问题。

提交规范：遵循 Git 提交规范，提交信息格式为`type: 提交描述`（示例：feat: 新增价格计算服务、fix: 移除无用导入）。

## 贡献指南

欢迎各位开发者贡献代码、提出问题或建议，具体流程如下：

1. Fork 本项目；

2. 创建特性分支（git checkout -b feature/xxx）；

3. 提交代码（git commit -m "feat: 新增xxx功能"）；

4. 推送分支（git push origin feature/xxx）；

5. 提交 Pull Request，等待审核。

# 注意事项

- 确保依赖服务（MySQL、Redis、RabbitMQ）版本符合要求，避免因版本不兼容导致服务启动失败；

- 配置核心参数时，需确保连接地址、账号密码正确无误，否则会导致服务无法正常连接依赖组件；

- 启动服务时，需按顺序启动（先启动依赖服务，再启动微服务），避免因服务间依赖导致启动失败；

- 生成 Swagger 文档前，需确保 swag 工具已正确安装，否则会导致文档生成失败；

- 本项目暂未发布正式版本，可基于 main 分支进行开发与测试，后续将逐步完善版本迭代；

- 克隆项目时若出现“link dead”报错，需检查GitHub链接有效性及网络连接状态；

- 访问 Swagger 文档时若出现“invalid link”报错，需检查网关服务运行状态及访问地址正确性。
> （注：文档部分内容可能由 AI 生成）
