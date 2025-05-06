package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/uptrace/bun"
	"log/slog"
	"supmap-users/internal/models"
)

type Interactions struct {
	log *slog.Logger
	bun *bun.DB
}

func NewInteractions(db *bun.DB, log *slog.Logger) *Interactions {
	return &Interactions{
		log: log,
		bun: db,
	}
}

func (i *Interactions) InsertTx(ctx context.Context, exec bun.IDB, interaction *models.Interaction) error {
	_, err := exec.NewInsert().
		Model(interaction).
		Returning("id").
		Exec(ctx)
	return err
}

func (i *Interactions) FindInteractionByIdTx(ctx context.Context, exec bun.IDB, id int64) (*models.Interaction, error) {
	var interaction models.Interaction

	err := exec.NewSelect().
		Model(&interaction).
		Relation("Incident").
		Relation("Incident.Type").
		Relation("Incident.Interactions").
		Where("ii.id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &interaction, nil
}

func (i *Interactions) FindInteractionById(ctx context.Context, id int64) (*models.Interaction, error) {
	return i.FindInteractionByIdTx(ctx, i.bun, id)
}
