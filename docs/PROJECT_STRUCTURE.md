# 项目模块化设计 (MRS Go 服务)

本文档详细描述了电影预订系统 (MRS) 后端服务的 Go 项目结构和模块化设计。项目遵循分层架构和高内聚低耦合的原则，以便于开发、维护和未来的扩展。

## 顶层目录结构

```
mrs/
├── cmd/                        # 可执行文件入口点
│   ├── migrate/               # 数据库迁移工具
│   └── server/               # 主服务器入口
├── test/                      # 测试目录
│   ├── integration/          # 集成测试
│   ├── e2e/                 # 端到端测试
│   └── performance/         # 性能测试
├── internal/                   # 项目内部私有代码
├── pkg/                        # 可被外部项目导入的公共库代码
├── docs/                       # 项目文档
│   ├── API_SPECIFICATION.md   # API 规范文档
│   ├── DATA_MODEL.md         # 数据模型文档
│   ├── DEVELOPMENT_ROADMAP.md # 开发路线图
│   ├── PROJECT_STRUCTURE.md  # 项目结构文档
│   └── SERVICE_CLASSIFICATION.md # 服务分类文档
├── scripts/                    # 构建、部署、工具类脚本
├── var/                       # 可变文件目录（如日志）
├── go.mod                      # Go 模块定义文件
├── go.sum                      # Go 模块校验和文件
├── Dockerfile                  # 服务构建 Dockerfile
├── docker-compose.yml          # 本地开发环境编排文件
└── README.md                   # 项目主 README
```

## 领域层目录结构

```
internal/
└── domain/
    ├── booking/
    │   ├── booking.go              # Booking 聚合根实体定义
    │   ├── booked_seat.go         # BookedSeat 实体定义
    │   ├── booking_repository.go  # BookingRepository 接口定义
    │   ├── booked_seat_repository.go # BookedSeatRepository 接口定义
    │   └── errors.go              # 领域错误定义
    │
    ├── cinema/
    │   ├── cinema_hall.go         # CinemaHall 聚合根实体定义
    │   ├── seat.go                # Seat 实体定义
    │   ├── cinema_hall_repository.go # CinemaHallRepository 接口定义
    │   ├── seat_repository.go     # SeatRepository 接口定义
    │   ├── seat_cache.go         # 座位缓存接口定义
    │   └── errors.go              # 领域错误定义
    │
    ├── movie/
    │   ├── movie.go               # Movie 聚合根实体定义
    │   ├── genre.go               # Genre 实体定义
    │   ├── movie_repository.go    # MovieRepository 接口定义
    │   ├── genre_repository.go    # GenreRepository 接口定义
    │   ├── movie_cache.go        # 电影缓存接口定义
    │   └── errors.go              # 领域错误定义
    │
    ├── showtime/
    │   ├── showtime.go            # Showtime 聚合根实体定义
    │   ├── showtime_repository.go # ShowtimeRepository 接口定义
    │   ├── showtime_cache.go     # 场次缓存接口定义
    │   └── errors.go              # 领域错误定义
    │
    ├── user/
    │   ├── user.go                # User 聚合根实体定义
    │   ├── role.go                # Role 实体定义
    │   ├── user_repository.go     # UserRepository 接口定义
    │   ├── role_repository.go     # RoleRepository 接口定义
    │   └── errors.go              # 领域错误定义
    │
    └── shared/                    # 跨领域共享的元素
        ├── errors.go              # 通用领域错误定义
        ├── lock/                  # 分布式锁
        │   ├── lock.go           # 锁接口定义
        │   ├── provider.go       # 锁提供者接口
        │   └── errors.go         # 锁相关错误定义
        ├── uow.go                # 工作单元（Unit of Work）接口
        └── vo/                   # 值对象（Value Objects）
            └── identifier.go     # ID类型定义
```

## 应用层目录结构

