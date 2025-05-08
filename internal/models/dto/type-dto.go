package dto

import "supmap-users/internal/models"

type TypeDTO struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	NeedRecalculation bool   `json:"need_recalculation"`
}

func TypeToDTO(iType *models.Type) *TypeDTO {
	return &TypeDTO{
		ID:                iType.ID,
		Name:              iType.Name,
		Description:       iType.Description,
		NeedRecalculation: iType.NeedRecalculation,
	}
}
