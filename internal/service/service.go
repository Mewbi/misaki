package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"misaki/internal/repository"
	"misaki/types"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service struct {
	logger     *zap.Logger
	repository repository.Repository
}

func NewService(logger *zap.Logger, repo repository.Repository) *Service {
	return &Service{
		logger:     logger,
		repository: repo,
	}
}

func (s *Service) CreateUser(ctx context.Context, user *types.User) (*types.User, error) {
	userFound, err := s.repository.GetUser(ctx, user)
	if err != sql.ErrNoRows {
		if userFound != nil {
			return nil, fmt.Errorf("user already exist")
		}

		return nil, err
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	user.UserID = userID
	user.CreatedAt = time.Now()

	if err := s.repository.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) GetUser(ctx context.Context, user *types.User) (*types.User, error) {
	if user.UserID == uuid.Nil && user.TelegramID <= 0 {
		return nil, fmt.Errorf("missing identifiers to search user")
	}

	return s.repository.GetUser(ctx, user)
}

func (s *Service) DeleteUser(ctx context.Context, user *types.User) error {
	if user.UserID == uuid.Nil && user.TelegramID <= 0 {
		return fmt.Errorf("missing identifiers to delete user")
	}
	return s.repository.DeleteUser(ctx, user)
}

func (s *Service) IsUserAdmin(ctx context.Context, user *types.User) (bool, error) {
	user, err := s.GetUser(ctx, user)
	if err != nil {
		return false, err
	}

	return user.Admin, nil
}
