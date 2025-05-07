package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
	"log"
	"log/slog"
	"os"
	"supmap-users/internal/api"
	"supmap-users/internal/config"
	"supmap-users/internal/repository"
	"supmap-users/internal/services"
	rediss "supmap-users/internal/services/redis"
	"supmap-users/internal/services/scheduler"
	"supmap-users/migrations"
	"time"
)

// @title SupMap Incidents API
// @version 1.0
// @description Cette API permet de gérer les incidents de SupMap.

// @contact.name Ewen
// @contact.email ewen.bosquet@supinfo.com

// @host localhost:8081
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	conf, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	// Configure logger
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	// Run database migrations
	if err := migrations.Migrate("pgx", conf.DbUrl, logger); err != nil {
		logger.Error("migration failed", "err", err)
	}

	// Open SQL connection
	conn, err := sql.Open("pgx", conf.DbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Create Bun client
	bunDB := bun.NewDB(conn, pgdialect.New())
	if conf.ENV == "development" {
		bunDB.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	if err := bunDB.Ping(); err != nil {
		log.Fatal(fmt.Errorf("failed to connect to database: %w", err))
	}

	// Create users repository
	incidents := repository.NewIncidents(bunDB, logger)
	interactions := repository.NewInteractions(bunDB, logger)

	// Redis service
	rdb := redis.NewClient(&redis.Options{
		Addr: conf.RedisHost + ":" + conf.RedisPort,
		DB:   0,
	})
	redisService := rediss.NewRedis(rdb, logger)
	redisService.Run(context.Background())

	// Create users service
	service := services.NewService(logger, conf, incidents, interactions, redisService)

	// Taches actives pour l'auto modération des incidents
	tasks := scheduler.NewScheduler(1*time.Second, incidents, interactions, redisService)
	//tasks.Run()
	defer tasks.Stop()

	// Create the HTTP server
	server := api.NewServer(conf, logger, service)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
