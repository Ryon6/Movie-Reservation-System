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

	"gorm.io/gorm"
)

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

func dropExistingTables(db *gorm.DB, logger applog.Logger) error {
	// 定义需要删除的表名
	tables := []interface{}{
		&models.BookedSeatGorm{},
		&models.BookingGorm{},
		&models.ShowtimeGorm{},
		&models.SeatGorm{},
		&models.CinemaHallGorm{},
		&models.MovieGorm{},
		&models.GenreGorm{},
		&models.UserGorm{},
		&models.RoleGorm{},
	}

	logger.Info("开始删除已存在的表")

	// 删除表（注意顺序：先删除有外键依赖的表）
	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("删除表失败: %w", err)
		}
	}

	logger.Info("已成功删除所有已存在的表")
	return nil
}

func main() {
	// 确保日志目录存在
	if err := os.MkdirAll("./var/log", 0755); err != nil {
		log.Fatalf("Failed to ensure log directory: %v", err)
	}

	components, cleanup, err := InitializeMigrate(config.ConfigInput{
		Path: "config",
		Name: "app.dev",
		Type: "yaml",
	})
	if err != nil {
		log.Fatalf("Failed to initialize migrate: %v", err)
	}
	defer cleanup()

	logger := components.Logger
	db := components.DB
	hasher := components.Hasher

	logger.Info("开始数据库迁移程序")

	// 先删除已存在的表
	if err := dropExistingTables(db, logger); err != nil {
		logger.Fatal("删除已存在表失败", applog.Error(err))
	}

	logger.Info("开始迁移模型")

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
