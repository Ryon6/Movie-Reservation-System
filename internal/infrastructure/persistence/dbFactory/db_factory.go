package dbFactory

import (
	"mrs/internal/infrastructure/config"

	"gorm.io/gorm"
)

// 工厂模式
// DBConnectionFactory 是一个接口，用于创建数据库连接。
type DBConnectionFactory interface {
	CreateDBConnection(dbConfig config.DatabaseConfig) (*gorm.DB, error)
}
