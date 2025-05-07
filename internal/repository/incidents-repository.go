package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/uptrace/bun"
	"log/slog"
	"supmap-users/internal/models"
	"supmap-users/internal/models/dto"
	"time"
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

func (i *Incidents) GetAllActive(ctx context.Context, exec bun.IDB) ([]models.Incident, error) {
	var types []models.Incident
	err := exec.NewSelect().
		Model(&types).
		Relation("Type").
		Relation("Interactions").
		Where("deleted_at IS NULL").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return types, nil
}

func (i *Incidents) FindAllIncidentTypes(ctx context.Context) ([]models.Type, error) {
	var types []models.Type
	err := i.bun.NewSelect().
		Model(&types).
		Order("id ASC").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return types, nil
}

func (i *Incidents) FindIncidentTypeById(ctx context.Context, id *int64) (*models.Type, error) {
	var incidentType models.Type
	query := i.bun.NewSelect().
		Model(&incidentType)
	if id != nil {
		query = query.Where("id = ?", id)
	}

	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &incidentType, nil
}

func (i *Incidents) FindIncidentByIdTx(ctx context.Context, exec bun.IDB, id int64) (*models.Incident, error) {
	var incident models.Incident

	// Verrouille l'incident en cours de récupération avant de récupérer les relations
	err := exec.NewSelect().
		Model(&incident).
		Where("i.id = ?", id).
		For("UPDATE"). // Verrouille l'incident
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Charge les relations après avoir verrouillé l'incident
	err = exec.NewSelect().
		Model(&incident).
		Relation("Type").
		Relation("Interactions"). // Charger les relations après
		Where("i.id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return &incident, nil
}

func (i *Incidents) FindIncidentById(ctx context.Context, id int64) (*models.Incident, error) {
	return i.FindIncidentByIdTx(ctx, i.bun, id)
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

func (i *Incidents) FindIncidentsInZone(ctx context.Context, lat, lon *float64, radius int64, typeId *int64) ([]models.IncidentWithDistance, error) {
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
				AND longitude BETWEEN (? - ? / 111000.0) AND (? + ? / 111000.0)
		`

	args := []any{
		lat, lon, lat,
		lat, radius, lat, radius,
		lon, radius, lon, radius,
	}

	if typeId != nil {
		rawSQL += " AND type_id = ?"
		args = append(args, typeId)
	}

	rawSQL += `
		) AS sub
		WHERE distance <= ?
		ORDER BY distance;
		`
	args = append(args, radius)

	err := i.bun.NewRaw(rawSQL, args...).Scan(ctx, &incidents)

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

func (i *Incidents) UpdateIncidentTx(ctx context.Context, exec bun.IDB, incident *models.Incident) error {
	incident.UpdatedAt = time.Now()

	_, err := exec.NewUpdate().
		Model(incident).
		Where("id = ?", incident.ID).
		OmitZero().
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}
