
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
    *   **功能**:
        1.  **初始化配置**: 加载 `configs/app.yaml` (通过 `internal/infrastructure/config/loader.go`)。
        2.  **初始化日志**: 设置全局日志器 (通过 `pkg/log` 或直接使用 `zap`)。
        3.  **初始化数据库连接**: 调用 `internal/infrastructure/persistence/mysql/db_client.go` 中的函数获取数据库实例 (e.g., `*gorm.DB`)。
        4.  **初始化缓存连接**: 调用 `internal/infrastructure/cache/redis_client.go` 中的函数获取 Redis 客户端实例。
        5.  **依赖注入 (DI)**:
            *   创建各基础设施层仓库的实例 (e.g., `NewMysqlUserRepository(db)`).
            *   创建各应用层服务的实例, 并将仓库实例作为依赖注入 (e.g., `NewUserService(userRepo, passHasher)`).
            *   创建各 API 层处理器的实例, 并将应用服务实例作为依赖注入 (e.g., `NewAuthHandler(authSvc, userSvc)`).
        6.  **初始化 Web 框架 (Gin)**: 创建 Gin 引擎实例。
        7.  **注册中间件**: 向 Gin 引擎注册全局中间件 (如日志、CORS、Panic恢复)。
        8.  **注册路由**: 调用 `internal/api/router.go` 中的函数设置 API 路由，并将处理器实例传递给路由设置函数。
        9.  **启动 HTTP 服务器**: 监听指定端口并启动服务。
        10.  **优雅关闭 (Graceful Shutdown)**: 实现信号监听，以便在收到终止信号时平滑关闭服务 (关闭数据库连接、Redis连接，等待进行中的请求完成等)。

*   **`cmd/worker/main.go` (未来规划)**:
    *   **功能**:
        1.  初始化配置、日志、数据库连接、缓存连接 (与 `cmd/server/main.go` 类似，但可能只需要部分依赖)。
        2.  依赖注入：创建消息队列消费者、相关应用服务实例。
        3.  启动消息队列消费者，开始监听和处理消息。
        4.  实现优雅关闭。

*   **`cmd/rpc/[service_name]/main.go` (未来规划)**:
    *   **功能**:
        1.  初始化配置、日志、数据库连接等 (根据该 RPC 服务所需依赖)。
        2.  依赖注入：创建该 RPC 服务的实现实例 (通常是实现了 gRPC 生成的接口的服务)。
        3.  创建并启动 gRPC 服务器，注册服务。
        4.  实现优雅关闭。

## `internal/` - 内部业务逻辑

### `internal/api/` - API 接口层

*   **`internal/api/handlers/`**:
    *   **`auth_handler.go`**:
        *   包含 `AuthHandler` 结构体 (通常包含 `AuthService` 和 `UserService` 依赖)。
        *   实现 `Register(c *gin.Context)`: 处理用户注册请求，调用 `UserService.Register`。
        *   实现 `Login(c *gin.Context)`: 处理用户登录请求，调用 `AuthService.Login`。
        *   实现 `Logout(c *gin.Context)`: (可选) 处理用户登出请求，可能涉及 JWT 黑名单。
    *   **`user_handler.go`**:
        *   包含 `UserHandler` 结构体 (通常包含 `UserService` 依赖)。
        *   实现 `GetProfile(c *gin.Context)`: 获取当前登录用户的信息，从 Context 中获取用户ID，调用 `UserService.GetProfile`。
        *   实现 `UpdateProfile(c *gin.Context)`: 更新用户信息。
        *   (Admin) 实现 `PromoteUserToAdmin(c *gin.Context)`: 提升用户角色。
    *   **`movie_handler.go`**:
        *   包含 `MovieHandler` 结构体 (包含 `MovieService`, `ShowtimeService` 依赖)。
        *   (Public) 实现 `ListMovies(c *gin.Context)`: 获取电影列表（分页、筛选）。
        *   (Public) 实现 `GetMovieByID(c *gin.Context)`: 获取特定电影详情。
        *   (Public) 实现 `ListGenres(c *gin.Context)`: 获取电影类型列表。
        *   (Admin) 实现 `CreateMovie(c *gin.Context)`: 添加新电影。
        *   (Admin) 实现 `UpdateMovie(c *gin.Context)`: 更新电影信息。
        *   (Admin) 实现 `DeleteMovie(c *gin.Context)`: 删除电影。
    *   **`showtime_handler.go`**:
        *   包含 `ShowtimeHandler` 结构体 (包含 `ShowtimeService`, `MovieService` 依赖)。
        *   (Public) 实现 `ListShowtimesByDate(c *gin.Context)`: 获取特定日期的放映场次。
        *   (Public) 实现 `GetShowtimeSeats(c *gin.Context)`: 获取特定场次的座位图及可用状态。
        *   (Admin) 实现 `CreateShowtime(c *gin.Context)`: 添加放映时间。
        *   (Admin) 实现 `UpdateShowtime(c *gin.Context)`: 更新放映时间。
        *   (Admin) 实现 `DeleteShowtime(c *gin.Context)`: 删除放映时间。
    *   **`booking_handler.go`**:
        *   包含 `BookingHandler` 结构体 (包含 `BookingService` 依赖)。
        *   实现 `CreateBooking(c *gin.Context)`: 用户创建新预订。
        *   实现 `GetMyBookings(c *gin.Context)`: 用户查看自己的预订列表。
        *   实现 `GetBookingByID(c *gin.Context)`: 用户查看特定预订详情。
        *   实现 `CancelBooking(c *gin.Context)`: 用户取消预订。
    *   **`admin_handler.go` (或分散到对应模块的Admin方法)**:
        *   包含 `AdminHandler` 结构体 (包含多种 Admin 相关服务依赖，如 `ReportService`, `UserService` 等)。
        *   实现 `ListAllBookings(c *gin.Context)`: 管理员查看所有预订。
        *   实现 `GetSystemReport(c *gin.Context)`: 获取系统报告 (如收入、上座率)。
    *   **`health_handler.go` (通常会有一个)**:
        *   实现 `CheckHealth(c *gin.Context)`: 返回系统健康状态。

