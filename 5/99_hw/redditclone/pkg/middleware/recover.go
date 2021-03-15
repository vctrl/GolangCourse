package middleware

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func Recover(logger *zap.SugaredLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err)

				resp, _ := json.Marshal(map[string]string{"error": "Internal server error"})
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(resp)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
