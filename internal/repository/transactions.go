package repository

import (
	"context"
	"github.com/uptrace/bun"
)

func (i *Incidents) AskForTx(ctx context.Context) (*bun.Tx, error) {
	tx, err := i.bun.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}
