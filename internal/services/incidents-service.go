package services

import (
	"context"
	"log/slog"
	"supmap-users/internal/config"
	"supmap-users/internal/models"
	"supmap-users/internal/repository"
)

type Service struct {
	log    *slog.Logger
	config *config.Config
	incidents *repository.Incidents
}

func NewService(log *slog.Logger, config *config.Config, incidents *repository.Incidents) *Service {
	return &Service{
		log: log,
		config: config,
		incidents: incidents,
	}
}
func (s *Service) GetIncidents(ctx context.Context) ([]models.Incident, error) {
	return s.incidents.GetAllIncidents(ctx)
}
