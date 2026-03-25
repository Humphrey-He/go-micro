Go Microservice Practical Project (English Documentation)

Project Introduction

This project is a practical e-commerce order fulfillment microservice project developed based on the Go language. It strictly follows real enterprise-level development processes and engineering standards, simulating the entire business process from user order placement, inventory deduction, order fulfillment to task scheduling in e-commerce scenarios. The core goal of the project is to cover high-frequency interview points and engineering practice essentials for backend development, helping developers master core capabilities such as microservice architecture design, distributed system problem solving, and high-availability architecture construction. It can be directly used as a core personal interview project to quickly improve job-hunting competitiveness.

Project Address: https://github.com/Humphrey-He/go-micro

Project Background and Goals

Project Background

With the rapid development of the e-commerce industry, high concurrency, high availability, and data consistency have become core requirements for backend development. Taking "e-commerce order fulfillment" as the core business scenario, this project decomposes the core processes in real businesses, adopts a microservice architecture model, solves engineering pain points such as service communication, data consistency, cache stability, and observability in distributed scenarios, simulates real team delivery processes, and builds a practical project that can be directly implemented and expanded.

Core Goals

- Cover the entire microservice development process: fully restore enterprise-level development scenarios from service splitting, interface design to deployment and operation and maintenance;

- Conquer high-frequency interview points: focus on core interview points such as service splitting, RPC communication, message queue application, data consistency, and cache stability;

- Implement engineering practices: standardize code structure, improve unit test coverage, achieve observability, and enhance project maintainability and scalability;

- Build an interview-oriented practical project: with prominent project highlights and mainstream technology stack, it can be directly used as a core personal project for job interviews.

Directory Structure

The project directory strictly follows Go engineering standards, with a clear structure and distinct responsibilities, facilitating development, maintenance and expansion. The details are as follows:

cmd/          # Entry points for each microservice (gateway, order, inventory, etc.)
internal/     # Core business logic (business implementation and data processing of each service)
pkg/          # Public libraries and toolkits (general tools, middleware, public logic)
proto/        # gRPC protocol definition and automatically generated code (dependency for inter-service RPC communication)
deploy/       # Deployment-related resources (database scripts, deployment configurations)
docs/         # Project documents (interface documents, instruction documents, Swagger resources)
.github/      # CI/CD configuration (GitHub Actions workflow)
.gitignore    # Git ignore file configuration
go.mod        # Go module dependency configuration
go.sum        # Go dependency version locking file
README.md     # Project English documentation
README_ZH.md  # Project Chinese documentation
README_EN.md  # Supplementary English documentation for the project

Technology Stack

The project adopts a mainstream and mature technology stack, which is suitable for actual enterprise development scenarios, balancing performance, reliability and maintainability. The specific selection is as follows:

- Development Language: Go (concise and efficient, natively supports high concurrency, suitable for microservice development)

- Web Framework: Gin (lightweight and high-performance, used for API development of gateway services)

- RPC Communication: gRPC (strong interface constraints, low latency, easy governance, used for internal communication between services)

- Database: MySQL 8 (relational database, stores core business data and ensures data consistency)

- Cache: Redis (high-performance cache, used to reduce database pressure and improve interface response speed)

- Message Queue: RabbitMQ (supports DLQ (Dead-Letter Queue) to achieve reliable message delivery and consumption)

- Configuration Management: Environment variables + unified configuration loading (concise and flexible, adaptable to different deployment environments)

- Logging System: Zap (high-performance, structured logging, facilitating problem troubleshooting and log analysis)

- API Documentation: Swagger (based on swag + gin-swagger, automatically generates interface documents for easy interface debugging)

- CI/CD: GitHub Actions (implements automated building, testing, and improves development and delivery efficiency)

Core Project Highlights

Different from ordinary practice projects, this project focuses on enterprise-level engineering practices and high-frequency interview pain points. The core highlights are as follows:

1. Clear and Reasonable Microservice Splitting: Services are split strictly according to business responsibilities. The five services (gateway, order, inventory, user, and task orchestration) perform their respective duties, with low coupling and high cohesion, complying with microservice design principles and can be directly referenced for enterprise-level project splitting.

2. Standardized Internal Communication: gRPC is used to realize inter-service communication, and standardized proto protocols are defined to achieve strong interface constraints, reduce coupling between services, and ensure low-latency and easy-to-govern communication, which is consistent with actual enterprise microservice communication.

3. Data Consistency Guarantee: The Outbox pattern is adopted to implement the same transaction for order writing and message writing, and asynchronous delivery ensures eventual data consistency in distributed scenarios, solving core pain points such as "order success but no inventory deduction" and "message loss" in e-commerce scenarios.

4. Reliable Message Consumption Mechanism: Based on RabbitMQ, a manual ACK confirmation mechanism is implemented, combined with DLQ (Dead-Letter Queue) to handle message consumption failure scenarios, avoid message loss and duplicate consumption, and ensure the reliability and stability of message consumption.

5. Cache Stability Design: Aiming at high-frequency problems such as cache penetration and cache avalanche, schemes such as null value caching and TTL jitter optimization are implemented to improve cache system stability and avoid service avalanche caused by cache abnormalities, meeting the needs of high-concurrency scenarios.

6. Unified Error Code Specification: Define consistent error codes and error semantics across services, simplify problem troubleshooting processes, improve collaboration efficiency between services, and comply with error handling specifications for enterprise-level projects.

7. Strong Testability: Core logic (order, inventory, cache) is covered by unit tests, standardizing the testing process, improving code quality, and facilitating subsequent function iteration and problem troubleshooting, reflecting engineering development thinking.

