package scheduler

import (
	"context"
	"fmt"
	"github.com/uptrace/bun"
	"supmap-users/internal/models/dto"
	rediss "supmap-users/internal/services/redis"
	"time"
)

// CheckLifetimeWithoutConfirmation godoc
// Vérifie les incidents en cours qui n'ont pas eu d'intéraction
// selon le temps définit par type d'incident
func (s *Scheduler) CheckLifetimeWithoutConfirmation(ctx context.Context, exec *bun.Tx) {
	incidents, err := s.incidents.GetAllActive(ctx, exec)
	if err != nil {
		s.log.Error("failed to retrieve all active incidents", "error", err)
		return
	}

	for _, incident := range incidents {
		noInteractionThreshold := time.Duration(incident.Type.LifetimeWithoutConfirmation)
		if time.Since(incident.UpdatedAt) > noInteractionThreshold*time.Second {
			now := time.Now()
			incident.DeletedAt = &now

			if err = s.incidents.UpdateIncidentTx(ctx, exec, &incident); err != nil {
				s.log.Error("failed to updated incident", "error", err)
			}

			err = s.redis.PublishMessage(s.config.IncidentChannel, &rediss.IncidentMessage{
				Data:   *dto.IncidentToRedis(&incident),
				Action: rediss.Deleted,
			})
			if err != nil {
				s.log.Error("failed to send redis message", "error", err)
				return
			}

			s.log.Info(fmt.Sprintf("incident %d deleted as it has had no reaction since %f seconds", incident.ID, noInteractionThreshold.Seconds()))
		}
	}
}

// CheckGlobalLifeTime godoc
// Vérifie les incidents en cours qui ont dépassé leur durée de vie maximale
func (s *Scheduler) CheckGlobalLifeTime(ctx context.Context, exec *bun.Tx) {
	incidents, err := s.incidents.GetAllActive(ctx, exec)
	if err != nil {
		s.log.Error("failed to retrieve all active incidents", "error", err)
		return
	}

	for _, incident := range incidents {
		incidentTTL := time.Duration(incident.Type.GlobalLifetime)
		if time.Since(incident.CreatedAt) > incidentTTL*time.Second {
			now := time.Now()
			incident.DeletedAt = &now

			if err = s.incidents.UpdateIncidentTx(ctx, exec, &incident); err != nil {
				s.log.Error("failed to updated incident", "error", err)
				return
			}

			err = s.redis.PublishMessage(s.config.IncidentChannel, &rediss.IncidentMessage{
				Data:   *dto.IncidentToRedis(&incident),
				Action: rediss.Deleted,
			})
			if err != nil {
				s.log.Error("failed to send redis message", "error", err)
				return
			}

			s.log.Info(fmt.Sprintf("incident %d deleted because it has reached its maximum ttl %f", incident.ID, incidentTTL.Seconds()))
		}
	}
}