```
internal/
└── app/
    ├── auth_service.go          # 认证服务
    ├── booking_service.go       # 预订服务
    ├── cinema_service.go        # 影院服务
    ├── movie_service.go         # 电影服务
    ├── report_service.go        # 报告服务
    ├── showtime_service.go      # 场次服务
    └── user_service.go          # 用户服务
```

## API层目录结构

```
internal/
└── api/
    ├── dto/                     # 数据传输对象
    │   ├── request/            # 请求DTO
    │   │   ├── auth_request.go
    │   │   ├── booking_request.go
    │   │   ├── cinema_request.go
    │   │   ├── common.go
    │   │   ├── movie_request.go
    │   │   ├── report_request.go
    │   │   ├── showtime_request.go
    │   │   └── user_request.go
    │   └── response/           # 响应DTO
    │       ├── auth_response.go
    │       ├── booking_response.go
    │       ├── cinema_response.go
    │       ├── common.go
    │       ├── health_response.go
    │       ├── movie_reponse.go
    │       ├── report_response.go
    │       ├── showtime_response.go
    │       └── user_response.go
    ├── handlers/               # HTTP处理器
    │   ├── auth_handler.go
    │   ├── booking_handler.go
    │   ├── cinema_handler.go
    │   ├── common.go
    │   ├── health_handler.go
    │   ├── movie_handler.go
    │   ├── report_handler.go
    │   ├── showtime_handler.go
    │   └── user_handler.go
    ├── middleware/            # HTTP中间件
    │   └── auth_middleware.go
    └── routers/              # 路由配置
        └── router.go
```

## 基础设施层目录结构

```
internal/
└── infrastructure/
    ├── cache/                  # 缓存实现
    │   ├── movie_redis_cache.go
    │   ├── redis_client.go
    │   ├── redis_lock_provider.go
    │   ├── redis_lock.go
    │   ├── seat_redis_cache.go
    │   └── showtime_redis_cache.go
    ├── config/                 # 配置管理
    │   ├── loader.go
    │   └── model.go
    └── persistence/           # 数据持久化
        ├── dbFactory/        # 数据库工厂
        │   └── db_factory.go
        └── mysql/           # MySQL实现
            ├── models/     # GORM模型
            │   ├── booked_seat_gorm.go
            │   ├── booking_gorm.go
            │   ├── cinema_hall_gorm.go
            │   ├── genre_gorm.go
            │   ├── movie_gorm.go
            │   ├── role_gorm.go
            │   ├── seat_gorm.go
            │   ├── showtime_gorm.go
            │   └── user_gorm.go
            └── repository/  # 仓储实现
                ├── booked_seat_repository.go
                ├── booking_repository.go
                ├── cinema_hall_repository.go
                ├── common.go
                ├── db_client.go
                ├── genre_repository.go
                ├── movie_repository.go
                ├── role_repository.go
                ├── seat_repository.go
                ├── showtime_repository.go
                ├── uow.go
                └── user_repository.go
```

## 工具层目录结构

```
internal/
└── utils/                    # 工具函数
    ├── errors.go            # 错误处理工具
    ├── hasher.go           # 密码哈希工具
    └── jwt.go              # JWT工具

pkg/
└── log/                    # 日志工具
    ├── logger.go          # 日志接口
    └── ZapLogger.go       # Zap实现
```

## 测试目录结构

```
test/                           # 测试根目录
├── integration/               # 集成测试
├── e2e/                      # 端到端测试
└── performance/              # 性能测试
```

## 配置文件

项目使用 YAML 格式的配置文件，支持不同环境的配置。配置文件通过环境变量和命令行参数进行覆盖。

## 构建和部署

项目使用 Docker 进行容器化部署，使用 docker-compose 进行本地开发环境的管理。生产环境可以使用 Kubernetes 进行部署。

## 开发工具

- Go 1.22 或更高版本
- Docker 和 docker-compose
- IDE 推荐使用 VSCode 或 Cursor

---
