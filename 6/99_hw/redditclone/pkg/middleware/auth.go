package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"redditclone/pkg/session"
	"strings"
	"time"

	"go.uber.org/zap"
)

type key int

const (
	keyUserID key = iota
	keyUsername
)

var authRoutes = map[string]string{
	"/api/posts": http.MethodPost,
}

func Auth(logger *zap.SugaredLogger, sm session.SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, ok := authRoutes[r.URL.Path]

		if (!ok || m != r.Method) && !strings.HasSuffix(r.URL.Path, "vote") &&
			(!strings.HasPrefix(r.URL.Path, "/api/post/") || r.Method != http.MethodPost) {
			next.ServeHTTP(w, r)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		sess, err := sm.Check(ctx, r)
		if err != nil {
			logger.Error(err.Error())
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			errorBody, _ := json.Marshal(map[string]string{"message": "unauthorized"})
			w.Write(errorBody)

			return
		}

		ctx = context.WithValue(r.Context(), session.SessionKey, sess)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
