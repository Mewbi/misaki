package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (b *TelegramBot) DownloadYoutubeMidia(ctx context.Context, m *tgbotapi.Message) {
	url := m.CommandArguments()

	msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("📶 Downloading midia..."))
	msg.ReplyToMessageID = m.MessageID
	if _, err := b.Bot.Send(msg); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
		return
	}

	midia, err := b.service.DownloadYoutubeMidia(ctx, url)
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("⚠️ Error to dowload midia: %s", err.Error()))

		if _, err := b.Bot.Send(msg); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	if midia.OnlyAudio {
		msgMidia := tgbotapi.NewAudio(m.Chat.ID, tgbotapi.FileBytes{
			Name:  midia.Name,
			Bytes: midia.Content.Bytes(),
		})

		msgMidia.ReplyToMessageID = m.MessageID
		if _, err := b.Bot.Send(msgMidia); err != nil {
			b.logger.Error("error while sending message", zap.Error(err))
		}
		return
	}

	msgMidia := tgbotapi.NewVideo(m.Chat.ID, tgbotapi.FileBytes{
		Name:  midia.Name,
		Bytes: midia.Content.Bytes(),
	})

	msgMidia.ReplyToMessageID = m.MessageID
	if _, err := b.Bot.Send(msgMidia); err != nil {
		b.logger.Error("error while sending message", zap.Error(err))
	}
	return
}
