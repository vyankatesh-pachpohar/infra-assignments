package service

import (
	"context"
	"errors"

	"github.com/vyankatesh-pachpohar/infra-assignments/internal/domain"
	"github.com/vyankatesh-pachpohar/infra-assignments/internal/repository"
)

var ErrInvalidConfig = errors.New("invalid config: host, app_name, and port are required")

type ConfigService struct {
	repo *repository.ConfigRepository
}

func NewConfigService(repo *repository.ConfigRepository) *ConfigService {
	return &ConfigService{repo: repo}
}

func (s *ConfigService) HealthCheck(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *ConfigService) GetConfig(ctx context.Context, id string) (*domain.Config, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ConfigService) SaveConfig(ctx context.Context, c *domain.Config) error {
	if c.ID == "" || c.Host == "" || c.AppName == "" || c.Port == 0 {
		return ErrInvalidConfig
	}
	if c.LogLevel == "" {
		c.LogLevel = "INFO" // sensible default per spec
	}
	return s.repo.Upsert(ctx, c)
}
