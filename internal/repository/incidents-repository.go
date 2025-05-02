package repository

import (
	"context"
	"github.com/uptrace/bun"
	"log/slog"
	"supmap-users/internal/models"
)

type Incidents struct {
	log *slog.Logger
	bun *bun.DB
}

func NewIncidents(db *bun.DB, log *slog.Logger) *Incidents {
	return &Incidents{
		log: log,
		bun: db,
	}
}

func (i *Incidents) GetAllIncidents(ctx context.Context) ([]models.Incident, error) {
	var incidents []models.Incident

	err := i.bun.NewSelect().
		Model(&incidents).
		Relation("Type").
		Relation("Interactions").
		Where("i.deleted_at IS NULL").
		Order("i.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return incidents, nil
}
