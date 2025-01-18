package service

import (
	"misaki/internal/repository"

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

func (s *Service) Something() {
	log := s.logger.Sugar()
	log.Infow("Im on service now!")
	s.repository.Something()
}
