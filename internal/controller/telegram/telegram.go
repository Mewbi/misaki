package telegram

import (
	"context"
	"fmt"
	"strings"

	"misaki/config"
	"misaki/internal/service"
	"misaki/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type TelegramBot struct {
	logger  *zap.Logger
	config  *config.Telegram
	service *service.Service
	Bot     *tgbotapi.BotAPI
	router  *CommandRouter
}

func NewTelegramBot(config *config.Config, logger *zap.Logger, s *service.Service) *TelegramBot {
	return &TelegramBot{
		logger:  logger,
		config:  &config.Telegram,
		service: s,
	}
}

func (b *TelegramBot) StartBot() error {
	bot, err := tgbotapi.NewBotAPI(b.config.Token)
	if err != nil {
		return err
	}

	b.logger.Info("Connected", zap.String("Bot Name", bot.Self.FirstName))

	bot.Debug = b.config.Debug
	b.Bot = bot

	return nil
}

func (b *TelegramBot) Handle(message *tgbotapi.Message) {
	if endpoint, ok := b.router.handlers[message.Command()]; ok {
		b.logger.Info("Running command", zap.String("command", message.Command()))
		ctx := context.Background()

		for _, handler := range endpoint.Middlewares {
			if pass := handler(ctx, message); !pass {
				return
			}
		}

		endpoint.Handler(ctx, message)
		return
	}
	b.logger.Info("Unknown command", zap.String("command", message.Command()))
}

func (b *TelegramBot) RequireAdmin(ctx context.Context, m *tgbotapi.Message) bool {
	admin, err := b.service.IsUserAdmin(ctx, &types.User{TelegramID: m.From.ID})
	if err != nil {
		b.logger.Error("error validating user permission", zap.Int64("TelegramID", m.From.ID), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error validating user permission"))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.Bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return false
	}

	if !admin {

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ User don't have required permission"))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.Bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}

		return false
	}

	return true
}

func (b *TelegramBot) Reply(ctx context.Context, m *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(m.Chat.ID, m.Text)
	msg.ReplyToMessageID = m.MessageID
	if _, err := b.Bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}

func (b *TelegramBot) Help(ctx context.Context, m *tgbotapi.Message) {
	messageText := fmt.Sprintf(
		"📝 <b>Commands List</b>\n\n"+
			"%s",
		strings.Join(b.router.commands, "\n"))
	msg := tgbotapi.NewMessage(m.Chat.ID, messageText)
	msg.ReplyToMessageID = m.MessageID
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := b.Bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}
