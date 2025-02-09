package telegram

import (
	"context"
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"misaki/types"

	"go.uber.org/zap"
)

func (b *TelegramBot) CreateUser(ctx context.Context, m *tgbotapi.Message) {
	name := fmt.Sprintf("%s %s", m.From.FirstName, m.From.LastName)
	newUser := types.User{
		TelegramID:   m.From.ID,
		TelegramName: name,
	}

	// Set user admin
	newUser.Admin = m.From.ID == b.config.AdminUser

	user, err := b.service.CreateUser(ctx, &newUser)
	if err != nil {

		b.logger.Error("error creating user", zap.Int64("TelegramID", newUser.TelegramID), zap.String("Name", newUser.TelegramName), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error creating user %s (%d)", newUser.TelegramName, newUser.TelegramID))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.Bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	messageText := fmt.Sprintf(
		"🎉 *User Created Successfully!*\n\n"+
			"👤 *User Details:*\n"+
			"🆔 *ID:* `%s`\n"+
			"🌎 *Telegram ID:* `%d`\n"+
			"💬 *Telegram Name:* `%s`\n"+
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
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.Bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}

func (b *TelegramBot) GetUser(ctx context.Context, m *tgbotapi.Message) {
	id := m.CommandArguments()

	user, err := b.parseUserIdentifier(id)

	// ID is empty, use ID from user that sent the message
	if id == "" {
		user = &types.User{
			TelegramID: m.From.ID,
		}
	} else if err != nil {

		b.logger.Info("invalid user identifier informed", zap.String("id", id))
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get user, invalid id informed: %s", id))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.Bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	userFound, err := b.service.GetUser(ctx, user)
	if err != nil {
		b.logger.Error("failed to get user", zap.String("ID", id), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Internal error while getting user: %s", id))
		if err == sql.ErrNoRows {
			msg = tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ User %s not found", id))
		}

		if _, err := b.Bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	messageText := fmt.Sprintf(
		"👤 *User Details:*\n"+
			"🆔 *ID:* `%s`\n"+
			"🌎 *Telegram ID:* `%d`\n"+
			"💬 *Telegram Name:* `%s`\n"+
			"👮 *Admin:* %t\n"+
			"📅 *Created At:* %s\n",
		userFound.UserID.String(),
		userFound.TelegramID,
		userFound.TelegramName,
		userFound.Admin,
		userFound.CreatedAt.Format("2006-01-02 15:04:05"),
	)

	msg := tgbotapi.NewMessage(m.Chat.ID, messageText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.Bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
	return
}

func (b *TelegramBot) DeleteUser(ctx context.Context, m *tgbotapi.Message) {
	id := m.CommandArguments()
	user, err := b.parseUserIdentifier(id)

	// ID is empty, use ID from user that sent the message
	if id == "" {
		user.TelegramID = m.From.ID
	} else if err != nil {

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get user, invalid id informed: %s", id))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.Bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	if err := b.service.DeleteUser(ctx, user); err != nil {

		b.logger.Error("error deleting user", zap.Int64("TelegramID", user.TelegramID), zap.String("UserID", user.UserID.String()), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error deleting user %s", id))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.Bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("👤 *User Deleted:* %s", id))
	msg.ReplyToMessageID = m.MessageID
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.Bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}
