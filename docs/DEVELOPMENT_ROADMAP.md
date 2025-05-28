# 项目开发路线图 (MRS)

本文档概述了电影预订系统 (MRS) 的迭代式开发计划。每个阶段都有明确的目标、核心任务和预期产出，旨在逐步构建一个功能完善、高性能且可扩展的后端服务。项目将遵循 [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md) 中定义的架构进行开发。

**核心开发原则：**

*   **接口驱动设计**: 严格遵守在领域层定义接口，基础设施层实现接口的模式。
*   **分层职责明确**: 各层代码逻辑清晰，不越权。
*   **配置先行**: 所有可配置项通过统一配置管理。
*   **日志完备**: 从初期就集成结构化日志。
*   **依赖注入**: 在 `cmd/.../main.go` 中清晰地组织和注入依赖。
*   **测试驱动**: 鼓励编写单元测试和集成测试。

---

**阶段 0: 奠基与基础设施搭建**

1.  **目标**: 建立完整的项目骨架，初始化核心基础设施组件，确保基础服务可运行并具备初步的可观测性。
2.  **核心任务**:
    *   **项目结构**: 按照 `PROJECT_STRUCTURE.md` 创建完整的目录树。
    *   **Go 模块初始化**: `go mod init mrs`。
    *   **`cmd/server/main.go` 基础**:
        *   实现基础的 HTTP 服务器启动 (使用 Gin)。
        *   初步的依赖注入逻辑框架。
    *   **配置模块 (`internal/infrastructure/config`)**:
        *   实现 `loader.go` (使用 Viper) 从 `configs/app.yaml` 加载配置。
        *   定义 `model.go` 存放配置结构体 (服务端口, 数据库连接字符串, Redis 地址, JWT 密钥等)。
    *   **日志模块 (`pkg/log` 及应用)**:
        *   在 `pkg/log` 中封装结构化日志库 (如 Zap)。
        *   在 `cmd/server/main.go` 和其他关键初始化流程中集成日志记录。
    *   **数据库客户端 (`internal/infrastructure/persistence/mysql`)**:
        *   实现 `db_client.go` 初始化 GORM 并管理数据库连接池。
    *   **缓存客户端 (`internal/infrastructure/cache`)**:
        *   实现 `redis_client.go` 初始化并管理 Redis 连接。
    *   **健康检查 API (`/health`)**:
        *   定义 `internal/api/dto/response/health_response.go`。
        *   实现 `internal/api/handlers/health_handler.go`。
        *   在 `internal/api/router.go` 中注册路由。
    *   **Docker 环境**:
        *   为 Go 应用编写 `Dockerfile`。
        *   创建 `docker-compose.yml` 用于本地开发环境 (Go 服务, MySQL, Redis)。
    *   **基础 `README.md` 更新**: 填充项目概述、环境要求和初步运行指令。
3.  **预期产出**:
    *   一个可以成功编译并运行的 Go 服务，能响应 `/health` 端点。
    *   核心基础设施（数据库连接、缓存连接、配置加载、日志系统）初始化完毕。
    *   项目结构完整，为后续功能模块开发做好准备。

---

**阶段 1: 用户认证与授权模块**

1.  **目标**: 实现用户注册、登录、角色管理和基于角色的访问控制 (RBAC) 基础，确保 API 安全。
2.  **核心任务**:
    *   **领域层 (`internal/domain/user`)**:
        *   定义 `User` 和 `Role` 实体 (`user.go`)。
        *   定义 `UserRepository` 和 `RoleRepository` 接口 (`repository.go`)。
    *   **基础设施层 (`internal/infrastructure/persistence/mysql`)**:
        *   实现 `UserRepository` 和 `RoleRepository` 的 GORM 版本。
    *   **工具库 (`internal/utils`)**:
        *   实现 `hasher.go` (密码哈希与比对，例如 bcrypt)。
        *   实现 `jwt.go` (JWT 生成与校验)。
    *   **应用层 (`internal/app`)**:
        *   实现 `auth_service.go` (处理登录逻辑、生成 JWT)。
        *   实现 `user_service.go` (处理用户注册、分配角色、查询用户信息等)。服务将依赖领域仓库接口。
    *   **API 层 (`internal/api`)**:
        *   为认证和用户操作定义请求/响应 DTO (`internal/api/dto/`)。
        *   实现 `handlers/auth_handler.go` 和 `handlers/user_handler.go`。
        *   实现 `middleware/auth_middleware.go` (JWT 校验、提取用户信息到 Context)。
        *   更新 `router.go`，添加相关路由并应用认证中间件。
    *   **数据库初始化/迁移**: 创建 `users` 和 `roles` 表，植入初始管理员账户和角色数据。
