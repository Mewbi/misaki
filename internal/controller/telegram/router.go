package telegram

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
	commands []string
}

// NewCommandRouter creates a new CommandRouter.
func NewCommandRouter() *CommandRouter {
	return &CommandRouter{handlers: make(map[string]Endpoint)}
}

// register adds a command and its handler to the router.
func (r *CommandRouter) register(command string, handler CommandHandler, middlewares ...MiddlwareHandler) {
	r.handlers[command] = Endpoint{
		Handler:     handler,
		Middlewares: middlewares,
	}

	r.commands = append(r.commands, "/"+command)
}

func (b *TelegramBot) RegisterRoutes() {
	b.router = NewCommandRouter()
	b.router.register("reply", b.Reply)
	b.router.register("help", b.Help)

	// User handlers
	b.router.register("user", b.GetUser)
	b.router.register("user_add", b.CreateUser)
	b.router.register("user_del", b.DeleteUser, b.RequireAdmin)

	// Billing handlers
	b.router.register("billing", b.GetBilling)
	b.router.register("billing_list", b.ListBillings)
	b.router.register("billing_add", b.CreateBilling, b.RequireAdmin)
	b.router.register("billing_del", b.DeleteBilling, b.RequireAdmin)

	// Payment handlers
	b.router.register("payment_associate", b.AssociatePayment, b.RequireAdmin)
	b.router.register("payment_disassociate", b.DisassociatePayment, b.RequireAdmin)
	b.router.register("billing_pay", b.PayBilling)
	b.router.register("billing_unpay", b.UnpayBilling)
	b.router.register("billing_pay_admin", b.PayBillingAdmin, b.RequireAdmin)
	b.router.register("billing_unpay_admin", b.UnpayBillingAdmin, b.RequireAdmin)

	// Download handlers
	b.router.register("youtube", b.DownloadYoutubeMidia)
}