*   **`internal/api/middleware/`**:
    *   **`auth_middleware.go`**:
        *   实现 `JWTMiddleware()`: 校验请求中的 JWT，如果有效则将用户信息 (如用户ID、角色)存入 `gin.Context`，否则返回未授权错误。
        *   实现 `AdminRequiredMiddleware()`: 检查用户是否为管理员角色。
    *   **`logging_middleware.go`**:
        *   实现 `LoggingMiddleware()`: 记录每个请求的详细信息（方法、路径、状态码、耗时、User-Agent等）。
    *   **`cors_middleware.go` (如果需要)**:
        *   实现 `CORSMiddleware()`: 配置跨域资源共享策略。
    *   **`panic_recovery_middleware.go` (通常Gin自带或可自定义增强)**:
        *   捕获 panic，记录错误，并返回统一的服务器内部错误响应。
    *   **`tracing_middleware.go` (未来规划)**:
        *   实现 `TracingMiddleware()`: 从请求头提取或生成 trace ID，并将其注入到后续的日志和跨服务调用中。

*   **`internal/api/router.go`**:
    *   **功能**:
        *   定义一个或多个函数 (e.g., `SetupRouter(engine *gin.Engine, authHandler *AuthHandler, ...)`).
        *   在此函数中，使用 `engine.GET`, `engine.POST` 等方法定义所有 API 路由。
        *   将路由分组 (e.g., `/api/v1`, `/admin`)。
        *   为特定路由或路由组应用中间件。

*   **`internal/api/dto/`**:
    *   **`request/[operation]_request.go`**:
        *   定义对应 API 操作的请求体结构体，包含 JSON 标签和验证标签 (e.g., `binding:"required"`).
        *   例如: `user_creation_request.go`, `login_request.go`, `create_movie_request.go`.
    *   **`response/[resource]_response.go`**:
        *   定义对应 API 操作的响应体结构体，包含 JSON 标签。
        *   例如: `user_profile_response.go`, `movie_details_response.go`, `booking_summary_response.go`.
        *   可能包含用于将领域模型转换为 DTO 的辅助函数。

### `internal/app/` - 应用层

*   **`auth_service.go`**:
    *   包含 `AuthService` 接口和其实现 `authServiceImpl` 结构体 (依赖 `UserRepository`, `PasswordHasher`, `JWTGenerator`)。
    *   实现 `Login(ctx context.Context, username, password string) (string, error)`: 校验用户凭证，成功则生成并返回 JWT。
*   **`user_service.go`**:
    *   包含 `UserService` 接口和其实现 `userServiceImpl` 结构体 (依赖 `UserRepository`, `RoleRepository`, `PasswordHasher`)。
    *   实现 `Register(ctx context.Context, req *dto.UserCreationRequest) (*domain.User, error)`: 创建新用户，哈希密码，分配默认角色。
    *   实现 `GetProfile(ctx context.Context, userID uint) (*domain.User, error)`: 根据用户ID获取用户信息。
    *   实现 `AssignRole(ctx context.Context, userID uint, roleName string) error`: 给用户分配角色。
