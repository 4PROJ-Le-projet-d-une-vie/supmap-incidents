package dto

import "supmap-users/internal/models"

type TypeDTO struct {
	ID                          int64  `json:"id"`
	Name                        string `json:"name"`
	LifetimeWithoutConfirmation int    `json:"lifetime-without-confirmation"`
	NegativeReportsThreshold    int    `json:"negative-reports-threshold"`
	GlobalLifeTime              int    `json:"global-life-time"`
}

func TypeToDTO(iType *models.Type) *TypeDTO {
	return &TypeDTO{
		ID:                          iType.ID,
		Name:                        iType.Name,
		LifetimeWithoutConfirmation: iType.LifetimeWithoutConfirmation,
		NegativeReportsThreshold:    iType.NegativeReportsThreshold,
		GlobalLifeTime:              iType.GlobalLifetime,
	}
}
