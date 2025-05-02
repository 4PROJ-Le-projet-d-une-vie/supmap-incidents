package api

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type AuthError struct {
	Error string `json:"error"`
}

var sessionExpired = &AuthError{Error: "session is expired"}
var invalidToken = &AuthError{Error: "invalid token"}
var invalidUser = &AuthError{Error: "invalid user"}

func (s *Server) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO implement authentication middleware
			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}

func (s *Server) AdminMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO implement admin authentication check
			next.ServeHTTP(w, r)
		})
	}
}

func decodeJWT(tokenStr string, secret string) (*int64, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	userId, ok := claims["userId"].(float64)
	if !ok {
		return nil, fmt.Errorf("userId is missing or of wrong type")
	}

	convertedId := int64(userId)
	return &convertedId, nil
}
