//go:build wireinject
// +build wireinject

package main

import (
	"mrs/internal/di"
	"mrs/internal/infrastructure/config"
	"mrs/internal/utils"
	applog "mrs/pkg/log"

	"github.com/google/wire"
	"gorm.io/gorm"
)

type MigrateComponents struct {
	DB     *gorm.DB
	Logger applog.Logger
	Hasher utils.PasswordHasher
}

func NewMigrateComponents(db *gorm.DB, logger applog.Logger, hasher utils.PasswordHasher) *MigrateComponents {
	return &MigrateComponents{
		DB:     db,
		Logger: logger,
		Hasher: hasher,
	}
}

func InitializeMigrate(input config.ConfigInput) (*MigrateComponents, func(), error) {
	wire.Build(
		di.ConfigSet,
		di.LoggerSet,
		di.DatabaseSet,
		di.UtilsSet,

		NewMigrateComponents,
	)
	return nil, nil, nil
}