8. Improved Engineering Standards: Follow Go engineering standards, with clear directory structure and unified code style, integrate CI/CD processes and automatic interface document generation, which is consistent with actual enterprise development and delivery standards.

Core Business Logic

Aggregate View Status Mapping

The project provides an aggregate interface GET /api/v1/order-views/{order_no} to return the main status and detailed status of the order. view_status is a unified status for callers. When there is a conflict between underlying statuses (order status, task status), it is mapped according to the following priority rules (from high to low):

1. When order_status == CANCELED and task_type == TIMEOUT_CANCEL, view_status = TIMEOUT (order canceled due to timeout);

2. When order_status == CANCELED, view_status = CANCELED (order actively canceled);

3. When task_status == DEAD, view_status = DEAD (task execution failed and cannot be retried);

4. When task_status == FAILED, view_status = FAILED (task execution failed, can be retried);

5. When order_status == SUCCESS, view_status = SUCCESS (order fulfillment successful);

6. When task_status == RUNNING, view_status = PROCESSING (order being fulfilled);

7. When order_status == RESERVED and task_status in (PENDING, NOT_FOUND), view_status = PENDING (order pending fulfillment).

Return Structure Example

{
  "order_no": "BIZ-xxxx",
  "view_status": "CANCELED",
  "order_status": "CANCELED",
  "task_status": "DEAD",
  "reservation_status": "RELEASED",
  "cancel_reason": "timeout"
}

Status Mapping Example Table

order_status  task_status  reservation_status  view_status
RESERVED      PENDING      RESERVED            PENDING
PROCESSING    RUNNING      RESERVED            PROCESSING
SUCCESS       SUCCESS      CONFIRMED           SUCCESS
CANCELED      FAILED       RELEASED            CANCELED
CANCELED      DEAD         RELEASED            TIMEOUT/DEAD

Quick Start

Running Dependencies

Please ensure that the following dependencies are installed locally or on the server, and the versions meet the requirements:

- MySQL 8 (core business data storage)

- Redis (cache service)

- RabbitMQ (message queue, need to support DLQ Dead-Letter Queue)

- Go 1.18+ (project development language)

Deployment Steps

8. Clone the Projectgit clone https://github.com/Humphrey-He/go-micro.git
cd go-microNote: If you encounter the "link dead" error when cloning, please check the validity of the GitHub link and ensure a stable network connection.

9. Initialize the DatabaseExecute the deployment script to initialize the database table structure:mysql -u username -p password < deploy/sql/schema.sql

10. Configure Core ParametersSet environment variables and configure core dependency addresses (you can also modify the configuration reading logic as needed):

  - MYSQL_DSN: MySQL connection address (format: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local)

  - REDIS_ADDR: Redis connection address (format: host:port)

  - MQ_URL: RabbitMQ connection address (format: amqp://user:password@host:port/)

  - ORDER_GRPC_ADDR: Order service gRPC address

  - INVENTORY_GRPC_ADDR: Inventory service gRPC address

11. Start Dependent ServicesStart MySQL, Redis, and RabbitMQ, ensure the services are running normally, and verify the connection through the corresponding client.

12. Start MicroservicesStart the following services in sequence (can be executed in different terminals):# Start gateway service
go run cmd/gateway-api/main.go

# Start order service
go run cmd/order-service/main.go

# Start inventory service
go run cmd/inventory-service/main.go

# Start user service
go run cmd/user-service/main.go

# Start task service
go run cmd/task-service/main.go

Swagger API Documentation

Generate Documentation

Manually generate Swagger documentation (need to install the swag tool in advance: go install github.com/swaggo/swag/cmd/swag@latest):

swag init -g cmd/gateway-api/main.go -o ./docs/swagger

Access Documentation

After the service is started, access the following address to view and debug the interface:

http://localhost:8080/swagger/index.html

Note: If you encounter the "invalid link" error when accessing, please confirm that the gateway service is running and the access address is correct (check if localhost:8080 is correct and if the service port has been modified).

Project Maintenance and Contribution

Project Maintenance

The project is currently maintained by Humphrey-He. The latest commit records can be viewed in the project's GitHub commit history. The main update directions are: improving business logic, optimizing engineering practices, and fixing known issues.

Commit Specification: Follow Git commit specifications, and the commit message format is type: commit description (e.g., feat: add price calculation service, fix: remove unused imports).

Contribution Guidelines

Developers are welcome to contribute code, raise questions or suggestions. The process is as follows:

1. Fork this project;

2. Create a feature branch (git checkout -b feature/xxx);

3. Commit code (git commit -m "feat: add xxx function");

4. Push the branch (git push origin feature/xxx);

5. Submit a Pull Request and wait for review.

Notes

- Ensure that the versions of dependent services (MySQL, Redis, RabbitMQ) meet the requirements to avoid service startup failure due to version incompatibility;

- When configuring core parameters, ensure that the connection address, account and password are correct, otherwise the service will not be able to connect to the dependencies normally;

- When starting services, start them in order (first start dependent services, then start microservices) to avoid service startup failure due to inter-service dependencies;

- Before generating Swagger documentation, ensure that the swag tool is installed correctly, otherwise the execution will fail;

- The project has not yet released an official version. You can develop and test based on the main branch, and version iteration will be gradually improved in the future;

- If you encounter "link dead" when cloning the project, check the GitHub link validity and network connection;

- If you encounter "invalid link" when accessing Swagger, check whether the gateway service is started and the access address is correct.
