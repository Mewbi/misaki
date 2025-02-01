package repository

import (
	"context"

	"misaki/types"
)

type Repository interface {
	repositoryUser
	repositoryBilling
}

type repositoryUser interface {
	CreateUser(ctx context.Context, user *types.User) error
	GetUser(ctx context.Context, user *types.User) (*types.User, error)
	DeleteUser(ctx context.Context, user *types.User) error
}

type repositoryBilling interface {
	GetBilling(ctx context.Context, billing *types.Billing) (*types.Billing, error)
	ListBillings(ctx context.Context) ([]*types.Billing, error)
	CreateBilling(ctx context.Context, billing *types.Billing) error
	DeleteBilling(ctx context.Context, billing *types.Billing) error
	AssociatePayment(ctx context.Context, payment *types.Payment) error
	DisassociatePayment(ctx context.Context, payment *types.Payment) error
	ChangePaymentStatus(ctx context.Context, payment *types.Payment) error
}
