package controller

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type controller struct {
	logger      *zap.Logger
	telegramBot *telegramBot
}

func NewController(logger *zap.Logger, telegramBot *telegramBot) *controller {
	return &controller{
		logger:      logger,
		telegramBot: telegramBot,
	}
}

// TODO: Create user: only yourself can create your user with a single command or button
// - Create a payment CRUD

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
	bot, err := tgbotapi.NewBotAPI(c.telegramBot.config.Token)
	if err != nil {
		return err
	}

	c.logger.Info("Connected", zap.String("Bot Name", bot.Self.FirstName))

	bot.Debug = c.telegramBot.config.Debug
	c.telegramBot.bot = bot
	c.telegramBot.router = NewCommandRouter()
	c.telegramBot.router.Register("reply", c.telegramBot.Reply)
	c.telegramBot.router.Register("help", c.telegramBot.Help)

	// User handlers
	c.telegramBot.router.Register("user", c.telegramBot.GetUser)
	c.telegramBot.router.Register("user_add", c.telegramBot.CreateUser)
	c.telegramBot.router.Register("user_del", c.telegramBot.DeleteUser, c.telegramBot.RequireAdmin)

	// Billing handlers
	c.telegramBot.router.Register("billing", c.telegramBot.GetBilling)
	c.telegramBot.router.Register("billing_list", c.telegramBot.ListBillings)
	c.telegramBot.router.Register("billing_add", c.telegramBot.CreateBilling, c.telegramBot.RequireAdmin)
	c.telegramBot.router.Register("billing_del", c.telegramBot.DeleteBilling, c.telegramBot.RequireAdmin)

	// Payment handlers
	c.telegramBot.router.Register("payment_associate", c.telegramBot.AssociatePayment, c.telegramBot.RequireAdmin)
	c.telegramBot.router.Register("payment_disassociate", c.telegramBot.DisassociatePayment, c.telegramBot.RequireAdmin)
	c.telegramBot.router.Register("billing_pay", c.telegramBot.PayBilling)
	c.telegramBot.router.Register("billing_unpay", c.telegramBot.UnpayBilling)
	c.telegramBot.router.Register("billing_pay_admin", c.telegramBot.PayBillingAdmin, c.telegramBot.RequireAdmin)
	c.telegramBot.router.Register("billing_unpay_admin", c.telegramBot.UnpayBillingAdmin, c.telegramBot.RequireAdmin)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			if update.Message != nil {
				c.telegramBot.Handle(update.Message)
				c.logger.Sugar().Infof("%v", update.Message.Entities)
			}
		}
	}()

	return nil
}
