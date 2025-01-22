package types

import (
	"time"

	"github.com/google/uuid"
)

const (
	TELEGRAM_ID_EMPTY = 0
)

type User struct {
	UserID       uuid.UUID
	TelegramID   int64
	TelegramName string
	Admin        bool
	CreatedAt    time.Time
}
