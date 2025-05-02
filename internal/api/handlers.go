package api

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/matheodrd/httphelper/handler"
	"net/http"
	"strconv"
	"supmap-users/internal/api/validations"
	"supmap-users/internal/models/dto"
)

type InternalErrorResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// @Summary Récupérer tous les incidents
// @Description Récupère tous les incidents non supprimés avec inclusion optionnelle des interactions
// @Tags incidents
// @Accept json
// @Produce json
// @Param include query string false "Inclure des données additionnelles : 'interactions' pour le détail complet des interactions ou 'summary' pour les statistiques" Enums(interactions,summary)
// @Success 200 {array} dto.IncidentDTO "Liste des incidents récupérée avec succès"
// @Success 200 {object} []dto.IncidentDTO{interactions=[]dto.InteractionDTO} "Exemple avec les interactions incluses"
// @Success 200 {object} []dto.IncidentDTO{interactions_summary=dto.InteractionsSummaryDTO} "Exemple avec le résumé inclus"
// @Failure 500 {object} InternalErrorResponse "Erreur interne du serveur"
// @Router /incidents [get]
func (s *Server) getIncidents() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		incidents, err := s.service.GetIncidents(r.Context())
		if err != nil {
			return encodeNil(http.StatusInternalServerError, w)
		}

		interactionState := decodeIncludeParam(r)
		var incidentsDTOs = make([]dto.IncidentDTO, len(incidents))
		for i, incident := range incidents {
			incidentsDTOs[i] = *dto.IncidentToDTO(&incident, interactionState)
		}

		return encode(incidentsDTOs, http.StatusOK, w)
	})
}

func decodeParamAsInt64(param string, r *http.Request) (int64, error) {
	value := r.PathValue(param)
	converted, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return converted, nil
}

func decodeIncludeParam(r *http.Request) dto.InteractionsResultState {
	include := r.URL.Query().Get("include")

	var interactionState dto.InteractionsResultState
	switch include {
	case "interactions":
		interactionState = dto.IncludeInteractions
	case "summary":
		interactionState = dto.IncludeAsSummary
	default:
		interactionState = dto.Ignore
	}

	return interactionState
}

func encodeNil(status int, w http.ResponseWriter) error {
	return encode(nil, status, w)
}

func encode(body any, status int, w http.ResponseWriter) error {
	if body == nil {
		w.WriteHeader(status)
		return nil
	}

	if err := handler.Encode(body, status, w); err != nil {
		return err
	}
	return nil
}

func buildValidationErrors(err error, w http.ResponseWriter) error {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return err
	}

	errs := make(map[string]string)

	for _, fieldErr := range ve {
		errs[fieldErr.Field()] = fmt.Sprintf("failed on '%s'", fieldErr.Tag())
	}

	validationErrorResponse := validations.ValidationError{Message: "Validation Error", Details: errs}
	return encode(validationErrorResponse, http.StatusBadRequest, w)
}
