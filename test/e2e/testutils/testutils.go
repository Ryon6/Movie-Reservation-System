package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/api/handlers"
	"mrs/internal/api/middleware"
	"mrs/internal/api/routers"
	"mrs/internal/app"
	"mrs/internal/domain/user"
	"mrs/internal/infrastructure/cache"
	"mrs/internal/infrastructure/config"
	"mrs/internal/infrastructure/persistence/mysql/models"
	appmysql "mrs/internal/infrastructure/persistence/mysql/repository"
	"mrs/internal/utils"
	applog "mrs/pkg/log"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestServer 封装测试服务器和相关工具
type TestServer struct {
	Router     *gin.Engine
	Server     *httptest.Server
	DB         *gorm.DB
	RDB        *redis.Client
	Logger     applog.Logger
	AdminToken string
	UserToken  string
}

// NewTestServer 创建并初始化一个完整的测试服务器
func NewTestServer(t *testing.T) *TestServer {
	gin.SetMode(gin.TestMode)

	// 初始化配置
	cfg, err := config.LoadConfig("../fixtures/config", "app.e2e", "yaml")
	assert.NoError(t, err)

	// 初始化日志
	logger, err := applog.NewZapLogger(cfg.LogConfig)
	assert.NoError(t, err)

	// 初始化数据库
	dbFactory := appmysql.NewMysqlDBFactory(logger)
	db, err := dbFactory.CreateDBConnection(cfg.DatabaseConfig)
	assert.NoError(t, err)

	// 清理并迁移数据库
	CleanAndMigrateDB(t, db, cfg.AdminConfig, logger)

	// 初始化 Redis
	rdbClient, err := cache.NewRedisClient(cfg.RedisConfig, logger)
	assert.NoError(t, err)
	rdb := rdbClient.(*redis.Client)
	CleanRedis(t, rdb)

	// 实用工具
	hasher := utils.NewBcryptHasher(cfg.AuthConfig.HasherCost)
	jwtManager, err := utils.NewJWTManagerImpl(
		cfg.JWTConfig.SecretKey,
		cfg.JWTConfig.Issuer,
		int64(cfg.JWTConfig.AccessTokenDuration.Hours()),
	)
	assert.NoError(t, err)

	// 基础设施层
	userRepo := appmysql.NewGormUserRepository(db, logger)
	roleRepo := appmysql.NewGormRoleRepository(db, logger)
	movieRepo := appmysql.NewGormMovieRepository(db, logger)
	genreRepo := appmysql.NewGormGenreRepository(db, logger)
	cinemaRepo := appmysql.NewGormCinemaHallRepository(db, logger)
	seatRepo := appmysql.NewGormSeatRepository(db, logger)
	showtimeRepo := appmysql.NewGormShowtimeRepository(db, logger)
	bookingRepo := appmysql.NewGormBookingRepository(db, logger)
	movieCache := cache.NewRedisMovieCache(rdb, logger, time.Second*5)
	showtimeCache := cache.NewRedisShowtimeCache(rdb, logger, time.Second*5)
	seatCache := cache.NewRedisSeatCache(rdb, logger, time.Second*5)
	lockProvider := cache.NewRedisLockProvider(rdb, logger)
	uow := appmysql.NewGormUnitOfWork(db, logger)

	// 应用层
	userService := app.NewUserService(cfg.AuthConfig.DefaultRoleName, uow, userRepo, roleRepo, hasher, logger)
	authService := app.NewAuthService(uow, userRepo, hasher, jwtManager, logger)
	movieService := app.NewMovieService(uow, movieRepo, genreRepo, movieCache, logger)
	cinemaService := app.NewCinemaService(uow, cinemaRepo, seatRepo, logger)
	showtimeService := app.NewShowtimeService(uow, showtimeRepo, seatRepo, bookingRepo, showtimeCache, seatCache, lockProvider, logger)
	bookingService := app.NewBookingService(uow, bookingRepo, showtimeRepo, seatCache, showtimeCache, showtimeService, lockProvider, logger)
	reportService := app.NewReportService(logger, bookingRepo)

	// 接口层
	healthHandler := handlers.NewHealthHandler(db, rdb, logger)
	authHandler := handlers.NewAuthHandler(authService, logger)
	userHandler := handlers.NewUserHandler(userService, logger)
	movieHandler := handlers.NewMovieHandler(movieService, logger)
	cinemaHandler := handlers.NewCinemaHandler(cinemaService, logger)
	showtimeHandler := handlers.NewShowtimeHandler(showtimeService, logger)
	bookingHandler := handlers.NewBookingHandler(bookingService, logger)
	reportHandler := handlers.NewReportHandler(reportService, logger)

	// 设置路由
	router := routers.SetupRouter(healthHandler,
		authHandler,
		userHandler,
		movieHandler,
		cinemaHandler,
		showtimeHandler,
		bookingHandler,
		reportHandler,
		middleware.AuthMiddleware(jwtManager, logger),
		middleware.AdminMiddleware(jwtManager, logger), // 修复 AdminMiddleware 参数
	)

	server := httptest.NewServer(router)

	ts := &TestServer{
		Router: router,
		Server: server,
		DB:     db,
		RDB:    rdb,
		Logger: logger,
	}

	// 为测试播种基础用户数据
	err = SeedUsers(db, hasher, cfg.AdminConfig.Username, cfg.AdminConfig.Password, cfg.AdminConfig.Email)
	assert.NoError(t, err)

	return ts
}

// Close 关闭测试服务器和连接
func (ts *TestServer) Close() {
	ts.Server.Close()
	sqlDB, _ := ts.DB.DB()
	sqlDB.Close()
	ts.RDB.Close()
}

