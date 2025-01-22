package controller

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"misaki/config"
	"misaki/internal/service"
	"misaki/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type telegramBot struct {
	logger  *zap.Logger
	config  *config.Telegram
	service *service.Service
	bot     *tgbotapi.BotAPI
	router  *CommandRouter
}

func NewTelegramBot(config *config.Config, logger *zap.Logger, s *service.Service) *telegramBot {
	return &telegramBot{
		logger:  logger,
		config:  &config.Telegram,
		service: s,
	}
}

func (b *telegramBot) Handle(message *tgbotapi.Message) {
	if handler, ok := b.router.handlers[message.Command()]; ok {
		b.logger.Info("Running command", zap.String("command", message.Command()))
		handler(context.Background(), message)
		return
	}
	b.logger.Info("Unknown command", zap.String("command", message.Command()))
}

func (b *telegramBot) Reply(ctx context.Context, m *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(m.Chat.ID, m.Text)
	msg.ReplyToMessageID = m.MessageID
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
	return
}

func (b *telegramBot) CreateUser(ctx context.Context, m *tgbotapi.Message) {
	name := fmt.Sprintf("%s %s", m.From.FirstName, m.From.LastName)
	newUser := types.User{
		TelegramID:   m.From.ID,
		TelegramName: name,
	}

	user, err := b.service.CreateUser(ctx, &newUser)
	if err != nil {

		b.logger.Error("error creating user", zap.Int64("TelegramID", newUser.TelegramID), zap.String("Name", newUser.TelegramName), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Error creating user %s (%d)", newUser.TelegramName, newUser.TelegramID))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	messageText := fmt.Sprintf(
		"🎉 *User Created Successfully!*\n\n"+
			"👤 *User Details:*\n"+
			"🆔 *ID:* `%s`\n"+
			"🌎 *Telegram ID:* `%d`\n"+
			"💬 *Telegram Name:* %s\n"+
			"👮 *Admin:* %t\n"+
			"📅 *Created At:* %s\n",
		user.UserID.String(),
		user.TelegramID,
		user.TelegramName,
		user.Admin,
		user.CreatedAt.Format("2006-01-02 15:04:05"),
	)

	// Send the response message
	msg := tgbotapi.NewMessage(m.Chat.ID, messageText)
	msg.ParseMode = "Markdown"

	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}

func (b *telegramBot) GetUser(ctx context.Context, m *tgbotapi.Message) {
	user := types.User{}
	id := m.CommandArguments()

	telegramID, errTg := strconv.ParseInt(id, 10, 64)
	if errTg == nil {
		user.TelegramID = telegramID
	}

	userID, errUuid := uuid.Parse(id)
	if errUuid == nil {
		user.UserID = userID
	}

	// ID is empty, use ID from user that sent the message
	if id == "" {
		user.TelegramID = m.From.ID
		errTg = nil
		errUuid = nil
	}

	if errTg != nil && errUuid != nil {

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Error to get user, invalid id informed: %s", id))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	userFound, err := b.service.GetUser(ctx, &user)
	if err != nil {
		b.logger.Error("failed to get user", zap.String("ID", id), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Internal error while getting user: %s", id))
		if err == sql.ErrNoRows {
			msg = tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("User %s not found", id))
		}

		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	messageText := fmt.Sprintf(
		"👤 *User Details:*\n"+
			"🆔 *ID:* `%s`\n"+
			"🌎 *Telegram ID:* `%d`\n"+
			"💬 *Telegram Name:* %s\n"+
			"👮 *Admin:* %t\n"+
			"📅 *Created At:* %s\n",
		userFound.UserID.String(),
		userFound.TelegramID,
		userFound.TelegramName,
		userFound.Admin,
		userFound.CreatedAt.Format("2006-01-02 15:04:05"),
	)

	msg := tgbotapi.NewMessage(m.Chat.ID, messageText)
	msg.ParseMode = "Markdown"
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
	return
}

func (b *telegramBot) DeleteUser(ctx context.Context, m *tgbotapi.Message) {
	user := types.User{}
	id := m.CommandArguments()

	telegramID, errTg := strconv.ParseInt(id, 10, 64)
	if errTg == nil {
		user.TelegramID = telegramID
	}

	userID, errUuid := uuid.Parse(id)
	if errUuid == nil {
		user.UserID = userID
	}

	if errTg != nil && errUuid != nil {

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Error to get user, invalid id informed: %s", id))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	admin, err := b.service.IsUserAdmin(ctx, &types.User{TelegramID: m.From.ID})
	if err != nil {
		b.logger.Error("error validating user permission", zap.Int64("TelegramID", m.From.ID), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Error validating user permission"))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	if !admin {

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("User don't have required permission"))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	if err := b.service.DeleteUser(ctx, &user); err != nil {

		b.logger.Error("error deleting user", zap.Int64("TelegramID", user.TelegramID), zap.String("UserID", user.UserID.String()), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Error deleting user %s", id))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("👤 *User Deleted:* %s", id))
	msg.ReplyToMessageID = m.MessageID
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}
