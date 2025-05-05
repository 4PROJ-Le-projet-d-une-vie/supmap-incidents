package services

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"supmap-users/internal/api/validations"
	"supmap-users/internal/config"
	"supmap-users/internal/models"
	"supmap-users/internal/models/dto"
	"supmap-users/internal/repository"
	"time"
)

type Service struct {
	log       *slog.Logger
	config    *config.Config
	incidents *repository.Incidents
}

func NewService(log *slog.Logger, config *config.Config, incidents *repository.Incidents) *Service {
	return &Service{
		log:       log,
		config:    config,
		incidents: incidents,
	}
}

type ErrorWithCode struct {
	Message string `json:"error"`
	Code    int    `json:"-"`
}

func (e ErrorWithCode) Error() string {
	return e.Message
}

func DecodeErrorWithCode(err error) *ErrorWithCode {
	var ewc *ErrorWithCode
	if errors.As(err, &ewc) {
		return ewc
	}
	return nil
}

type ErrorWithBody[T any] struct {
	ErrorWithCode
	Body T
}

func (e ErrorWithBody[T]) Error() string {
	return e.Message
}

func (e ErrorWithBody[T]) GetBody() any {
	return e.Body
}

func DecodeErrorWithBody[T any](err error) *ErrorWithBody[T] {
	var ewb *ErrorWithBody[T]
	if errors.As(err, &ewb) {
		return ewb
	}
	return nil
}

func (s *Service) CreateIncident(ctx context.Context, user *dto.PartialUserDTO, body *validations.CreateIncidentValidator) (*models.Incident, error) {

	// Check si le type existe
	incidentType, err := s.incidents.GetTypeById(ctx, body.TypeId)
	if err != nil {
		return nil, err
	}

	if incidentType == nil {
		return nil, &ErrorWithCode{
			Message: "Incident type does not exists",
			Code:    400,
		}
	}

	// Check le dernier report de l'utilisateur (un signalement par minute)
	last, err := s.incidents.GetLastUserIncident(ctx, user)
	if err != nil {
		return nil, err
	}

	if last != nil {
		if time.Since(last.CreatedAt) < 60*time.Second {
			return nil, &ErrorWithCode{
				Message: "Too many incidents reported",
				Code:    http.StatusTooManyRequests,
			}
		}
	}

	// Si un ou plusieurs incidents existent déjà dans un rayon de 100m :
	// Le signalement n’en crée pas un nouveau, mais devient une interaction attachée à un incident existant.
	// Si plusieurs incidents existent :
	// On choisit celui avec le plus d’interactions.
	// S’il y a égalité : on prend le plus proche.
	// En cas d’égalité parfaite : on choisit arbitrairement (ex. premier de la liste).
	incidents, err := s.incidents.FindIncidentsInZone(ctx, body.Latitude, body.Longitude, 100, &body.TypeId)
	if len(incidents) > 0 {
		incident, err := s.incidents.FindIncidentById(ctx, incidents[0].ID)
		if err != nil {
			return nil, err
		}

		// TODO ajouter réaction
		// Check si l'utilisateur à déjà réagit
		// S'il a déjà réagi, alors 429
		// Sinon 202 avec ajout de l'intéraction

		return nil, &ErrorWithBody[models.Incident]{
			ErrorWithCode: ErrorWithCode{
				Message: "Interaction added to incident with id",
				Code:    http.StatusAccepted,
			},
			Body: *incident,
		}
	}

	// Insérer l'incident
	incident := &models.Incident{
		TypeID:    incidentType.ID,
		UserID:    user.ID,
		Latitude:  *body.Latitude,
		Longitude: *body.Longitude,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err = s.incidents.CreateIncident(ctx, incident); err != nil {
		return nil, err
	}

	inserted, err := s.incidents.FindIncidentById(ctx, incident.ID)
	if err != nil {
		return nil, err
	}

	// TODO Envoie d'un event dans le pub/sub redis (go routine)

	return inserted, nil
}
