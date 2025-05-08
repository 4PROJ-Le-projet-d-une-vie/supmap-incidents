package services

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"supmap-users/internal/api/validations"
	"supmap-users/internal/config"
	"supmap-users/internal/models"
	"supmap-users/internal/models/dto"
	"supmap-users/internal/repository"
	"supmap-users/internal/services/redis"
	"time"
)

type Service struct {
	log          *slog.Logger
	config       *config.Config
	incidents    *repository.Incidents
	interactions *repository.Interactions
	redis        *redis.Redis
}

func NewService(log *slog.Logger, config *config.Config, incidents *repository.Incidents, interactions *repository.Interactions, redis *redis.Redis) *Service {
	return &Service{
		log:          log,
		config:       config,
		incidents:    incidents,
		interactions: interactions,
		redis:        redis,
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

func (s *Service) GetAllIncidentTypes(ctx context.Context) ([]models.Type, error) {
	return s.incidents.FindAllIncidentTypes(ctx)
}

func (s *Service) FindTypeById(ctx context.Context, id int64) (*models.Type, error) {
	t, err := s.incidents.FindIncidentTypeById(ctx, &id)
	if err != nil {
		return nil, err
	}

	if t == nil {
		return nil, &ErrorWithCode{
			Message: "This incident type does not exists",
			Code:    http.StatusNotFound,
		}
	}

	return t, err
}

func (s *Service) CreateIncident(ctx context.Context, user *dto.PartialUserDTO, body *validations.CreateIncidentValidator) (*models.Incident, error) {

	// Check si le type existe
	incidentType, err := s.incidents.FindIncidentTypeById(ctx, &body.TypeId)
	if err != nil {
		return nil, err
	}

	if incidentType == nil {
		return nil, &ErrorWithCode{
			Message: "Incident type does not exists",
			Code:    http.StatusBadRequest,
		}
	}

	// Check le dernier report de l'utilisateur (un signalement par minute)
	last, err := s.incidents.GetLastUserIncident(ctx, user)
	if err != nil {
		return nil, err
	}

	if last != nil && s.config.ENV == "production" {
		if time.Since(last.CreatedAt) < time.Minute {
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
		// Choisir l'évènement à intéragir
		sort.SliceStable(incidents, func(i, j int) bool {
			// Le plus d'intéractions
			if len(incidents[i].Interactions) != len(incidents[j].Interactions) {
				return len(incidents[i].Interactions) > len(incidents[j].Interactions)
			}
			// Sinon, le plus proche
			return incidents[i].Distance < incidents[j].Distance
		})
		chosen := incidents[0]

		incident, err := s.incidents.FindIncidentById(ctx, chosen.ID)
		if err != nil {
			return nil, err
		}

		newInteraction := &validations.CreateInteractionValidator{
			IncidentID:     chosen.ID,
			IsStillPresent: toPtr(true),
		}

		if _, err := s.CreateInteraction(ctx, user, newInteraction); err != nil {
			return nil, err
		}

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

	err = s.redis.PublishMessage(s.config.IncidentChannel, &redis.IncidentMessage{
		Data:   *dto.IncidentToRedis(inserted),
		Action: redis.Create,
	})
	if err != nil {
		return nil, err
	}

	return inserted, nil
}

func (s *Service) FindIncidentsInRadius(ctx context.Context, typeId *int64, lat, lon float64, radius int64) ([]models.IncidentWithDistance, error) {
	incidentType, err := s.incidents.FindIncidentTypeById(ctx, typeId)
	if err != nil {
		return nil, err
	}

	if incidentType == nil {
		return nil, &ErrorWithCode{
			Message: "Incident type does not exists",
			Code:    http.StatusNotFound,
		}
	}

	incidents, err := s.incidents.FindIncidentsInZone(ctx, &lat, &lon, radius, typeId)
	if err != nil {
		return nil, err
	}

	for i, incident := range incidents {
		completed, err := s.incidents.FindIncidentById(ctx, incident.ID)
		if err != nil {
			return nil, err
		}

		incidents[i].Type = completed.Type
		incidents[i].Interactions = completed.Interactions
	}

	return incidents, err
}

func (s *Service) GetUserHistory(ctx context.Context, user *dto.PartialUserDTO) ([]models.Incident, error) {
	return s.incidents.FindUserHistory(ctx, user)
}

func toPtr[T any](t T) *T {
	return &t
}
