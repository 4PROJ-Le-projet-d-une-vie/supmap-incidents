package api

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	_ "supmap-users/docs"
	"supmap-users/internal/config"
	"supmap-users/internal/services"
)

type Server struct {
	Config  *config.Config
	log     *slog.Logger
	service *services.Service
}

func NewServer(config *config.Config, log *slog.Logger, service *services.Service) *Server {
	return &Server{
		Config:  config,
		log:     log,
		service: service,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/docs/", httpSwagger.WrapHandler)

	mux.Handle("GET /incidents", s.GetAllInRadius())
	mux.Handle("GET /incidents/me/history", s.AuthMiddleware()(s.GetUserHistory()))
	mux.Handle("GET /incidents/types", s.GetIncidentsTypes())
	mux.Handle("GET /incidents/types/{id}", s.GetIncidentTypeById())
	mux.Handle("POST /incidents", s.AuthMiddleware()(s.CreateIncident()))

	mux.Handle("POST /incidents/interactions", s.AuthMiddleware()(s.UserInteractWithIncident()))

	// These routes are not exposed outside the LAN
	//  server network and doesn't require securities
	mux.Handle("GET /internal/incidents", s.GetAllInRadius())

	server := &http.Server{
		Addr:    ":" + s.Config.PORT,
		Handler: WithCORS(mux),
	}

	s.log.Info("Starting server on port: " + server.Addr)
	if err := server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
