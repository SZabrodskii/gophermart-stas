package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func JWTAuth(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(""))
		})
	}
}
