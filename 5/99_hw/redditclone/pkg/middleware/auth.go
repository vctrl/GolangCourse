package middleware

import (
	"context"
	"net/http"
	"redditclone/pkg/session"
	"strings"

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

		sess, err := sm.Check(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), session.SessionKey, sess)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
