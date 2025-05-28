### 项目开发路线 (迭代式开发)

我们将采用迭代的方式进行开发，每个阶段都有明确的目标和产出。

**阶段 0: 项目初始化与基础架构搭建 (准备工作)**

1.  **目标**: 建立项目骨架，配置好开发环境。
2.  **任务**:
    *   创建 Git 仓库。
    *   初始化 Go Modules (`go mod init mrs`)。
    *   搭建基础目录结构 (如上所述)。
    *   引入 Gin Web 框架。
    *   引入配置管理库 (如 Viper) 并加载基础配置 (服务端口)。
    *   引入日志库 (如 Zap 或 Logrus) 并配置基础日志。
    *   设置 MySQL 数据库连接 (使用 GORM 或 `database/sql`)。
    *   设置 Redis 连接。
    *   编写一个简单的健康检查 API (`/health`)，确认服务能跑起来。
    *   编写 Dockerfile 和 docker-compose.yml (用于MySQL, Redis, 和应用本身)。
3.  **产出**: 一个可以运行的、包含基础设置的 Go 服务。

**阶段 1: 用户认证与授权模块 (核心安全)**

1.  **目标**: 实现用户注册、登录、角色管理和基于角色的访问控制。
2.  **任务**:
    *   在 `internal/domain` 中定义 `User` 和 `Role` 结构体。
    *   实现 `internal/user/repository`：用户数据的 CRUD。
    *   实现 `internal/auth/password_service`：密码加密与比对。
    *   实现 `internal/auth/jwt_service`：JWT 生成与校验。
    *   实现 `internal/user/service`：用户注册、登录逻辑。
    *   实现 `internal/api/handlers/auth_handler.go`：处理 `/register`, `/login` 路由。
    *   实现 `internal/api/middleware/auth_middleware.go`：JWT 认证中间件，保护需要登录的路由，并将用户信息存入 Context。
    *   数据库 seeding：创建初始管理员账户。
    *   （可选）管理员提升用户角色的 API。
3.  **产出**: 用户可以注册、登录，受保护的 API 得到保护。

**阶段 2: 电影与放映时间管理模块 (管理员功能)**

1.  **目标**: 管理员可以管理电影、类型和放映时间。引入基础缓存。
2.  **任务**:
    *   在 `internal/domain` 定义 `Movie`, `Genre`, `MovieGenre`, `CinemaHall`, `Showtime` 结构体。
    *   实现 `internal/movie/repository`：电影、类型、影厅、放映时间的 CRUD。
    *   实现 `internal/movie/service`：相应的业务逻辑。
    *   实现 `internal/cache/movie_cache.go` 和 `showtime_cache.go`：对电影列表、电影详情、场次列表等常用查询结果进行缓存（读操作）。当数据发生写操作（增删改）时，注意缓存的更新或失效策略。
    *   实现 `internal/api/handlers/movie_handler.go` 和 `showtime_handler.go`：供管理员使用的相关 API，并使用认证中间件和角色检查确保只有管理员能访问。
3.  **产出**: 管理员可以管理电影和场次信息，常用查询通过缓存加速。

**阶段 3: 用户预订流程模块 (核心用户功能与高并发应对)**

1.  **目标**: 用户可以浏览电影、查询场次、查看座位图、预订座位、查看和取消自己的预订。重点关注高并发下的数据一致性。
2.  **任务**:
    *   在 `internal/domain` 定义 `Seat`, `Booking`, `BookedSeat` 结构体。
    *   实现 `internal/movie/service` 和 `repository` 的用户侧查询功能：
        *   获取电影列表（带分页、筛选，利用缓存）。
        *   获取特定日期的电影及放映时间（利用缓存）。
        *   获取特定场次的座位图及可用状态（**关键点：缓存与实时性平衡**）。
    *   实现 `internal/booking/repository`：订单和已预订座位的 CRUD。
    *   实现 `internal/booking/service`：
        *   座位预订核心逻辑：**处理并发控制** (例如使用 Redis 分布式锁 `SETNX` 锁定座位或场次，或数据库行级锁/乐观锁)，防止超额预订。
        *   创建订单。
        *   查看用户预订。
        *   取消预订（校验逻辑）。
    *   `internal/cache/showtime_cache.go` 或新的 `seat_cache.go`：缓存座位状态，并设计好座位状态变更时的缓存更新逻辑。
    *   实现 `internal/api/handlers/movie_handler.go` (用户侧查询) 和 `booking_handler.go` (用户预订相关)。
3.  **产出**: 用户可以完成整个预订流程，系统在高并发预订场景下能保证数据正确性。

**阶段 4: 管理员报告功能**

1.  **目标**: 管理员可以查看系统报告。
2.  **任务**:
    *   实现 `internal/report/repository`：从数据库聚合数据生成报告（所有预订、上座率、收入）。
    *   实现 `internal/report/service`：报告生成的业务逻辑。
    *   实现 `internal/api/handlers/admin_handler.go` 或 `report_handler.go`：提供报告 API。
3.  **产出**: 管理员可以获取运营数据报告。

**阶段 5: 测试、优化与完善**

1.  **目标**: 提升系统质量、性能和可维护性。
2.  **任务**:
    *   **单元测试**: 为 `service` 层和 `repository` 层（如果逻辑复杂）编写单元测试。
    *   **集成测试**: 测试 API 端点，确保模块间协作正确。
    *   **性能测试 (可选但推荐)**: 针对高并发场景（如座位预订）进行压力测试，找出瓶颈并优化。
    *   代码审查和重构。
    *   完善日志记录和错误处理。
    *   API 文档 (例如使用 Swagger/OpenAPI)。
    *   安全性加固 (例如输入校验、防止 SQL 注入、XSS 等，虽然 GORM 和 Gin 会处理一些)。
3.  **产出**: 一个经过测试、相对稳定和高效的系统。

**阶段 6: 部署 (占位)**

1.  **目标**: 将应用部署到生产环境。
2.  **任务**: 准备生产环境配置，CI/CD 流水线等。

### 开发过程中的注意事项：

*   **迭代和反馈**: 每个阶段完成后，都可以进行测试和回顾，根据反馈调整后续计划。
*   **接口先行**: 可以先定义好各模块间的接口 (Go interface)，这有助于解耦和并行开发（即使是你一个人，这也是好习惯）。例如，`service` 依赖 `repository` 的接口，而不是具体实现。
*   **错误处理**: Go 的错误处理模式 (`if err != nil`) 要贯彻始终。设计统一的错误返回格式给前端。
*   **日志**: 记录关键操作和错误，便于调试和监控。
*   **依赖注入**: 在 `cmd/server/main.go` 中初始化依赖，并将它们注入到需要的地方（例如，将 `db` 实例和 `redis` 客户端注入到 `repository`，将 `repository` 注入到 `service`，将 `service` 注入到 `handler`）。