package main

import (
	"misaki/config"
	"misaki/internal/controller"
	"misaki/internal/controller/telegram"
	"misaki/internal/repository"
	"misaki/internal/service"
	"misaki/logger"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			repository.NewSQLite,
			service.NewService,
			controller.NewController,
			telegram.NewTelegramBot,
		),
		fx.Invoke(controller.Start),
	).Run()
}
