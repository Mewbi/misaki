package repository

import (
	"context"

	"misaki/types"
)

type Repository interface {
	CreateUser(ctx context.Context, user *types.User) error
	GetUser(ctx context.Context, user *types.User) (*types.User, error)
	DeleteUser(ctx context.Context, user *types.User) error
}