*   **`movie_service.go`**:
    *   包含 `MovieService` 接口和其实现 `movieServiceImpl` 结构体 (依赖 `MovieRepository`, `GenreRepository`, `MovieCache`)。
    *   实现电影相关的 CRUD 业务逻辑 (创建、更新、删除、按ID获取、列表查询 - 带缓存逻辑)。
    *   实现类型相关的业务逻辑。
*   **`showtime_service.go`**:
    *   包含 `ShowtimeService` 接口和其实现 `showtimeServiceImpl` 结构体 (依赖 `ShowtimeRepository`, `SeatRepository` (如果座位管理复杂), `ShowtimeCache`)。
    *   实现放映场次相关的 CRUD 业务逻辑。
    *   实现获取场次座位图及可用状态的逻辑（可能涉及与缓存的复杂交互）。
*   **`booking_service.go`**:
    *   包含 `BookingService` 接口和其实现 `bookingServiceImpl` 结构体 (依赖 `BookingRepository`, `BookedSeatRepository`, `ShowtimeRepository`, `SeatRepository`, `DistributedLockProvider` (来自cache), `UserPointService` (如果涉及积分等))。
    *   实现核心预订流程: 检查库存/座位 -> (尝试获取分布式锁) -> 创建订单 -> 扣减库存/标记座位 -> (释放锁) -> (可选：发送消息到队列)。
    *   实现用户查询和取消订单的逻辑。
*   **`report_service.go`**:
    *   包含 `ReportService` 接口和其实现 `reportServiceImpl` 结构体 (依赖多种 Repository)。
    *   实现生成各种业务报告的逻辑 (如上座率、收入统计)。

### `internal/domain/` - 领域层

*   **`internal/domain/[entity_name]/[entity_name].go`**:
    *   例如 `user/user.go`, `movie/movie.go`, `booking/booking.go`。
    *   **功能**:
        *   定义领域实体 (Entity) 结构体，包含其属性和 ID。
        *   定义与实体相关的业务方法（不依赖外部服务的纯粹业务逻辑，例如 `User.ChangePassword()`, `Booking.CalculateTotalPrice()`, `Showtime.IsBookingAllowed()`）。
        *   定义值对象 (Value Object) 结构体（如果适用，例如 `Money`, `Address`）。
*   **`internal/domain/[entity_name]/repository.go`**:
    *   例如 `user/repository.go`, `movie/repository.go`, `booking/repository.go`。
    *   **功能**:
        *   定义该领域实体的仓库接口 (Repository Interface)。
        *   接口方法应体现业务意图，例如 `UserRepository.FindByID(ctx context.Context, id uint) (*User, error)`, `MovieRepository.Save(ctx context.Context, movie *Movie) error`。
        *   定义可能的数据查找选项结构体 (e.g., `UserQueryOptions`)。
*   **`internal/domain/[entity_name]/service.go` (可选)**:
    *   例如 `booking/discount_calculation_service.go` (如果折扣逻辑复杂且跨多个实体)。
    *   **功能**:
        *   定义领域服务接口及其实现。
        *   封装那些不属于任何单个实体，但属于该领域的核心业务规则或协调逻辑。

*   **`internal/domain/common/`**:
    *   **`entity.go`**: (可选) 定义通用的基础实体结构或接口 (如包含 `ID`, `CreatedAt`, `UpdatedAt` 的 `BaseEntity`)。
    *   **`value_object.go`**: (可选) 定义可跨领域复用的值对象。
    *   **`error.go`**: 定义领域层特定的错误类型或常量 (e.g., `ErrUserNotFound`, `ErrInsufficientStock`)。

### `internal/infrastructure/` - 基础设施层

*   **`internal/infrastructure/persistence/mysql/`**:
    *   **`db_client.go`**:
        *   提供函数 (e.g., `NewMySQLConnection(cfg *config.DBConfig) (*gorm.DB, error)`) 来建立和配置数据库连接 (使用 GORM)。
        *   可能包含数据库连接池配置。
    *   **`user_repository.go`**:
        *   实现 `domain/user/repository.go` 中定义的 `UserRepository` 接口，使用 GORM 操作 `users` 表。
    *   **`movie_repository.go`**: 实现 `MovieRepository` 接口。
    *   **`booking_repository.go`**: 实现 `BookingRepository` 接口。
    *   ... (其他实体的 Repository 实现) ...
*   **`internal/infrastructure/persistence/sharding/` (未来规划)**:
    *   **`sharding_manager.go`**: 管理分片数据库连接，提供获取特定分片键对应连接的逻辑。
    *   **`sharding_rule.go`**: 定义分片规则 (e.g., 用户ID取模)。
