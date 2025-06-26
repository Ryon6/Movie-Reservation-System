//go:build wireinject
// +build wireinject

package testutils

import (
	"mrs/internal/di"
	"mrs/internal/infrastructure/config"
	"mrs/internal/utils"
	applog "mrs/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type TestServerComponents struct {
	Router *gin.Engine
	DB     *gorm.DB
	RDB    *redis.Client
	Logger applog.Logger
	Hasher utils.PasswordHasher
}

func NewTestServerComponents(router *gin.Engine,
	db *gorm.DB,
	rdb *redis.Client,
	logger applog.Logger,
	hasher utils.PasswordHasher,
) *TestServerComponents {
	return &TestServerComponents{
		Router: router,
		DB:     db,
		RDB:    rdb,
		Logger: logger,
		Hasher: hasher,
	}
}

func InitializeTestServer(input config.ConfigInput) (*TestServerComponents, func(), error) {
	wire.Build(
		di.FullAppSet,

		NewTestServerComponents,
	)
	return nil, nil, nil
}
