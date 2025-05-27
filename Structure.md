### 项目模块化设计 (Go Package 结构)

在 Go 中，模块化主要通过包 (package) 来实现。一个典型的 Go Web 服务项目结构可以如下组织，我们将 MRS 的功能映射到这个结构中：

```
mrs/
├── cmd/                    // 可执行文件入口
│   └── server/
│       └── main.go         // 应用启动、依赖注入、HTTP服务器启动
├── internal/               // 项目内部代码，不希望被其他项目导入
│   ├── api/                // HTTP API层 (也常被称为 handler, transport, controller)
│   │   ├── handlers/       // 具体的HTTP请求处理器
│   │   │   ├── auth_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── movie_handler.go
│   │   │   ├── showtime_handler.go
│   │   │   ├── booking_handler.go
│   │   │   └── admin_handler.go // (或将admin特定路由分散到对应模块)
│   │   ├── middleware/     // HTTP中间件 (如认证、日志、CORS)
│   │   │   ├── auth_middleware.go
│   │   │   └── logging_middleware.go
│   │   └── router.go       // API路由定义 (例如使用Gin的router)
│   │
│   ├── auth/               // 认证与授权服务
│   │   ├── jwt_service.go
│   │   └── password_service.go
│   │
│   ├── booking/            // 预订模块 (核心业务逻辑)
│   │   ├── service.go      // 预订相关的业务逻辑
│   │   └── repository.go   // 预订相关的数据持久化操作 (DB & Cache)
│   │
│   ├── cache/              // 缓存层抽象与实现
│   │   ├── redis_client.go // Redis客户端初始化和通用操作
│   │   ├── movie_cache.go  // 电影相关的缓存逻辑
│   │   └── showtime_cache.go // 场次座位相关的缓存逻辑
│   │
│   ├── config/             // 配置加载与管理
│   │   └── config.go       // (例如使用Viper)
│   │
│   ├── database/           // 数据库连接与迁移
│   │   ├── mysql_client.go // MySQL客户端初始化 (GORM)
│   │   └── migrations/     // 数据库迁移脚本 (如果使用迁移工具)
│   │
│   ├── domain/             // 核心领域模型/实体 (你的数据模型定义)
│   │   ├── user.go
│   │   ├── movie.go
│   │   ├── showtime.go
│   │   ├── booking.go
│   │   ├── seat.go
│   │   └── common.go       // (例如分页请求/响应结构)
│   │
│   ├── movie/              // 电影与场次管理模块
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── report/             // 报告模块
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── user/               // 用户管理模块
│   │   ├── service.go
│   │   └── repository.go
│   │
│   └── utils/              // 通用工具函数
│       ├── error_handler.go // 统一错误处理
│       ├── validators.go    // 请求数据校验
│       └── pagination.go    // 分页逻辑辅助
│
├── pkg/                    // (可选) 可以被外部项目导入的库代码
│   └── apierrors/          // 自定义API错误类型，如果需要跨项目共享
│
├── go.mod                  // Go模块文件
├── go.sum
├── Dockerfile              // Docker构建文件
├── docker-compose.yml      // (可选) 用于本地开发环境编排
└── README.md
```

**模块说明:**

*   **`cmd/server`**: 应用主入口，负责初始化所有依赖（配置、数据库、缓存、服务等），然后启动 HTTP 服务器。
*   **`internal/api`**: 处理所有 HTTP 相关逻辑。
    *   `handlers`: 接收 HTTP 请求，调用相应的 `service` 处理业务，然后返回 HTTP 响应。它们不应包含复杂的业务逻辑。
    *   `middleware`: 处理跨多个请求的通用逻辑，如身份验证、日志记录。
    *   `router`: 定义 URL 路径与 `handler` 的映射关系。
*   **`internal/auth`**: 专门处理用户认证（如 JWT 生成和校验）和密码安全（哈希、比较）。
*   **`internal/<domain_module>` (e.g., `booking`, `movie`, `user`, `report`)**: 按业务领域划分的核心模块。
    *   `service.go`: 包含该领域的核心业务逻辑。它会调用 `repository` 进行数据操作，可能也会调用其他 `service`。这是业务规则的所在地。
    *   `repository.go`: 数据访问层 (DAL)，负责与数据库和缓存的交互。它封装了数据读写的具体实现，`service` 层不直接关心数据是如何存储和检索的。
*   **`internal/cache`**: 缓存的具体实现和管理。可以提供通用的 Redis 客户端，也可以针对不同数据提供特定的缓存操作封装。
*   **`internal/config`**: 加载和管理应用的配置（如数据库连接字符串、JWT 密钥、服务器端口等）。
*   **`internal/database`**: 数据库连接的初始化和管理，可能还包括数据库迁移的逻辑。
*   **`internal/domain` (或 `internal/models`)**: 定义核心数据结构（对应你的数据模型中的表）。这些结构会在 `repository`、`service` 和 `api` 层之间传递。
*   **`internal/utils`**: 放置一些通用的辅助函数，如自定义错误处理、数据校验、分页逻辑等。
*   **`pkg/`**: 如果你有一些代码片段或库希望被其他项目复用，可以放在这里。对于大多数单体应用，`internal/` 已经足够。