*   **`internal/infrastructure/persistence/migrations/`**:
    *   存放数据库迁移文件 (e.g., 使用 `golang-migrate/migrate` 或 GORM 的 AutoMigrate 功能的脚本/代码)。

*   **`internal/infrastructure/cache/`**:
    *   **`redis_client.go`**:
        *   提供函数 (e.g., `NewRedisClient(cfg *config.RedisConfig) (*redis.Client, error)`) 来建立和配置 Redis 连接。
    *   **`movie_cache.go`**:
        *   (可选，如果缓存逻辑复杂) 封装与电影相关的缓存操作，例如 `GetMovieFromCache`, `SetMovieToCache`, `InvalidateMovieCache`。可能直接被应用层服务调用，或者应用层服务自己实现缓存逻辑。
    *   **`showtime_cache.go`**: 类似地封装与场次相关的缓存。
    *   **`distributed_lock.go` (如果需要抽象)**:
        *   (可选) 提供分布式锁的接口和 Redis 实现 (e.g., `AcquireLock`, `ReleaseLock`)。

*   **`internal/infrastructure/messagequeue/` (未来规划)**:
    *   **`kafka_producer.go` / `nats_producer.go`**: 实现消息生产者，提供发送消息到特定主题的方法。
    *   **`kafka_consumer.go` / `nats_consumer.go`**: 实现消息消费者，订阅主题并处理接收到的消息。

*   **`internal/infrastructure/discovery/` (未来规划)**:
    *   **`consul_client.go` / `etcd_client.go`**: 实现服务注册、注销、发现的客户端逻辑。

*   **`internal/infrastructure/config/`**:
    *   **`loader.go`**:
        *   使用 Viper 或类似库从文件 (YAML/JSON/ENV) 加载配置到 `model.go` 中定义的结构体。
        *   提供获取配置实例的函数。
    *   **`model.go`**:
        *   定义与 `configs/app.yaml` 结构对应的 Go 结构体，用于强类型访问配置项。

### `internal/utils/` - 通用工具库

*   **`hasher.go`**: 提供密码哈希 (e.g., bcrypt) 和比对的函数。
*   **`jwt.go`**: 提供 JWT 生成、解析和校验的函数。
*   **`validator.go`**: (可选) 封装通用的数据校验逻辑，或集成 `go-playground/validator` 的自定义校验规则。
*   **`pagination.go`**: 提供处理分页请求参数和生成分页响应信息的辅助函数。
*   **`error_utils.go` (或在 `common/error.go` 中)**: 提供统一的错误处理辅助函数，例如将 error 转换为 HTTP 状态码。
*   **`globalid/generator.go` (未来规划)**: 实现全局唯一 ID 生成算法。
*   **`time_utils.go` (如果常用)**: 封装常用的时间处理函数。

## `pkg/` - 公共代码库 (可选)

*   **`pkg/apierrors/errors.go`**: 定义可被外部（例如其他微服务或客户端生成代码）使用的标准化 API 错误结构体和常量。
*   **`pkg/log/logger.go`**:
    *   提供一个日志接口 (e.g., `Logger`) 和一个或多个具体实现 (e.g., 基于 Zap 的 `ZapLogger`)。
    *   使得应用代码可以解耦具体的日志库。
*   **`pkg/tracing/tracer.go` (未来规划)**:
    *   封装 OpenTelemetry 或其他追踪库的初始化和常用操作，提供简化的接口。
*   **`pkg/healthcheck/checker.go` (如果通用性高)**:
    *   提供通用的健康检查组件，可以聚合多个依赖服务的健康状态。

## `configs/` - 配置文件

*   **`app.yaml`**: 主配置文件，包含所有环境通用的配置或默认配置。
*   **`app.development.yaml`**: 开发环境特定配置 (可覆盖 `app.yaml` 中的同名项)。
*   **`app.production.yaml`**: 生产环境特定配置。
*   **.env (可选, 与app.yaml配合或替代)**: 存放环境变量，尤其适合敏感信息。

## `scripts/` - 脚本文件

*   **`build.sh`**: 构建 Go 应用二进制文件的脚本。
*   **`run.sh`**: (本地) 运行应用的脚本 (可能包含编译步骤)。
*   **`migrate.sh`**: 执行数据库迁移的脚本。
*   **`generate_mocks.sh` (如果使用 mock 工具如 gomock)**: 生成测试 mock 文件的脚本。
*   **`lint.sh`**: 运行静态代码检查的脚本。
*   **`test.sh`**: 运行所有测试的脚本。

---