3.  **预期产出**:
    *   用户可以注册账户并使用凭证登录系统。
    *   系统能够生成和验证 JWT，用于保护受限 API 端点。
    *   初步实现基于角色的权限区分。
    *   用户与认证模块遵循分层架构，职责清晰。

---

**阶段 2: 电影与放映管理模块 (管理员核心功能)**

1.  **目标**: 使管理员能够管理电影信息、类型、影厅、座位布局以及放映计划。引入基础数据缓存。
2.  **核心任务**:
    *   **领域层 (`internal/domain/movie`)**:
        *   定义 `Movie`, `Genre`, `CinemaHall`, `Seat`, `Showtime` 实体。
        *   定义相应的仓库接口 (`MovieRepository`, `GenreRepository`, `ShowtimeRepository` 等)。
    *   **基础设施层 (`internal/infrastructure`)**:
        *   `persistence/mysql/`: 实现电影、类型、影厅、座位、放映时间的仓库接口。
        *   `cache/`: 实现 `movie_cache.go`, `showtime_cache.go`，封装对电影列表/详情、场次列表等常用查询结果的 Redis 缓存操作。
    *   **应用层 (`internal/app`)**:
        *   实现 `movie_service.go`, `showtime_service.go` (包含电影、类型、影厅、座位、场次的增删改查逻辑)。
        *   在服务层实现缓存读取逻辑 (Cache-Aside) 和写操作后的缓存更新/失效策略。
    *   **API 层 (`internal/api`)**:
        *   为电影和放映时间管理功能定义 DTO。
        *   实现 `handlers/movie_handler.go`, `handlers/showtime_handler.go` (或统一到 `admin_handler.go` 下的子路由)，确保这些 API 受管理员权限保护。
        *   更新 `router.go`。
    *   **数据库初始化/迁移**: 创建相关数据表。
3.  **预期产出**:
    *   管理员可以通过 API 管理电影、类型、影厅、座位和放映计划。
    *   针对电影和场次等读取密集型数据的查询性能得到提升（通过缓存）。
    *   缓存与数据库的数据同步机制初步建立。

---

**阶段 3: 用户预订流程模块 (核心用户体验与高并发应对)**

1.  **目标**: 实现用户浏览电影、查询场次、查看座位图、选择座位、创建预订、查看和取消个人订单的完整流程。重点关注高并发场景下预订操作的数据一致性和性能。
2.  **核心任务**:
    *   **领域层 (`internal/domain/booking`, `internal/domain/movie` 可能需扩展)**:
        *   定义 `Booking`, `BookedSeat` 实体。
        *   定义 `BookingRepository`, `BookedSeatRepository` 接口。
        *   可能需要在 `Showtime` 或相关领域服务中添加获取座位状态的逻辑。
    *   **基础设施层 (`internal/infrastructure`)**:
        *   `persistence/mysql/`: 实现预订相关的仓库接口。
        *   `cache/`:
            *   增强 `showtime_cache.go` 或新建 `seat_cache.go` 管理场次座位实时/近实时可用状态 (例如使用 Redis Set, Hash, 或 Bitmap)。
            *   实现基于 Redis 的分布式锁 (`SETNX` 或 Redlock 变体) 用于座位锁定，防止并发冲突。
    *   **应用层 (`internal/app`)**:
        *   实现 `booking_service.go`:
            *   核心预订逻辑：查询座位可用性 -> (获取分布式锁) -> 尝试锁定座位 -> 创建预订记录和已预订座位记录 -> (释放锁) -> 更新缓存。
            *   处理并发预订，防止超卖。
            *   用户查询个人订单列表、订单详情、取消订单（校验业务规则）的逻辑。
        *   扩展 `movie_service.go` / `showtime_service.go` 以支持用户侧的电影查询、场次查询、座位图展示（利用缓存）。
    *   **API 层 (`internal/api`)**:
        *   为用户浏览和预订流程定义 DTO。
        *   实现/更新 `handlers/movie_handler.go`, `handlers/showtime_handler.go` (用户侧查询)。
        *   实现 `handlers/booking_handler.go` (处理预订创建、查看、取消请求)。
        *   更新 `router.go`。
    *   **数据库初始化/迁移**: 创建预订相关数据表。