// DoRequest 发送HTTP请求并返回响应
func (ts *TestServer) DoRequest(t *testing.T, method, path string, body interface{}, token string) (*http.Response, []byte) {
	var reqBody io.Reader
	targetURL := ts.Server.URL + path

	if body != nil {
		if method == http.MethodGet {
			// 对于GET请求，将body转换为查询参数
			query, err := StructToQuery(body)
			assert.NoError(t, err)
			if query != "" {
				targetURL += "?" + query
			}
		} else {
			// 对于其他方法，使用JSON body
			jsonBody, err := json.Marshal(body)
			assert.NoError(t, err)
			reqBody = bytes.NewBuffer(jsonBody)
		}
	}

	req, err := http.NewRequest(method, targetURL, reqBody)
	assert.NoError(t, err)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil && method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}

	ts.Logger.Debug("req", applog.Any("req", req))

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	return resp, respBody
}

// Login 执行登录操作并返回token
func (ts *TestServer) Login(t *testing.T, username, password string) string {
	loginReq := request.LoginRequest{
		Username: username,
		Password: password,
	}

	resp, body := ts.DoRequest(t, http.MethodPost, "/api/v1/auth/login", loginReq, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed for user %s: status code %d, body: %s", username, resp.StatusCode, string(body))
	}

	var loginResp response.LoginResponse
	err := json.Unmarshal(body, &loginResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResp.Token)

	return loginResp.Token
}

// AssertResponseCode 断言响应状态码
func AssertResponseCode(t *testing.T, expected, actual int, body []byte) {
	if expected != actual {
		t.Errorf("Expected status code %d but got %d. Response body: %s", expected, actual, string(body))
	}
}

// ParseResponse 解析响应体到指定结构
func ParseResponse(t *testing.T, body []byte, v interface{}) {
	err := json.Unmarshal(body, v)
	assert.NoError(t, err, "Failed to parse response body: %s", string(body))
}

// CleanAndMigrateDB 清理并迁移数据库
func CleanAndMigrateDB(t *testing.T, db *gorm.DB, adminCfg config.AdminConfig, logger applog.Logger) {
	// 获取所有表名
	var tableNames []string
	err := db.Raw("SHOW TABLES").Scan(&tableNames).Error
	assert.NoError(t, err)

	// 关闭外键检查
	err = db.Exec("SET FOREIGN_KEY_CHECKS = 0;").Error
	assert.NoError(t, err)

	// 删除所有表
	for _, tableName := range tableNames {
		err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", tableName)).Error
		assert.NoError(t, err)
	}

	// 开启外键检查
	err = db.Exec("SET FOREIGN_KEY_CHECKS = 1;").Error
	assert.NoError(t, err)

	// 重新进行数据库迁移
	applyMigrations(db, logger)
	assert.NoError(t, err)
}

// applyMigrations 在此文件中实现数据库迁移
func applyMigrations(db *gorm.DB, logger applog.Logger) {
	logger.Info("Starting database migration")
	err := db.AutoMigrate(
		&models.UserGorm{},
		&models.RoleGorm{},
		&models.MovieGorm{},
		&models.GenreGorm{},
		&models.CinemaHallGorm{},
		&models.SeatGorm{},
		&models.ShowtimeGorm{},
		&models.BookingGorm{},
		&models.BookedSeatGorm{},
	)
	if err != nil {
		logger.Fatal("Database migration failed", applog.Error(err))
	}
	logger.Info("Database migration completed successfully")
}

// CleanRedis 清空Redis数据库
func CleanRedis(t *testing.T, rdb *redis.Client) {
	err := rdb.FlushDB(context.Background()).Err()
	assert.NoError(t, err)
}

// SeedUsers 在数据库中植入基础用户数据
func SeedUsers(db *gorm.DB, hasher utils.PasswordHasher, adminUser, adminPass, adminEmail string) error {
	// 创建 Admin 角色
	adminRole := models.RoleGorm{Name: user.AdminRoleName, Description: "Administrator"}
	if err := db.FirstOrCreate(&adminRole, models.RoleGorm{Name: user.AdminRoleName}).Error; err != nil {
		return err
	}

	// 创建 User 角色
	userRole := models.RoleGorm{Name: user.UserRoleName, Description: "Regular User"}
	if err := db.FirstOrCreate(&userRole, models.RoleGorm{Name: user.UserRoleName}).Error; err != nil {
		return err
	}

	// 创建 Admin 用户
	hashedAdminPassword, err := hasher.Hash(adminPass)
	if err != nil {
		return err
	}
	admin := models.UserGorm{
		Username:     adminUser,
		PasswordHash: string(hashedAdminPassword),
		Email:        adminEmail,
		RoleID:       adminRole.ID,
	}
	if err := db.Create(&admin).Error; err != nil {
		if !errors.Is(err, user.ErrUserAlreadyExists) {
			return err
		}
	}

	// 创建一个普通用户用于测试
	hashedUserPassword, err := hasher.Hash("user123")
	if err != nil {
		return err
	}
	testUser := models.UserGorm{
		Username:     "user",
		PasswordHash: string(hashedUserPassword),
		Email:        "user@example.com",
		RoleID:       userRole.ID,
	}
	if err := db.Create(&testUser).Error; err != nil {
		if !errors.Is(err, user.ErrUserAlreadyExists) {
			return err
		}
	}

	return nil
}

// SortSeats 对座位进行排序以保证断言的稳定性
func SortSeats(seats []*response.SeatResponse) {
	sort.Slice(seats, func(i, j int) bool {
		if seats[i].RowIdentifier != seats[j].RowIdentifier {
			return seats[i].RowIdentifier < seats[j].RowIdentifier
		}
		return seats[i].SeatNumber < seats[j].SeatNumber
	})
}
