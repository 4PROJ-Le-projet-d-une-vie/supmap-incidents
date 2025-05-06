package dto

import (
	"supmap-users/internal/models"
	"time"
)

type InteractionDTO struct {
	ID             int64           `json:"id"`
	User           *PartialUserDTO `json:"user"`
	IsStillPresent bool            `json:"is-still-present"`
	CreatedAt      time.Time       `json:"created_at"`

	Incident *IncidentDTO `json:"incident,omitempty"`
}

func InteractionToDTO(interaction models.Interaction, includeIncidents InteractionsResultState) *InteractionDTO {
	partialUserDTO, _ := UserIdToDTO(interaction.UserID)
	interactionDTO := InteractionDTO{
		ID:             interaction.ID,
		User:           partialUserDTO,
		IsStillPresent: interaction.IsStillPresent,
		CreatedAt:      interaction.CreatedAt,
	}

	if includeIncidents != Ignore {
		interactionDTO.Incident = IncidentToDTO(interaction.Incident, includeIncidents)
	}

	return &interactionDTO
}

type InteractionsSummaryDTO struct {
	IsStillPresentSum int `json:"is-still-present"`
	NoStillPresentSum int `json:"no-still-present"`
	Total             int `json:"total"`
}

func InteractionsToSummaryDTO(interactions []models.Interaction) *InteractionsSummaryDTO {
	total := len(interactions)

	var isStillPresentSum = 0
	for _, interaction := range interactions {
		if interaction.IsStillPresent {
			isStillPresentSum++
		}
	}

	noStillPresentSum := total - isStillPresentSum

	return &InteractionsSummaryDTO{
		IsStillPresentSum: isStillPresentSum,
		NoStillPresentSum: noStillPresentSum,
		Total:             total,
	}
}
