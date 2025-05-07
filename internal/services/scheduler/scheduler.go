package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"supmap-users/internal/config"
	"supmap-users/internal/repository"
	"supmap-users/internal/services/redis"
	"time"
)

type Scheduler struct {
	log         *slog.Logger
	config      *config.Config
	ticker      *time.Ticker
	stop        chan bool
	incidents   *repository.Incidents
	interaction *repository.Interactions
	redis       *redis.Redis
}

func NewScheduler(delay time.Duration, config *config.Config, incidents *repository.Incidents, interactions *repository.Interactions, redis *redis.Redis, log *slog.Logger) *Scheduler {
	return &Scheduler{
		log:         log,
		config:      config,
		ticker:      time.NewTicker(delay),
		stop:        make(chan bool),
		incidents:   incidents,
		interaction: interactions,
		redis:       redis,
	}
}

func (s *Scheduler) Run() {
	go func() {
		for {
			select {
			case <-s.ticker.C: // Chaque fois que le ticker émet un signal
				s.log.Info("starting auto-moderation task")

				ctx, cancel := context.WithCancel(context.Background())

				tx, err := s.incidents.AskForTx(ctx)
				if err != nil {
					s.log.Error("failed to get transaction locker", "error", err)
					cancel()
					return
				}
				s.CheckLifetimeWithoutConfirmation(ctx, tx)
				s.CheckGlobalLifeTime(ctx, tx)

				_ = tx.Commit()
				cancel()

				s.log.Info("auto-moderation task terminated")
			case <-s.stop: // Permet d'arrêter proprement le scheduler
				fmt.Println("Scheduler arrêté.")
				s.ticker.Stop()
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.stop <- true // Envoyer un signal pour arrêter
}
