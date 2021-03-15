package session

import (
	"context"
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

type key int

const (
	SessionKey key = 1
)

type Session struct {
	User      *User `json:"user"`
	SessionID string
	jwt.StandardClaims
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(SessionKey).(*Session)
	if !ok {
		return nil, fmt.Errorf("Session not found")
	}

	return sess, nil
}
