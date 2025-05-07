package models

import (
	"github.com/uptrace/bun"
	"time"
)

type Incident struct {
	bun.BaseModel `bun:"table:incidents,alias:i"`

	ID        int64      `json:"id" bun:"id,pk,autoincrement"`
	TypeID    int64      `json:"-" bun:"type_id,notnull"`
	UserID    int64      `json:"user_id" bun:"user_id,notnull"`
	Latitude  float64    `json:"lat" bun:"latitude,notnull"`
	Longitude float64    `json:"lon" bun:"longitude,notnull"`
	CreatedAt time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	// Relations
	Type         *Type         `json:"type" bun:"rel:belongs-to,join:type_id=id"`
	Interactions []Interaction `json:"interactions" bun:"rel:has-many,join:id=incident_id"`
}

type IncidentWithDistance struct {
	Incident
	Distance float64 `json:"distance" bun:"distance"`
}
