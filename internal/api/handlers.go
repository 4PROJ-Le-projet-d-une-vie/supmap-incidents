package api

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/matheodrd/httphelper/handler"
	"net/http"
	"strconv"
	"supmap-users/internal/api/validations"
	"supmap-users/internal/models"
	"supmap-users/internal/models/dto"
	"supmap-users/internal/services"
)

type InternalErrorResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateIncident godoc
// @Summary Créer un incident
// @Description Crée un nouvel incident si aucun n'existe dans la zone (<100m). Sinon, ajoute une interaction à l'incident existant.
// @Tags incidents
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param include query string false "Inclure des données additionnelles : 'interactions' pour les interactions complètes ou 'summary' pour un résumé" Enums(interactions,summary)
// @Param incident body validations.CreateIncidentValidator true "Données nécessaires pour créer un incident"
// @Success 200 {object} dto.IncidentDTO "Incident créé avec succès"
// @Success 202 {object} dto.IncidentDTO "Interaction ajoutée à un incident existant"
// @Failure 400 {object} services.ErrorWithCode "Type d'incident inexistant ou données invalides"
// @Failure 401 {object} nil "Utilisateur non authentifié"
// @Failure 429 {object} services.ErrorWithCode "Signalement trop fréquent ou interaction déjà existante"
// @Failure 500 {object} InternalErrorResponse "Erreur interne du serveur"
// @Router /incident [post]
func (s *Server) CreateIncident() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		user, ok := r.Context().Value("user").(*dto.PartialUserDTO)
		if !ok {
			return encodeNil(http.StatusUnauthorized, w)
		}

		body, err := handler.Decode[validations.CreateIncidentValidator](r)
		if err != nil {
			return buildValidationErrors(err, w)
		}

		interactionState := decodeIncludeParam(r)
		incident, err := s.service.CreateIncident(r.Context(), user, &body)
		if err != nil {
			if ewc := services.DecodeErrorWithCode(err); ewc != nil {
				return encode(ewc, ewc.Code, w)
			}

			if ewb := services.DecodeErrorWithBody[models.Incident](err); ewb != nil {
				if incident, ok := ewb.GetBody().(models.Incident); ok {
					fmt.Println(incident)
					incidentDTO := dto.IncidentToDTO(&incident, interactionState)
					return encode(incidentDTO, ewb.Code, w)
				}
			}

			return err
		}

		incidentDTO := *dto.IncidentToDTO(incident, interactionState)

		return encode(incidentDTO, http.StatusOK, w)
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
