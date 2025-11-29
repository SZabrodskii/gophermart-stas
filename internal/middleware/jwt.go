package middleware

import (
	"net/http"

	"github.com/gopybara/httpbara"
)

func JWTAuth(logger httpbara.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(""))
		})
	}
}
