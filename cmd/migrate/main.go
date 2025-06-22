package main

import (
	"context"
	"fmt"
	"log"
	"mrs/internal/domain/user"
	"mrs/internal/infrastructure/config"
	"mrs/internal/infrastructure/persistence/mysql/models"
	appmysql "mrs/internal/infrastructure/persistence/mysql/repository"
	"mrs/internal/utils"
	applog "mrs/pkg/log"
	"os"

	"github.com/spf13/viper"
)

func initConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig("config", "app.dev", "yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func ensureLogDirectory(logPath string) error {
	if err := os.MkdirAll(logPath, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	return nil
}

func initLogger(cfg *config.Config) applog.Logger {
	logger, err := applog.NewZapLogger(cfg.LogConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	return logger
}

func createInitialRoles(roleRepo user.RoleRepository, logger applog.Logger) error {
	// 创建管理员角色
	adminRole := &user.Role{
		Name:        user.AdminRoleName,
		Description: "系统管理员",
	}
	ctx := context.Background()
	_, err := roleRepo.Create(ctx, adminRole)
	if err != nil {
		return fmt.Errorf("创建管理员角色失败: %w", err)
	}
	logger.Info("成功创建管理员角色")

	// 创建普通用户角色
	userRole := &user.Role{
		Name:        user.UserRoleName,
		Description: "普通用户",
	}

	_, err = roleRepo.Create(ctx, userRole)
	if err != nil {
		return fmt.Errorf("创建普通用户角色失败: %w", err)
	}
	logger.Info("成功创建普通用户角色")

	return nil
}

func createAdminUser(userRepo user.UserRepository, roleRepo user.RoleRepository, hasher utils.PasswordHasher, logger applog.Logger) error {
	// 获取管理员角色
	ctx := context.Background()
	adminRole, err := roleRepo.FindByName(ctx, user.AdminRoleName)
	if err != nil {
		return fmt.Errorf("获取管理员角色失败: %w", err)
	}

	// 创建管理员用户
	adminUser := &user.User{
		Username: "admin",
		Email:    "admin@example.com",
		Role:     adminRole,
	}

	// 设置密码
	if err := adminUser.SetPassword("admin123", hasher); err != nil {
		return fmt.Errorf("设置管理员密码失败: %w", err)
	}

	if err := userRepo.Create(ctx, adminUser); err != nil {
		return fmt.Errorf("创建管理员用户失败: %w", err)
	}
	logger.Info("成功创建管理员用户")

	return nil
}

func main() {
	// 初始化配置
	cfg, err := initConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	viper.Set("config", cfg)

	// 确保日志目录存在
	if err := ensureLogDirectory("./var/log"); err != nil {
		log.Fatalf("Failed to ensure log directory: %v", err)
	}

	// 初始化日志
	logger := initLogger(cfg)
	defer logger.Sync() // 确保所有日志都已刷新到磁盘

	logger.Info("开始数据库迁移程序")

	// 初始化数据库连接
	dbFactory := appmysql.NewMysqlDBFactory(logger)
	db, err := dbFactory.CreateDBConnection(cfg.DatabaseConfig)
	if err != nil {
		logger.Fatal("数据库连接失败", applog.Error(err))
	}

	logger.Info("数据库连接成功，开始迁移模型")

	// 自动迁移所有模型
	err = db.AutoMigrate(
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
		logger.Fatal("模型迁移失败", applog.Error(err))
	}

	logger.Info("数据库迁移完成，开始创建初始角色和管理员用户")

	// 创建密码哈希工具
	hasher := utils.NewBcryptHasher(cfg.AuthConfig.HasherCost)

	// 创建仓储实例
	roleRepo := appmysql.NewGormRoleRepository(db, logger)
	userRepo := appmysql.NewGormUserRepository(db, logger)

	// 创建初始角色
	if err := createInitialRoles(roleRepo, logger); err != nil {
		logger.Fatal("创建初始角色失败", applog.Error(err))
	}

	// 创建管理员用户
	if err := createAdminUser(userRepo, roleRepo, hasher, logger); err != nil {
		logger.Fatal("创建管理员用户失败", applog.Error(err))
	}

	logger.Info("初始化完成")
}
