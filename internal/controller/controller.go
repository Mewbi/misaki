package controller

import (
	"context"

	"misaki/config"
	"misaki/internal/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type controller struct {
	logger   *zap.Logger
	telegram *config.Telegram
	service  *service.Service
}

func NewController(config *config.Config, logger *zap.Logger, s *service.Service) *controller {
	return &controller{
		logger:   logger,
		telegram: &config.Telegram,
		service:  s,
	}
}

func Start(lc fx.Lifecycle, c *controller) {
	log := c.logger.Sugar()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := c.StartTelegramBot(); err != nil {
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Infow("Shutting down bot")
			return nil
		},
	})
}

func (c *controller) StartTelegramBot() error {
	log := c.logger.Sugar()
	log.Infow("I'm about to start telegram")
	c.service.Something()
	return nil
}
