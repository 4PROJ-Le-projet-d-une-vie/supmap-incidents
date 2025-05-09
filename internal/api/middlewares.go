package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/matheodrd/httphelper/handler"
	"io"
	"net/http"
	"supmap-users/internal/models/dto"
)

func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions { // Ignore preflight requests because OPTIONS handler is not implemented
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type AuthError struct {
	Error string `json:"error"`
}

var missingHeader = &AuthError{Error: "Authorization Header is missing"}
var sessionExpired = &AuthError{Error: "session is expired"}
var invalidToken = &AuthError{Error: "invalid token"}
var invalidUser = &AuthError{Error: "invalid user"}

func (s *Server) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				if err := handler.Encode(missingHeader, http.StatusBadRequest, w); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			}

			// Requête vers le service users
			req, err := http.NewRequestWithContext(r.Context(), "GET", fmt.Sprintf("%s/internal/users/check-auth", s.Config.UsersBaseUrl), nil)
			if err != nil {
				s.log.Error("failed to create auth check request", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Header.Set("Authorization", authHeader)

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				s.log.Error("failed to check auth", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}(res.Body)

			if res.StatusCode != http.StatusOK {
				w.WriteHeader(res.StatusCode)
				switch res.StatusCode {
				case http.StatusUnauthorized:
					err = json.NewEncoder(w).Encode(invalidToken)
				case http.StatusForbidden:
					err = json.NewEncoder(w).Encode(sessionExpired)
				default:
					err = json.NewEncoder(w).Encode(invalidUser)
				}

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}

				return
			}

			// Désérialisation de l'utilisateur
			var user dto.PartialUserDTO
			if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
				s.log.Error("failed to decode user from auth response", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), "user", &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *Server) AdminMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value("user").(*dto.PartialUserDTO)
			if !ok {
				s.log.Warn("unauthenticated user tried to access admin route")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if user.Role == nil || user.Role.Name != "ROLE_ADMIN" {
				s.log.Warn("Non admin user tried to access admin route")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
