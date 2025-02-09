package controller

import (
	"context"

	"misaki/internal/controller/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type controller struct {
	logger      *zap.Logger
	telegramBot *telegram.TelegramBot
}

func NewController(logger *zap.Logger, telegramBot *telegram.TelegramBot) *controller {
	return &controller{
		logger:      logger,
		telegramBot: telegramBot,
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
	if err := c.telegramBot.StartBot(); err != nil {
		return err
	}

	c.telegramBot.RegisterRoutes()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.telegramBot.Bot.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			if update.Message != nil {
				c.telegramBot.Handle(update.Message)
			}
		}
	}()

	return nil
}
