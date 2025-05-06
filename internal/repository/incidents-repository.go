package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/uptrace/bun"
	"log/slog"
	"supmap-users/internal/models"
	"supmap-users/internal/models/dto"
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

func (i *Incidents) GetTypeById(ctx context.Context, id int64) (*models.Type, error) {
	var incidentType models.Type
	err := i.bun.NewSelect().
		Model(&incidentType).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &incidentType, nil
}

func (i *Incidents) FindIncidentById(ctx context.Context, id int64) (*models.Incident, error) {
	var incident models.Incident

	err := i.bun.NewSelect().
		Model(&incident).
		Relation("Type").
		Relation("Interactions").
		Where("i.id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &incident, nil
}

func (i *Incidents) FindUserHistory(ctx context.Context, user *dto.PartialUserDTO) ([]models.Incident, error) {
	var incidents []models.Incident

	err := i.bun.NewSelect().
		Model(&incidents).
		Relation("Type").
		Relation("Interactions").
		Where("i.user_id = ?", user.ID).
		Where("i.deleted_at IS NOT NULL").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return incidents, err
}

func (i *Incidents) GetLastUserIncident(ctx context.Context, user *dto.PartialUserDTO) (*models.Incident, error) {
	var incident models.Incident
	err := i.bun.NewSelect().
		Model(&incident).
		Relation("Type").
		Relation("Interactions").
		Where("i.user_id = ?", user.ID).
		OrderExpr("i.created_at DESC").
		Limit(1).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &incident, nil
}

func (i *Incidents) FindIncidentsInZone(ctx context.Context, lat, lng *float64, radius int64, typeId *int64) ([]models.IncidentWithDistance, error) {
	var incidents []models.IncidentWithDistance

	rawSQL := `
		SELECT *
		FROM (SELECT *,
					 (
						 6371000 * acos(
								 cos(radians(?)) * cos(radians(latitude)) *
								 cos(radians(longitude) - radians(?)) +
								 sin(radians(?)) * sin(radians(latitude))
								   )
						 ) AS distance
			  FROM incidents
			  WHERE deleted_at IS NULL
				AND latitude BETWEEN (? - ? / 111000.0) AND (? + ? / 111000.0)
				AND longitude BETWEEN (? - ? / 111000.0) AND (? + ? / 111000.0)) AS sub
		WHERE distance <= ?
		ORDER BY distance;
		`

	err := i.bun.NewRaw(rawSQL,
		lat, lng, lat,
		lat, radius, lat, radius,
		lng, radius, lng, radius,
		radius,
	).Scan(ctx, &incidents)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return incidents, nil
}

func (i *Incidents) CreateIncident(ctx context.Context, incident *models.Incident) error {
	if _, err := i.bun.NewInsert().Model(incident).Returning("id").Exec(ctx); err != nil {
		return err
	}
	return nil
}
