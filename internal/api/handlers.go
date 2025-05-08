package api

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/matheodrd/httphelper/handler"
	"net/http"
	"reflect"
	"strconv"
	"strings"
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

// GetAllInRadius godoc
// @Summary Récupérer les incidents dans un rayon donné
// @Description Récupère tous les incidents non supprimés situés dans un rayon donné à partir des coordonnées passées en paramètre.
// @Description Cette requête est très couteuse car elle effectue de nombreux appels à la base de données pour charger l'entièreté des données. Elle est à utiliser avec précautions !
// @Tags incidents
// @Accept json
// @Produce json
// @Param lat query number true "Latitude du centre de la recherche"
// @Param lon query number true "Longitude du centre de la recherche"
// @Param radius query integer true "Rayon de recherche en mètres"
// @Param include query string false "Inclure des données additionnelles : 'interactions' pour le détail complet ou 'summary' pour un résumé" Enums(interactions,summary)
// @Success 200 {array} dto.IncidentWithDistanceDTO "Liste des incidents dans le rayon, avec distance calculée"
// @Failure 400 {object} ErrorResponse "Paramètres invalides ou manquants"
// @Failure 500 {object} InternalErrorResponse "Erreur interne du serveur"
// @Router /incidents [get]
func (s *Server) GetAllInRadius() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		include := decodeIncludeParam(r)

		latitude, err := decodeParamAs[float64](r, "lat")
		if err != nil {
			return encode(&ErrorResponse{Error: err.Error()}, http.StatusBadRequest, w)
		}

		longitude, err := decodeParamAs[float64](r, "lon")
		if err != nil {
			return encode(&ErrorResponse{Error: err.Error()}, http.StatusBadRequest, w)
		}

		radius, err := decodeParamAs[int64](r, "radius")
		if err != nil {
			return encode(&ErrorResponse{Error: err.Error()}, http.StatusBadRequest, w)
		}

		incidentType, _ := decodeParamAs[*int64](r, "type_id")

		incidents, err := s.service.FindIncidentsInRadius(r.Context(), incidentType, latitude, longitude, radius)
		if err != nil {
			if ewc := services.DecodeErrorWithCode(err); ewc != nil {
				return encode(ewc, ewc.Code, w)
			}
		}

		var incidentsDTOs = make([]dto.IncidentWithDistanceDTO, len(incidents))
		for i, incident := range incidents {
			incidentsDTOs[i] = *dto.IncidentWithDistanceToDTO(&incident, include)
		}

		return encode(incidentsDTOs, http.StatusOK, w)
	})
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
// @Router /incidents [post]
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

// GetUserHistory godoc
// @Summary Récupérer l’historique des incidents de l’utilisateur
// @Description Récupère tous les incidents créés par l’utilisateur authentifié.
// @Description L'historique ne comprend que les incidents qui ont été supprimés.
// @Tags incidents
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param include query string false "Inclure des données additionnelles : 'interactions' pour les détails complets ou 'summary' pour les statistiques" Enums(interactions,summary)
// @Success 200 {array} dto.IncidentDTO "Liste des anciens incidents (supprimés) de l'utilisateur"
// @Failure 401 {object} ErrorResponse "Utilisateur non authentifié"
// @Failure 500 {object} InternalErrorResponse "Erreur interne du serveur"
// @Router /incidents/me/history [get]
func (s *Server) GetUserHistory() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		user, ok := r.Context().Value("user").(*dto.PartialUserDTO)
		if !ok {
			return encodeNil(http.StatusUnauthorized, w)
		}

		incidents, err := s.service.GetUserHistory(r.Context(), user)
		if err != nil {
			return err
		}

		interactionState := decodeIncludeParam(r)
		var incidentsDTOs = make([]dto.IncidentDTO, len(incidents))
		for i, incident := range incidents {
			incidentsDTOs[i] = *dto.IncidentToDTO(&incident, interactionState)
		}

		return encode(incidentsDTOs, http.StatusOK, w)
	})
}

