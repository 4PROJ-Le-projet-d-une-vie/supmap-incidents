package models

import (
	"github.com/uptrace/bun"
)

type Type struct {
	bun.BaseModel `bun:"table:incident_types,alias:it"`

	ID                          int64  `bun:"id,pk,autoincrement"`
	Name                        string `bun:"name,notnull"`
	Description                 string `bun:"description"`
	LifetimeWithoutConfirmation int    `bun:"lifetime_without_confirmation,notnull"`
	NegativeReportsThreshold    int    `bun:"negative_reports_threshold,notnull"`
	GlobalLifetime              int    `bun:"global_lifetime,notnull"`
	PositiveReportsThreshold    int    `bun:"positive_reports_threshold,notnull"`
	NeedRecalculation           bool   `bun:"need_recalculation"`
}