3.  **预期产出**:
    *   用户可以顺畅完成从浏览电影到成功预订座位的整个流程。
    *   系统在高并发预订场景下能有效防止座位超卖，保证数据一致性。
    *   用户可以管理自己的订单。

---

**阶段 4: 管理员报告与分析模块**

1.  **目标**: 为管理员提供系统运营数据的基本报告功能。
2.  **核心任务**:
    *   **应用层 (`internal/app`)**:
        *   实现 `report_service.go`。此服务将调用相关的仓库接口（预订、场次、电影等）来聚合数据，生成报告所需的统计信息。
    *   **API 层 (`internal/api`)**:
        *   为报告功能定义请求/响应 DTO。
        *   在 `handlers/admin_handler.go` (或新建 `report_handler.go`) 中实现报告相关的 API 端点。
        *   更新 `router.go`。
3.  **预期产出**:
    *   管理员可以通过 API 获取关于预订量、上座率、收入等基本运营数据报告。

---

**阶段 5: 测试、优化与文档完善**

1.  **目标**: 全面提升系统质量、性能、稳定性和可维护性。
2.  **核心任务**:
    *   **单元测试**:
        *   为 `internal/app` 层的服务编写单元测试 (mock 领域仓库接口)。
        *   为 `internal/domain` 中复杂的业务逻辑编写单元测试。
        *   为 `internal/utils` 中的工具函数编写单元测试。
    *   **集成测试**:
        *   测试 `internal/infrastructure/persistence` 中的仓库实现与真实数据库（测试环境）的交互。
        *   测试 `internal/infrastructure/cache` 中的缓存操作与真实 Redis（测试环境）的交互。
    *   **API/端到端测试**:
        *   使用 `httptest` 包或外部工具（如 Postman + Newman, k6）测试 API 端点，验证整个请求处理链路的正确性。
    *   **性能测试与调优**:
        *   针对高并发场景（特别是座位预订、热门电影查询）进行压力测试，识别性能瓶颈。
        *   根据测试结果进行代码优化、SQL 优化、缓存策略调整等。
    *   **代码审查与重构**:
        *   组织代码审查，提升代码质量和一致性。
        *   对现有代码进行必要的重构，以提高可读性和可维护性。
    *   **日志与错误处理**:
        *   完善和标准化应用的日志记录，确保关键操作和异常均有记录。
        *   优化错误处理机制，提供更友好和明确的错误响应。
    *   **API 文档**:
        *   使用 Swagger/OpenAPI (例如通过 `swaggo/gin-swagger`) 规范生成和维护 API 文档。
    *   **安全加固**:
        *   排查常见的 Web 应用安全漏洞 (如 SQL 注入、XSS - GORM 和 Gin 会有所防护，但仍需注意业务逻辑层面，以及输入校验)。
        *   实现 API 限流策略。
3.  **预期产出**:
    *   一个经过充分测试、质量可靠、性能达标且相对稳定的 MRS 应用。
    *   一套覆盖核心功能的自动化测试用例。
    *   详细且最新的 API 文档。
    *   系统在安全性和健壮性方面得到增强。

---

**阶段 6: 架构演进与分布式能力预备 (着眼长远)**