// GetIncidentsTypes godoc
// @Summary Récupère tous les types d'incidents
// @Description Permet de récupérer la liste de tous les types d'incidents.
// @Tags incidents types
// @Accept json
// @Produce json
// @Success 200 {array} dto.TypeDTO "Liste des types d'incidents"
// @Failure 500 {object} services.ErrorWithCode "Erreur interne du serveur"
// @Router /incidents/types [get]
func (s *Server) GetIncidentsTypes() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		types, err := s.service.GetAllIncidentTypes(r.Context())
		if err != nil {
			return err
		}

		typesDTOs := make([]dto.TypeDTO, len(types))
		for i, t := range types {
			typesDTOs[i] = *dto.TypeToDTO(&t)
		}

		return encode(typesDTOs, http.StatusOK, w)
	})
}

// GetIncidentTypeById godoc
// @Summary Récupère un type d'incident par son ID
// @Description Permet de récupérer les détails d'un type d'incident spécifié par son ID.
// @Tags incidents types
// @Accept json
// @Produce json
// @Param id path int64 true "ID du type d'incident"
// @Success 200 {object} dto.TypeDTO "Type d'incident trouvé avec succès"
// @Failure 400 {object} services.ErrorWithCode "ID du type d'incident invalide"
// @Failure 404 {object} services.ErrorWithCode "Type d'incident non trouvé"
// @Failure 500 {object} services.ErrorWithCode "Erreur interne du serveur"
// @Router /incidents/types/{id} [get]
func (s *Server) GetIncidentTypeById() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		id, err := decodeParamAsInt64("id", r)
		if err != nil {
			return err
		}

		t, err := s.service.FindTypeById(r.Context(), id)
		if err != nil {
			if ewc := services.DecodeErrorWithCode(err); ewc != nil {
				return encode(ewc, ewc.Code, w)
			}
			return err
		}

		typeDTO := dto.TypeToDTO(t)
		return encode(typeDTO, http.StatusOK, w)
	})
}

// UserInteractWithIncident godoc
// @Summary Crée une interaction avec un incident
// @Description Permet à un utilisateur d'interagir avec un incident en fonction de son ID et de son statut d'interaction.
// @Tags interactions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body validations.CreateInteractionValidator true "Informations de l'interaction"
// @Success 200 {object} dto.InteractionDTO "Interaction créée avec succès"
// @Failure 400 {object} services.ErrorWithCode "Paramètres invalides"
// @Failure 404 {object} services.ErrorWithCode "Incident non trouvé"
// @Failure 409 {object} services.ErrorWithCode "Incident verrouillé"
// @Failure 429 {object} services.ErrorWithCode "Trop d'interactions avec cet incident"
// @Failure 500 {object} services.ErrorWithCode "Erreur interne du serveur"
// @Router /incidents/interactions [post]
func (s *Server) UserInteractWithIncident() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		user, ok := r.Context().Value("user").(*dto.PartialUserDTO)
		if !ok {
			return encodeNil(http.StatusUnauthorized, w)
		}

		body, err := handler.Decode[validations.CreateInteractionValidator](r)
		if err != nil {
			return buildValidationErrors(err, w)
		}

		interaction, err := s.service.CreateInteraction(r.Context(), user, &body)
		if err != nil {
			if ewc := services.DecodeErrorWithCode(err); ewc != nil {
				return encode(ewc, ewc.Code, w)
			}
			return err
		}

		includeParam := decodeIncludeParam(r)
		interactionDTO := dto.InteractionToDTO(*interaction, includeParam)
		return encode(interactionDTO, http.StatusOK, w)
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

func decodeParamAs[T any](r *http.Request, param string) (T, error) {
	var zero T

	query := r.URL.Query()
	var value = query.Get(param)
	if !query.Has(param) || value == "" {
		return zero, fmt.Errorf("parameter %s not provided", param)
	}

	var result any
	var err error

	switch any(zero).(type) {
	case float64:
		result, err = strconv.ParseFloat(strings.TrimSpace(value), 64)
	case *int64:
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil {
			result = &parsed
		}
	case int64:
		result, err = strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	default:
		err = fmt.Errorf("unsupported type %s", reflect.TypeOf(zero).String())
	}

	if err != nil {
		return zero, fmt.Errorf("invalid value for %s: %w", param, err)
	}

	return result.(T), nil
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

	if status == http.StatusNoContent {
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
