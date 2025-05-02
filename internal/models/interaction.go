package models

import (
	"github.com/uptrace/bun"
	"time"
)

type Interaction struct {
	bun.BaseModel `bun:"table:incident_interactions,alias:ii"`

	ID             int64     `json:"id" bun:"id,pk,autoincrement"`
	IncidentID     int64     `json:"-" bun:"incident_id,notnull"`
	UserID         int64     `json:"user_id" bun:"user_id,notnull"`
	IsStillPresent bool      `json:"is_still_present" bun:"is_still_present,notnull"`
	CreatedAt      time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`

	// Relations
	Incident *Incident `json:"incident,omitempty" bun:"rel:belongs-to,join:incident_id=id"`
}