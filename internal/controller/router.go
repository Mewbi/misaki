package controller

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CommandHandler defines the signature for command handler functions.
type CommandHandler func(context.Context, *tgbotapi.Message)

// CommandRouter maps commands to handlers.
type CommandRouter struct {
	handlers map[string]CommandHandler
}

// NewCommandRouter creates a new CommandRouter.
func NewCommandRouter() *CommandRouter {
	return &CommandRouter{handlers: make(map[string]CommandHandler)}
}

// Register adds a command and its handler to the router.
func (r *CommandRouter) Register(command string, handler CommandHandler) {
	r.handlers[command] = handler
}
