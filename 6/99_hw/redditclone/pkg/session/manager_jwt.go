package session

import (
	"context"
	"crypto/rsa"
	"fmt"
	"math/rand"
	"net/http"

	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	charSet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type SessionManager interface {
	Create(ctx context.Context, w http.ResponseWriter, u *User, sessID string, expiresAt int64) (string, error)
	Check(ctx context.Context, r *http.Request) (*Session, error)
	Destroy(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	DestroyAll(ctx context.Context, user *User) error
}

type SessionManagerJWT struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewSessionsJWTManager(privateKeyBytes, publicKeyBytes []byte) (*SessionManagerJWT, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)

	return &SessionManagerJWT{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (sm *SessionManagerJWT) Create(ctx context.Context, w http.ResponseWriter, user *User, sessID string, expiresAt int64) (string, error) {
	sess := &Session{
		User: &User{Username: user.Username, ID: user.ID},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	if sessID != "" {
		sess.SessionID = sessID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, sess)
	signed, err := token.SignedString(sm.privateKey)
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (sm *SessionManagerJWT) Check(ctx context.Context, request *http.Request) (*Session, error) {
	authHeader := request.Header.Get("authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	payload := &Session{}
	token, err := jwt.ParseWithClaims(tokenString, payload, func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodRSA)
		if !ok || method.Alg() != "RS256" {
			return nil, fmt.Errorf("bad sign method")
		}
		return sm.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return payload, nil
}

func (sm *SessionManagerJWT) Destroy(context.Context, http.ResponseWriter, *http.Request) error {
	// ¯\_(ツ)_/¯
	return nil
}

func (sm *SessionManagerJWT) DestroyAll(context.Context, *User) error {
	// ¯\_(ツ)_/¯
	return nil
}

func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	res := make([]byte, 0, 20)
	for i := 0; i < n; i++ {
		res[i] = charSet[rand.Intn(len(charSet))]
	}

	return string(res)
}