1.  **目标**: 为系统未来的分布式演进、数据库分库分表、微服务化等做好技术和架构上的准备。此阶段的任务可根据实际需求和优先级选择性实施，并可能与前序阶段并行或后续持续进行。
2.  **核心任务 (选做与探索)**:
    *   **全局唯一ID生成器 (`internal/utils/globalid`)**:
        *   研究并实现一种全局唯一ID生成方案 (如 Snowflake 算法的变种，或基于 Redis/etcd 的发号器)。
        *   规划在需要分片的表或未来微服务中的实体ID上应用此方案。
    *   **配置中心集成 (`internal/infrastructure/config`)**:
        *   调整 `loader.go`，使其支持从配置中心 (如 Consul, etcd, Nacos, Apollo) 加载和动态更新配置。
    *   **异步任务处理与消息队列 (`cmd/worker`, `internal/infrastructure/messagequeue`)**:
        *   选择并集成一个消息队列中间件 (如 NATS, Kafka, RabbitMQ)。
        *   将一些非核心、耗时的操作（如订单成功后发送邮件/短信通知）改造为通过消息队列异步处理。
        *   实现 `cmd/worker/main.go` 作为异步任务的消费者服务。
    *   **服务注册与发现 (`internal/infrastructure/discovery`)**:
        *   集成服务发现客户端 (如 Consul, etcd, Nacos)。
        *   将主 API 服务注册到服务发现中心。
    *   **分布式追踪 (`internal/api/middleware/tracing_middleware.go`, `pkg/tracing`)**:
        *   集成 OpenTelemetry SDK，实现分布式追踪的上下文传递和数据上报 (例如上报到 Jaeger, Zipkin)。
    *   **数据库分库分表策略研究与原型验证 (`internal/infrastructure/persistence/sharding`)**:
        *   针对数据量可能巨大的核心表 (如 `Booking`, `BookedSeat`)，研究分片键的选择和分片策略。
        *   在 `infrastructure` 层对某个仓库的实现进行改造，尝试引入分片路由逻辑 (例如使用 `database/sql` + 自定义路由，或调研 ShardingSphere-Go, Vitess 等方案)。
    *   **RPC 通信探索 (`cmd/rpc`, 领域服务接口化)**:
        *   为某个边界清晰的领域模块 (如 `Auth` 或 `User` 管理) 定义 gRPC 服务接口。
        *   创建对应的 `cmd/rpc/[service_name]/main.go` 作为其独立 RPC 服务的潜在入口。
    *   **API 网关评估**:
        *   随着服务数量增加，评估引入 API 网关的必要性和选型。
    *   **可观测性体系完善**:
        *   建立更完善的 Metrics 监控 (例如 Prometheus + Grafana)，结合已有的 Logging 和 Tracing，形成完整的可观测性解决方案。
3.  **预期产出**:
    *   项目具备了向分布式架构演进的技术基础和部分实践经验。
    *   关键模块为未来的独立部署和服务化做好了准备。
    *   团队对分布式系统相关的技术栈和挑战有了更深入的理解。

---

**贯穿所有阶段的注意事项：**

*   **版本控制**: 严格使用 Git 进行版本控制，遵循合适的分支模型 (如 Git Flow 或 GitHub Flow)，编写清晰的 Commit Message。
*   **代码审查**: 所有主要的代码变更都应经过至少一位其他开发者的审查。
*   **持续集成/持续部署 (CI/CD)**: 尽早搭建 CI/CD 流水线，自动化构建、测试和部署流程。
*   **文档同步**: 随着开发的进行，及时更新 `README.md`, `PROJECT_STRUCTURE.md`, `DATA_MODEL.md`, `DEVELOPMENT_ROADMAP.md` 以及 API 文档。
*   **迭代与反馈**: 每个阶段结束后进行回顾，根据实际情况和反馈调整后续计划。敏捷开发，小步快跑。

此开发路线图为 MRS 项目的成功实施提供了一个结构化的指引。在实际开发过程中，应保持灵活性，并根据具体的技术选型和业务发展进行调整。
