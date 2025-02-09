package types

import (
	"bytes"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/wader/goutubedl"
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

type Billing struct {
	ID           uuid.UUID
	Name         string
	Value        float64
	CreatedAt    time.Time
	ValuePerUser float64
	Payments     []Payment
}

type Payment struct {
	BillingID uuid.UUID
	UserID    uuid.UUID
	Paid      bool
	PaidAt    time.Time
	UserInfo  User
}

type Midia struct {
	Quality   string
	OnlyAudio bool
	Url       *url.URL
	Name      string
	Content   *bytes.Buffer
	Data      goutubedl.Result
}
