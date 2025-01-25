package controller

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CommandHandler defines the signature for command handler functions.
type (
	CommandHandler   func(context.Context, *tgbotapi.Message)
	MiddlwareHandler func(context.Context, *tgbotapi.Message) bool
)

type Endpoint struct {
	Middlewares []MiddlwareHandler
	Handler     CommandHandler
}

// CommandRouter maps commands to handlers.
type CommandRouter struct {
	handlers map[string]Endpoint
}

// NewCommandRouter creates a new CommandRouter.
func NewCommandRouter() *CommandRouter {
	return &CommandRouter{handlers: make(map[string]Endpoint)}
}

// Register adds a command and its handler to the router.
func (r *CommandRouter) Register(command string, handler CommandHandler, middlewares ...MiddlwareHandler) {
	r.handlers[command] = Endpoint{
		Handler:     handler,
		Middlewares: middlewares,
	}
}
