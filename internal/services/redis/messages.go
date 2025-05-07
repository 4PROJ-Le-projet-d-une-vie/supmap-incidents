package redis

import (
	"supmap-users/internal/models/dto"
)

type Action string

const (
	Create    Action = "create"
	Certified Action = "certified"
	Deleted   Action = "deleted"
)

type IncidentMessage struct {
	Data   dto.IncidentRedis `json:"data"`
	Action Action            `json:"action"`
}
