//go:build wireinject
// +build wireinject

package main

import (
	"mrs/internal/di"
	"mrs/internal/infrastructure/config"
	applog "mrs/pkg/log"

	"github.com/google/wire"
	"gorm.io/gorm"
)

type SeedComponents struct {
	DB     *gorm.DB
	Logger applog.Logger
}

func NewSeedComponents(db *gorm.DB, logger applog.Logger) *SeedComponents {
	return &SeedComponents{
		DB:     db,
		Logger: logger,
	}
}

func InitializeSeed(input config.ConfigInput) (*SeedComponents, func(), error) {
	wire.Build(
		di.ConfigSet,
		di.LoggerSet,
		di.DatabaseSet,

		NewSeedComponents,
	)
	return nil, nil, nil
}
