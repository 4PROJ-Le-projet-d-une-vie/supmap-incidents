package services

import (
	"context"
	"net/http"
	"sort"
	"supmap-users/internal/api/validations"
	"supmap-users/internal/models"
	"supmap-users/internal/models/dto"
	rediss "supmap-users/internal/services/redis"
	"time"
)

func (s *Service) CreateInteraction(ctx context.Context, user *dto.PartialUserDTO, body *validations.CreateInteractionValidator) (*models.Interaction, error) {

	tx, err := s.incidents.AskForTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			s.log.Info("Rollback de la transaction")
			_ = tx.Rollback()
		} else {
			s.log.Info("Commit de la transaction")
			_ = tx.Commit()
		}
	}()

	// Check si l'incident existe
	incident, err := s.incidents.FindIncidentByIdTx(ctx, tx, body.IncidentID)
	if err != nil {
		return nil, err
	}

	if incident == nil {
		return nil, &ErrorWithCode{
			Message: "Incident does not exists",
			Code:    http.StatusNotFound,
		}
	}

	if incident.DeletedAt != nil {
		return nil, &ErrorWithCode{
			Message: "This incident is locked",
			Code:    http.StatusLocked,
		}
	}

	if incident.UserID == user.ID {
		return nil, &ErrorWithCode{
			Message: "You can't interact with your own incident",
			Code:    http.StatusForbidden,
		}
	}

	//	Check si l'utilisateur à déjà intéragis avec l'incident
	//	Si intéragis, il y a moins d'une heure → Too many requests
	//	Sinon on accepte l'intéraction
	for _, interaction := range incident.Interactions {
		if time.Since(interaction.CreatedAt) < 60*time.Minute {
			return nil, &ErrorWithCode{
				Message: "Too many interactions with this incident",
				Code:    http.StatusTooManyRequests,
			}
		}
	}

	// Créer l'intéraction
	toInsert := &models.Interaction{
		IncidentID:     incident.ID,
		UserID:         user.ID,
		IsStillPresent: *body.IsStillPresent,
		CreatedAt:      time.Now(),
	}

	if err := s.interactions.InsertTx(ctx, tx, toInsert); err != nil {
		return nil, err
	}

	err = s.incidents.UpdateIncidentTx(ctx, tx, incident)
	if err != nil {
		return nil, err
	}

	inserted, err := s.interactions.FindInteractionByIdTx(ctx, tx, toInsert.ID)
	if err != nil {
		return nil, err
	}

	// Trie les éléments selon leur date (récente → ancienne)
	interactions := inserted.Incident.Interactions
	sort.SliceStable(interactions, func(i, j int) bool {
		return interactions[i].CreatedAt.After(interactions[j].CreatedAt)
	})

	// Compte le nombre d'intéractions "négatives"
	negativeCount := 0
	for _, interaction := range interactions {
		if interaction.IsStillPresent {
			break // Coupe l'exécution dès qu'une intéraction positive est trouvée
		}
		negativeCount++
	}

	// Cas où il y a suffisamment d'intéractions négatives
	// pour considérer l'incident comme terminé
	if negativeCount >= inserted.Incident.Type.NegativeReportsThreshold {
		now := time.Now()
		inserted.Incident.DeletedAt = &now

		if err := s.incidents.UpdateIncidentTx(ctx, tx, inserted.Incident); err != nil {
			return nil, err
		}

		err := s.redis.PublishMessage(s.config.IncidentChannel, rediss.IncidentMessage{
			Data:   *dto.IncidentToRedis(inserted.Incident),
			Action: rediss.Deleted,
		})
		if err != nil {
			s.log.Error("failed to send message to redis", "channel", "test", "message", "message")
		}

		return nil, &ErrorWithCode{
			Code: http.StatusNoContent,
		}
	}

	return inserted, err
}
