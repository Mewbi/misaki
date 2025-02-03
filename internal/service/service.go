package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func (s *Service) GetBilling(ctx context.Context, billing *types.Billing) (*types.Billing, error) {
	if billing.ID == uuid.Nil && billing.Name == "" {
		return nil, fmt.Errorf("missing billing id")
	}

	billing, err := s.repository.GetBilling(ctx, billing)
	if err != nil {
		return nil, err
	}

	if len(billing.Payments) > 0 {
		billing.ValuePerUser = billing.Value / float64(len(billing.Payments))
	}

	return billing, nil
}

func (s *Service) ListBillings(ctx context.Context) ([]*types.Billing, error) {
	return s.repository.ListBillings(ctx)
}

func (s *Service) CreateBilling(ctx context.Context, billing *types.Billing) (*types.Billing, error) {
	// Check invalid names
	if len(billing.Name) == 0 || strings.Contains(billing.Name, " ") {
		return nil, fmt.Errorf("invalid name informed: %s", billing.Name)
	}

	// Check billing with same name exist
	copyVal := *billing
	billingFound, err := s.repository.GetBilling(ctx, &copyVal)
	if err != sql.ErrNoRows {
		if billingFound != nil {
			return nil, fmt.Errorf("billing already exist")
		}

		return nil, err
	}

	billing.ID, err = uuid.NewV7()
	if err != nil {
		return nil, err
	}
	billing.CreatedAt = time.Now()

	if err := s.repository.CreateBilling(ctx, billing); err != nil {
		return nil, err
	}

	return billing, nil
}

func (s *Service) DeleteBilling(ctx context.Context, billing *types.Billing) error {
	if billing.ID == uuid.Nil && billing.Name == "" {
		return fmt.Errorf("missing identifiers to delete user")
	}
	return s.repository.DeleteBilling(ctx, billing)
}

func (s *Service) ChangePaymentAssociation(ctx context.Context, payment *types.Payment, assoaciate bool) error {
	if assoaciate {
		payment.Paid = false
		return s.repository.AssociatePayment(ctx, payment)
	}
	return s.repository.DisassociatePayment(ctx, payment)
}

func (s *Service) ChangePaymentStatus(ctx context.Context, payment *types.Payment) error {
	if payment.BillingID == uuid.Nil || payment.UserID == uuid.Nil {
		return fmt.Errorf("missing billing or user identifier")
	}

	if payment.Paid {
		payment.PaidAt = time.Now()
	}

	return s.repository.ChangePaymentStatus(ctx, payment)
}

func (s *Service) PaymentAssociationExist(ctx context.Context, payment *types.Payment) (bool, error) {
	searchPayment := *payment
	_, err := s.repository.GetPaymentAssociation(ctx, &searchPayment)
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
