
# 项目模块化设计 (MRS Go 服务)

本文档详细描述了电影预订系统 (MRS) 后端服务的 Go 项目结构和模块化设计。项目遵循分层架构和高内聚低耦合的原则，以便于开发、维护和未来的扩展。

## 顶层目录结构

```
mrs/
├── cmd/                        # 可执行文件入口点
├── internal/                   # 项目内部私有代码
├── pkg/                        # 可被外部项目导入的公共库代码 (可选)
├── configs/                    # 应用程序配置文件
├── scripts/                    # 构建、部署、工具类脚本
├── go.mod                      # Go 模块定义文件
├── go.sum                      # Go 模块校验和文件
├── Dockerfile                  # 服务构建 Dockerfile
├── docker-compose.yml          # 本地开发环境编排文件
└── README.md                   # 项目主 README
```

## `cmd/` - 命令入口

此目录存放项目所有可执行应用的 `main` 包。

*   **`cmd/server/main.go`**:
    *   **作用**: HTTP API 服务的主入口。
    *   **职责**: 初始化应用依赖（配置、日志、数据库、缓存、应用服务等），设置 HTTP 路由，并启动 HTTP 服务器。
*   **`cmd/worker/main.go` (未来规划)**:
    *   **作用**: 异步任务处理服务的入口。
    *   **职责**: 启动消费者，监听消息队列中的任务（例如，订单成功后的通知发送、定时报表生成），并执行相应的处理逻辑。
*   **`cmd/rpc/[service_name]/main.go` (未来规划)**:
    *   **作用**: 如果特定领域模块被拆分为独立的微服务，则此处为其 RPC 服务入口。
    *   **职责**: 启动 gRPC 服务器，注册并提供该微服务的 RPC 接口。

## `internal/` - 内部业务逻辑

此目录包含项目核心的业务逻辑和内部实现，根据 Go 的约定，`internal` 下的包不能被其他外部项目导入。

### `internal/api/` - API 接口层

负责处理所有面向外部的 HTTP 请求和响应。

*   **`internal/api/handlers/`**:
    *   存放具体的 HTTP 请求处理器 (Handlers)，例如 `auth_handler.go`, `movie_handler.go`。
    *   职责：解析请求、校验参数、调用 `internal/app` 层的应用服务处理业务，构建并返回 HTTP 响应。Handler 应保持轻量，不包含复杂业务逻辑。
*   **`internal/api/middleware/`**:
    *   存放 HTTP 中间件，例如 `auth_middleware.go` (身份认证)、`logging_middleware.go` (请求日志)、`tracing_middleware.go` (分布式追踪上下文传递)。
    *   职责：处理跨多个请求的通用横切关注点。
*   **`internal/api/router.go`**:
    *   定义 API 的路由规则，将 URL 路径映射到对应的 Handler。通常使用 Web 框架 (如 Gin) 的路由功能。
*   **`internal/api/dto/`**: (Data Transfer Objects)
    *   存放 API 请求体 (request) 和响应体 (response) 的数据结构定义。
    *   `request/`: 例如 `user_creation_request.go`。
    *   `response/`: 例如 `movie_details_response.go`。
    *   DTO 与领域模型 (`internal/domain`) 分离，用于适配 API 接口的数据格式。

### `internal/app/` - 应用层

编排和协调业务流程，实现具体的应用用例。

*   例如 `auth_service.go`, `user_service.go`, `movie_service.go`, `booking_service.go`。
*   **职责**:
    *   调用 `internal/domain` 层的领域服务或仓库接口来执行业务操作。
    *   处理事务边界（如果跨多个领域操作）。
    *   不包含具体的业务规则（这些应在领域层），主要负责流程控制。
    *   在分布式系统中，可能需要调用其他微服务的客户端。

### `internal/domain/` - 领域层

项目的核心，包含业务逻辑的根本。

*   **`internal/domain/[entity_name]/`**: 按领域实体或聚合根组织，例如 `user/`, `movie/`, `booking/`。
    *   **`[entity_name].go`**: 定义领域实体 (Entity) 和值对象 (Value Object)，例如 `user/user.go`。包含实体的属性和不涉及外部依赖的业务方法。
    *   **`repository.go`**: 定义该领域实体的仓库接口 (Repository Interface)，例如 `user/repository.go`。此接口约定了数据的持久化和检索操作，由基础设施层实现。
    *   **`service.go` (可选)**: 定义领域服务。如果某些业务逻辑不适合放在单个实体中，而是跨多个实体或需要协调，则可定义领域服务。
*   **`internal/domain/common/`**:
    *   存放跨领域的通用定义，例如基础实体 ID 类型、通用的值对象、领域特定的错误定义等。

### `internal/infrastructure/` - 基础设施层

提供与外部系统和技术细节交互的具体实现。

*   **`internal/infrastructure/persistence/`**: 数据持久化相关实现。
    *   `mysql/`: 针对 MySQL 的具体实现。
        *   `db_client.go`: 初始化数据库连接 (例如 GORM 客户端)。
        *   `[entity_name]_repository.go`: 实现 `internal/domain/[entity_name]/repository.go` 中定义的仓库接口，例如 `mysql/user_repository.go`。
    *   `sharding/` (未来规划): 分库分表逻辑封装，例如分片键的路由规则、分片连接管理。
    *   `migrations/`: 数据库表结构迁移脚本或代码。
*   **`internal/infrastructure/cache/`**: 缓存服务实现。
    *   `redis_client.go`: 初始化 Redis 连接。
    *   `[feature]_cache.go`: 针对特定功能的缓存操作封装，例如 `movie_cache.go`。
*   **`internal/infrastructure/messagequeue/` (未来规划)**: 消息队列的生产者和消费者实现。
    *   例如 `kafka_producer.go`, `nats_consumer.go`。
*   **`internal/infrastructure/discovery/` (未来规划)**: 服务发现客户端实现。
    *   例如 `consul_client.go`。
*   **`internal/infrastructure/config/`**: 配置加载与管理。
    *   `loader.go`: 实现从文件 (如 YAML, JSON) 或配置中心加载配置的逻辑 (例如使用 Viper)。
    *   `model.go`: 定义配置项的结构体。

### `internal/utils/` - 通用工具库

存放项目内部使用的、与具体业务领域无关的通用辅助函数。

*   例如 `hasher.go` (密码哈希)、`jwt.go` (JWT 工具)、`validator.go` (通用数据校验)、`pagination.go` (分页逻辑辅助)、`globalid/` (未来规划的全局唯一ID生成器)。

## `pkg/` - 公共代码库 (可选)

此目录存放可以被其他外部项目安全导入的公共库代码。如果项目初期没有这样的需求，此目录可以省略。

*   例如 `pkg/apierrors/` (标准化的 API 错误定义)、`pkg/log/` (对标准日志库的封装，提供统一接口)、`pkg/tracing/` (分布式追踪的通用封装)。

## `configs/` - 配置文件

存放应用的配置文件。

*   例如 `app.yaml`, `app.development.yaml`, `app.production.yaml`。
*   应包含数据库连接信息、Redis 地址、JWT 密钥、服务端口等。
*   敏感信息不应直接硬编码在此，推荐使用环境变量或专门的密钥管理服务。

## `scripts/` - 脚本文件

存放用于项目构建、部署、数据库迁移、代码生成等辅助功能的脚本。

*   例如 `build.sh`, `deploy.sh`, `run_migrations.sh`。

---
