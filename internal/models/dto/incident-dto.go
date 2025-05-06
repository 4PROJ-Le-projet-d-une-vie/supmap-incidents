package dto

import (
	"supmap-users/internal/models"
	"time"
)

type IncidentDTO struct {
	ID        int64           `json:"id"`
	User      *PartialUserDTO `json:"user"`
	Type      *TypeDTO        `json:"type"`
	Latitude  float64         `json:"lat"`
	Longitude float64         `json:"lgn"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at,omitempty"`

	Interactions        []InteractionDTO        `json:"interactions,omitempty"`
	InteractionsSummary *InteractionsSummaryDTO `json:"interactions_summary,omitempty"`
}

type IncidentWithDistanceDTO struct {
	IncidentDTO
	Distance float64 `json:"distance"`
}

func IncidentToDTO(incident *models.Incident, interactionsState InteractionsResultState) *IncidentDTO {
	partialUserDTO, _ := UserIdToDTO(incident.UserID)
	incidentDTO := IncidentDTO{
		ID:        incident.ID,
		User:      partialUserDTO,
		Type:      TypeToDTO(incident.Type),
		Latitude:  incident.Latitude,
		Longitude: incident.Longitude,
		CreatedAt: incident.CreatedAt,
		UpdatedAt: incident.UpdatedAt,
		DeletedAt: incident.DeletedAt,
	}

	switch interactionsState {
	case IncludeInteractions:
		incidentDTO.Interactions = buildInteractionsDTO(incident.Interactions)
	case IncludeAsSummary:
		incidentDTO.InteractionsSummary = InteractionsToSummaryDTO(incident.Interactions)
	case Ignore:
		incidentDTO.Interactions = nil
		incidentDTO.InteractionsSummary = nil
	}

	return &incidentDTO
}

func buildInteractionsDTO(interactions []models.Interaction) []InteractionDTO {
	var interactionsDTO = make([]InteractionDTO, len(interactions))
	for i, dto := range interactions {
		interactionsDTO[i] = *InteractionToDTO(dto, Ignore)
	}
	return interactionsDTO
}

func IncidentWithDistanceToDTO(incident *models.IncidentWithDistance, interactionsState InteractionsResultState) *IncidentWithDistanceDTO {
	incidentDTO := *IncidentToDTO(&models.Incident{
		ID:           incident.ID,
		UserID:       incident.UserID,
		Type:         incident.Type,
		Latitude:     incident.Latitude,
		Longitude:    incident.Longitude,
		CreatedAt:    incident.CreatedAt,
		UpdatedAt:    incident.UpdatedAt,
		DeletedAt:    incident.DeletedAt,
		Interactions: incident.Interactions,
	}, interactionsState)

	return &IncidentWithDistanceDTO{
		IncidentDTO: incidentDTO,
		Distance:    incident.Distance,
	}
}
