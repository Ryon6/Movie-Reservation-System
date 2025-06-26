// mrs/cmd/server/wire.go
//go:build wireinject
// +build wireinject

package main

import (
	"mrs/internal/di"
	config "mrs/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// 我们将 gin.Engine 定义为 Server，因为它是我们最终要运行的东西
func InitializeServer(input config.ConfigInput) (*gin.Engine, func(), error) {
	// wire.Build 使用我们预先定义好的 FullAppSet
	// 只需要提供最开始的输入参数即可
	wire.Build(
		di.FullAppSet,
	)
	return nil, nil, nil
}
