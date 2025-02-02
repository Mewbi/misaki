package controller

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"misaki/config"
	"misaki/internal/service"
	"misaki/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func (b *telegramBot) RequireAdmin(ctx context.Context, m *tgbotapi.Message) bool {
	admin, err := b.service.IsUserAdmin(ctx, &types.User{TelegramID: m.From.ID})
	if err != nil {
		b.logger.Error("error validating user permission", zap.Int64("TelegramID", m.From.ID), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error validating user permission"))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return false
	}

	if !admin {

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ User don't have required permission"))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}

		return false
	}

	return true
}

func (b *telegramBot) Reply(ctx context.Context, m *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(m.Chat.ID, m.Text)
	msg.ReplyToMessageID = m.MessageID
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}

func (b *telegramBot) Help(ctx context.Context, m *tgbotapi.Message) {
	messageText := fmt.Sprintf(
		"📝 <b>Commands List</b>\n\n"+
			"%s",
		strings.Join(b.router.commands, "\n"))
	msg := tgbotapi.NewMessage(m.Chat.ID, messageText)
	msg.ReplyToMessageID = m.MessageID
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}

func (b *telegramBot) CreateUser(ctx context.Context, m *tgbotapi.Message) {
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

	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}

func (b *telegramBot) GetUser(ctx context.Context, m *tgbotapi.Message) {
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
		if _, err := b.bot.Send(msg); err != nil {
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

		if _, err := b.bot.Send(msg); err != nil {
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
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
	return
}

func (b *telegramBot) DeleteUser(ctx context.Context, m *tgbotapi.Message) {
	id := m.CommandArguments()
	user, err := b.parseUserIdentifier(id)

	// ID is empty, use ID from user that sent the message
	if id == "" {
		user.TelegramID = m.From.ID
	} else if err != nil {

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get user, invalid id informed: %s", id))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	if err := b.service.DeleteUser(ctx, user); err != nil {

		b.logger.Error("error deleting user", zap.Int64("TelegramID", user.TelegramID), zap.String("UserID", user.UserID.String()), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error deleting user %s", id))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("👤 *User Deleted:* %s", id))
	msg.ReplyToMessageID = m.MessageID
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}

func (b *telegramBot) GetBilling(ctx context.Context, m *tgbotapi.Message) {
	id := m.CommandArguments()

	billing, err := b.parseBillingIdentifier(id)
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get billing: %s", err.Error()))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	billing, err = b.service.GetBilling(ctx, billing)
	if err != nil {
		b.logger.Error("failed to get billing", zap.String("ID", id), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Internal error while getting billing: %s", id))
		if err == sql.ErrNoRows {
			msg = tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Billing %s not found", id))
		}

		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	messageText := fmt.Sprintf(
		"🤑 *Billing Details*\n\n"+
			"🆔 *ID:* `%s`\n"+
			"💬 *Name:* `%s`\n"+
			"👤 *Users Associated:* %d\n"+
			"💸 *Value:* %.2f\n"+
			"💸 *Value per User:* %.2f\n"+
			"📅 *Created At:* %s\n",
		billing.ID.String(),
		billing.Name,
		len(billing.Payments),
		billing.Value,
		billing.ValuePerUser,
		billing.CreatedAt.Format("2006-01-02 15:04:05"),
	)

	msg := tgbotapi.NewMessage(m.Chat.ID, messageText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
	return
}

func (b *telegramBot) ListBillings(ctx context.Context, m *tgbotapi.Message) {
	billings, err := b.service.ListBillings(ctx)
	if err != nil {
		b.logger.Error("failed to list billings", zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Internal error while getting billings")

		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	messageText := fmt.Sprintf(
		"🤑 *Billings Found:* %d\n\n",
		len(billings),
	)

	for _, billing := range billings {
		text := fmt.Sprintf("🆔 `%s` \n💬 `%s` \n💸 %.2f \n\n",
			billing.ID,
			billing.Name,
			billing.Value,
		)

		messageText += text
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, messageText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
	return
}

func (b *telegramBot) CreateBilling(ctx context.Context, m *tgbotapi.Message) {
	data := strings.Split(m.CommandArguments(), " ")

	if len(data) != 2 {
		b.logger.Error("invalid billing arguments", zap.Int("number arguments", len(data)))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Invalid number of arguments received, expected: <name> <value>")
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	name := data[0]
	value, err := strconv.ParseFloat(data[1], 64)
	if err != nil {
		b.logger.Error("invalid billing value", zap.String("value", data[1]), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Invalid value for billing, expected float, received: %s", data[1]))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	newBilling := &types.Billing{
		Name:  name,
		Value: value,
	}

	billing, err := b.service.CreateBilling(ctx, newBilling)
	if err != nil {

		b.logger.Error("error creating billing", zap.String("name", newBilling.Name), zap.Float64("value", newBilling.Value), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error creating billing %s (%f)", newBilling.Name, newBilling.Value))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	messageText := fmt.Sprintf(
		"🤑 *Billing Created Successfully!*\n\n"+
			"👤 *Billing Details:*\n"+
			"🆔 *ID:* `%s`\n"+
			"💬 *Name:* `%s`\n"+
			"💸 *Value:* %.2f\n"+
			"📅 *Created At:* %s\n",
		billing.ID.String(),
		billing.Name,
		billing.Value,
		billing.CreatedAt.Format("2006-01-02 15:04:05"),
	)

	// Send the response message
	msg := tgbotapi.NewMessage(m.Chat.ID, messageText)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
}

func (b *telegramBot) DeleteBilling(ctx context.Context, m *tgbotapi.Message) {
	id := m.CommandArguments()

	billing, err := b.parseBillingIdentifier(id)
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get billing: %s", err.Error()))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	err = b.service.DeleteBilling(ctx, billing)
	if err != nil {
		b.logger.Error("failed to delete billing", zap.String("ID", id), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Internal error while deleting billing: %s", id))

		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("🤑 *Billing Deleted:* %s", id))
	msg.ReplyToMessageID = m.MessageID
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}

	return
}

func (b *telegramBot) AssociatePayment(ctx context.Context, m *tgbotapi.Message) {
	b.changePaymentAssociation(ctx, m, true)
}

func (b *telegramBot) DisassociatePayment(ctx context.Context, m *tgbotapi.Message) {
	b.changePaymentAssociation(ctx, m, false)
}

func (b *telegramBot) changePaymentAssociation(ctx context.Context, m *tgbotapi.Message, associate bool) {
	data := strings.Split(m.CommandArguments(), " ")

	if len(data) != 2 {
		b.logger.Error("invalid payment association", zap.Int("number arguments", len(data)))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Invalid number of arguments received, expected: <billing-identifier> <user-identifier>")
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}
	// Parse billing
	billing, err := b.parseBillingIdentifier(data[0])
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get billing: %s", err.Error()))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	// Parse user
	user, err := b.parseUserIdentifier(data[1])
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get user, invalid id informed: %s", data[1]))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	// Search billing
	billing, err = b.service.GetBilling(ctx, billing)
	if err != nil {
		b.logger.Error("failed to get billing", zap.String("ID", data[0]), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Internal error while getting billing: %s", data[0]))
		if err == sql.ErrNoRows {
			msg = tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Billing %s not found", data[0]))
		}

		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	// Search user
	user, err = b.service.GetUser(ctx, user)
	if err != nil {
		b.logger.Error("failed to get user", zap.String("ID", data[1]), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Internal error while getting user: %s", data[1]))
		if err == sql.ErrNoRows {
			msg = tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ User %s not found", data[1]))
		}

		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	payment := &types.Payment{
		BillingID: billing.ID,
		UserID:    user.UserID,
	}

	if err := b.service.ChangePaymentAssociation(ctx, payment, associate); err != nil {

		b.logger.Error("failed to change association", zap.Bool("Associate", associate), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Internal error while changing payment association")
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	associateText := "Association"
	if !associate {
		associateText = "Disassociation"
	}

	messageText := fmt.Sprintf(
		"🔄 *Billing %s*\n\n"+
			"👤 *User ID:* `%s`\n"+
			"💸 *Billing ID:* `%s`\n",
		associateText,
		payment.UserID,
		payment.BillingID,
	)

	msg := tgbotapi.NewMessage(m.Chat.ID, messageText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
	return
}

func (b *telegramBot) PayBilling(ctx context.Context, m *tgbotapi.Message) {
	id := m.CommandArguments()
	// Parse billing
	billing, err := b.parseBillingIdentifier(id)
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get billing: %s", err.Error()))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	user := &types.User{
		TelegramID: m.From.ID,
	}
	b.changePaymentStatus(ctx, m, billing, user, true)
}

func (b *telegramBot) UnpayBilling(ctx context.Context, m *tgbotapi.Message) {
	id := m.CommandArguments()
	// Parse billing
	billing, err := b.parseBillingIdentifier(id)
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get billing: %s", err.Error()))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	user := &types.User{
		TelegramID: m.From.ID,
	}
	b.changePaymentStatus(ctx, m, billing, user, false)
}

func (b *telegramBot) PayBillingAdmin(ctx context.Context, m *tgbotapi.Message) {
	data := strings.Split(m.CommandArguments(), " ")

	if len(data) != 2 {
		b.logger.Error("invalid payment association", zap.Int("number arguments", len(data)))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Invalid number of arguments received, expected: <billing-identifier> <user-identifier>")
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	// Parse billing
	billing, err := b.parseBillingIdentifier(data[0])
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get billing: %s", err.Error()))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	// Parse user
	user, err := b.parseUserIdentifier(data[1])
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get user, invalid id informed: %s", data[1]))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	b.changePaymentStatus(ctx, m, billing, user, true)
}

func (b *telegramBot) UnpayBillingAdmin(ctx context.Context, m *tgbotapi.Message) {
	data := strings.Split(m.CommandArguments(), " ")

	if len(data) != 2 {
		b.logger.Error("invalid payment association", zap.Int("number arguments", len(data)))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Invalid number of arguments received, expected: <billing-identifier> <user-identifier>")
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	// Parse billing
	billing, err := b.parseBillingIdentifier(data[0])
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get billing: %s", err.Error()))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	// Parse user
	user, err := b.parseUserIdentifier(data[1])
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to get user, invalid id informed: %s", data[1]))
		msg.ReplyToMessageID = m.MessageID
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	b.changePaymentStatus(ctx, m, billing, user, false)
}

func (b *telegramBot) changePaymentStatus(ctx context.Context, m *tgbotapi.Message, billing *types.Billing, user *types.User, status bool) {
	// Search billing
	billing, err := b.service.GetBilling(ctx, billing)
	if err != nil {
		b.logger.Error("failed to get billing", zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Internal error while getting billing")
		if err == sql.ErrNoRows {
			msg = tgbotapi.NewMessage(m.Chat.ID, "⚠️ Billing not found")
		}

		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	// Search user
	user, err = b.service.GetUser(ctx, user)
	if err != nil {
		b.logger.Error("failed to get user", zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Internal error while getting user")
		if err == sql.ErrNoRows {
			msg = tgbotapi.NewMessage(m.Chat.ID, "⚠️ User %s not found")
		}

		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	payment := &types.Payment{
		BillingID: billing.ID,
		UserID:    user.UserID,
		Paid:      status,
	}

	// Payment Exist
	exist, err := b.service.PaymentAssociationExist(ctx, payment)
	if err != nil {
		b.logger.Error("failed to get check payment association", zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Internal error while checking payment association")

		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	if !exist {
		b.logger.Error("payment association not found", zap.String("BillingID", payment.BillingID.String()), zap.String("UserID", payment.UserID.String()))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Payment association not found")

		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	if err := b.service.ChangePaymentStatus(ctx, payment); err != nil {

		b.logger.Error("failed to change association", zap.Bool("Paid", status), zap.Error(err))

		msg := tgbotapi.NewMessage(m.Chat.ID, "⚠️ Internal error while changing payment association")
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	paidText := "💵 *Status*: Unpaid"
	if status {
		paidText = fmt.Sprintf(
			"💵 *Status*: Paid\n"+
				"📅 *Paid At*: %s",
			payment.PaidAt.Format("2006-01-02 15:04:05"),
		)
	}

	messageText := fmt.Sprintf(
		"🔄 *Billing Payment\n\n"+
			"👤 *User ID:* `%s`\n"+
			"💸 *Billing ID:* `%s`\n"+
			"%s\n",
		payment.UserID,
		payment.BillingID,
		paidText,
	)

	msg := tgbotapi.NewMessage(m.Chat.ID, messageText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
	return
}
