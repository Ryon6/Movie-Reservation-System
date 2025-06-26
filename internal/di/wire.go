package di

import (
	"mrs/internal/api/handlers"
	"mrs/internal/api/middleware"
	"mrs/internal/api/routers"
	"mrs/internal/app"
	"mrs/internal/infrastructure/cache"
	"mrs/internal/infrastructure/config"
	"mrs/internal/infrastructure/persistence/mysql/repository"
	"mrs/internal/utils"
	applog "mrs/pkg/log"

	"github.com/google/wire"
	"go.uber.org/zap"
)

// ConfigSet 提供了配置加载
var ConfigSet = wire.NewSet(
	config.LoadConfig,
	wire.FieldsOf(new(*config.Config), "DatabaseConfig", "RedisConfig", "LogConfig", "AuthConfig", "JWTConfig", "ServerConfig"),
)

// LoggerSet 提供了日志组件
var LoggerSet = wire.NewSet(
	applog.NewZapLogger,
	wire.Value([]zap.Option{}), // 提供一个空的 zap.Option 切片作为默认值
)

// UtilsSet 提供了工具组件
var UtilsSet = wire.NewSet(
	utils.NewBcryptHasher,
	utils.NewJWTManagerImpl,
)

// DatabaseSet 提供了数据库连接和工作单元 (UoW)
var DatabaseSet = wire.NewSet(
	repository.CreateDBConnection,
	repository.NewGormUnitOfWork,
)

// RedisSet 提供了Redis客户端
var RedisSet = wire.NewSet(
	cache.NewRedisClient,
	cache.NewRedisLockProvider,
)

// RepositorySet 提供了仓库组件
var RepositorySet = wire.NewSet(
	repository.NewGormUserRepository,
	repository.NewGormRoleRepository,
	repository.NewGormMovieRepository,
	repository.NewGormGenreRepository,
	repository.NewGormCinemaHallRepository,
	repository.NewGormShowtimeRepository,
	repository.NewGormBookingRepository,
	repository.NewGormBookedSeatRepository,
)

// CacheSet 提供了缓存组件
var CacheSet = wire.NewSet(
	cache.NewRedisMovieCache,
	cache.NewRedisShowtimeCache,
	cache.NewRedisSeatCache,
)

// ServiceSet 提供了服务组件
var ServiceSet = wire.NewSet(
	app.NewAuthService,
	app.NewUserService,
	app.NewMovieService,
	app.NewCinemaService,
	app.NewShowtimeService,
	app.NewBookingService,
	app.NewReportService,
)

// HandlerSet 提供了处理器组件
var HandlerSet = wire.NewSet(
	handlers.NewHealthHandler,
	handlers.NewAuthHandler,
	handlers.NewUserHandler,
	handlers.NewMovieHandler,
	handlers.NewCinemaHandler,
	handlers.NewShowtimeHandler,
	handlers.NewBookingHandler,
	handlers.NewReportHandler,
)

// MiddlewareSet 提供了中间件组件
var MiddlewareSet = wire.NewSet(
	middleware.AdminMiddleware,
	middleware.AuthMiddleware,
	middleware.LoggerMiddleware,
)

// RouterSet 提供了路由组件
var RouterSet = wire.NewSet(
	routers.SetupRouter,
)

// FullAppSet 是一个方便的集合，包含了启动一个完整 Web App 所需的所有组件
var FullAppSet = wire.NewSet(
	ConfigSet,
	LoggerSet,
	DatabaseSet,
	RedisSet,
	UtilsSet,
	RepositorySet,
	CacheSet,
	ServiceSet,
	HandlerSet,
	MiddlewareSet,
	RouterSet,
)
