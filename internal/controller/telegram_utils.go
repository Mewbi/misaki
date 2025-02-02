package controller

import (
	"fmt"
	"strconv"
	"strings"

	"misaki/types"

	"github.com/google/uuid"
)

func (b *telegramBot) parseUserIdentifier(id string) (*types.User, error) {
	user := &types.User{}

	telegramID, errTg := strconv.ParseInt(id, 10, 64)
	if errTg == nil {
		user.TelegramID = telegramID
	}

	userID, errUuid := uuid.Parse(id)
	if errUuid == nil {
		user.UserID = userID
	}

	if errTg != nil && errUuid != nil {
		return nil, fmt.Errorf("invalid identifier informed: %s", id)
	}

	return user, nil
}

func (b *telegramBot) parseBillingIdentifier(id string) (*types.Billing, error) {
	billing := &types.Billing{}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	billingID, err := uuid.Parse(id)
	if err == nil {
		billing.ID = billingID
	} else {
		billing.Name = id
	}

	return billing, nil
}
