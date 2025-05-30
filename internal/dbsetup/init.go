// internal/dbsetup/init.go (或者在 main.go 中)
package dbsetup

import (
	"errors"
	"fmt"
	"mrs/internal/domain/role" // 你的 Role 实体定义
	"mrs/internal/domain/user" // 你的 User 实体定义
	"mrs/internal/infrastructure/config"

	// 你的 PasswordHasher 工具
	applog "mrs/pkg/log" // 你的日志包

	"golang.org/x/crypto/bcrypt" // 可以直接使用 bcrypt 进行初始密码哈希
	"gorm.io/gorm"
)

// InitializeDatabase 执行数据库迁移和初始数据植入。
func InitializeDatabase(db *gorm.DB, admCfg config.AdminConfig, logger applog.Logger) error {
	logger.Info("Starting database initialization...")

	// 1. 自动迁移表结构
	// AutoMigrate 会创建表、缺失的列和索引，并且不会删除现有的列或数据。
	// 确保你的 User 和 Role 结构体定义中的 GORM 标签是正确的 (例如 `gorm:"foreignKey:RoleID"`)
	// 以便正确创建外键。
	logger.Info("Running auto-migration for Role and User tables...")
	err := db.AutoMigrate(&role.Role{})
	if err != nil {
		logger.Error("Failed to auto-migrate role tables", applog.Error(err))
		return fmt.Errorf("failed to auto-migrate role tables: %w", err)
	}
	err = db.AutoMigrate(&user.User{})
	if err != nil {
		logger.Error("Failed to auto-migrate user tables", applog.Error(err))
		return fmt.Errorf("failed to auto-migrate user tables: %w", err)
	}
	logger.Info("Auto-migration completed successfully.")

	// 2. 植入角色数据

	adminRole, err := seedRole(db, logger, role.AdminRoleName, "Administrator with full access")
	if err != nil {
		return fmt.Errorf("failed to seed admin role: %w", err)
	}

	_, err = seedRole(db, logger, role.UserRoleName, "Standard user with basic access")
	if err != nil {
		return fmt.Errorf("failed to seed user role: %w", err)
	}

	// 3. 植入管理员账户

	err = seedAdminUser(db, logger, admCfg.Username, admCfg.Email, admCfg.Password, adminRole)
	if err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	logger.Info("Database initialization completed successfully.")
	return nil
}

// seedRole 植入一个角色，如果它尚不存在。
func seedRole(db *gorm.DB, logger applog.Logger, name, description string) (*role.Role, error) {
	log := logger.With(applog.String("role_name", name))
	var existingRole role.Role

	// 检查角色是否已存在
	// GORM 的 FirstOrCreate 也可以用于此目的:
	// result := db.FirstOrCreate(&existingRole, role.Role{Name: name, Description: description})
	// if result.Error != nil { ... }
	// if result.RowsAffected > 0 { log.Info("Role created.") } else { log.Info("Role already exists.")}

	err := db.Where("name = ?", name).First(&existingRole).Error
	if err == nil {
		// 角色已存在
		log.Info("Role already exists, skipping creation.")
		return &existingRole, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 查询时发生其他错误
		log.Error("Failed to check if role exists", applog.Error(err))
		return nil, err
	}

	// 角色不存在，创建它
	newRole := role.Role{
		Name:        name, // [2]
		Description: description,
	}
	if err := db.Create(&newRole).Error; err != nil {
		log.Error("Failed to create role", applog.Error(err))
		return nil, err
	}
	log.Info("Role created successfully.", applog.Uint("role_id", newRole.ID))
	return &newRole, nil
}

// seedAdminUser 植入管理员账户，如果它尚不存在。
func seedAdminUser(db *gorm.DB, logger applog.Logger, username, email, password string, adminRole *role.Role) error {
	log := logger.With(applog.String("admin_username", username), applog.String("admin_email", email))
	var existingUser user.User

	// 检查管理员用户是否已存在（例如通过用户名或邮箱）
	err := db.Where("username = ? OR email = ?", username, email).First(&existingUser).Error
	if err == nil {
		log.Info("Admin user already exists, skipping creation.")
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("Failed to check if admin user exists", applog.Error(err))
		return err
	}

	// 哈希密码
	// 在实际应用中，应该使用之前创建的 utils.PasswordHasher 服务
	// hasher := utils.NewBcryptHasher(bcrypt.DefaultCost)
	// hashedPassword, err := hasher.Hash(password)
	// 为了简化这里的示例，我们直接使用 bcrypt，但在真实项目中应保持一致性。
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to hash admin password", applog.Error(err))
		return fmt.Errorf("failed to hash admin password: %w", err)
	}
	hashedPassword := string(hashedPasswordBytes)

	adminUser := user.User{ // [1]
		Username:     username,       // [1]
		Email:        email,          // [1]
		PasswordHash: hashedPassword, // [1]
		RoleID:       adminRole.ID,   // [1]
		Role:         *adminRole,     // [1]
		// IsActive:     true, // 如果你的 User 实体有这个字段
	}

	if err := db.Create(&adminUser).Error; err != nil {
		log.Error("Failed to create admin user", applog.Error(err))
		return err
	}

	log.Info("Admin user created successfully.", applog.Uint("user_id", adminUser.ID))
	return nil
}
