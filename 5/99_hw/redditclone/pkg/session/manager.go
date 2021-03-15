package session

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"redditclone/pkg/user"

	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type SessionManager interface {
	Create(http.ResponseWriter, *user.User) (string, error)
	Check(*http.Request) (*Session, error)
	Destroy(w http.ResponseWriter) error
	DestroyAll() error
}

type SessionManagerJWT struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewSessionsJWTManager() (*SessionManagerJWT, error) {
	privateKeyBytes, err := ioutil.ReadFile("key.rsa")
	if err != nil {
		return nil, err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	publicKeyBytes, err := ioutil.ReadFile("key.rsa.pub")
	if err != nil {
		return nil, err
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)

	return &SessionManagerJWT{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

const (
	charSet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func (sm *SessionManagerJWT) Create(w http.ResponseWriter, user *user.User) (string, error) {
	sess := &Session{
		User: &User{Username: user.Username, ID: user.ID},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, sess)
	signed, err := token.SignedString(sm.privateKey)
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (sm *SessionManagerJWT) Check(request *http.Request) (*Session, error) {
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

func (sm *SessionManagerJWT) Destroy(w http.ResponseWriter) error {
	return nil
}

func (sm *SessionManagerJWT) DestroyAll() error {
	return nil
}

func FromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(SessionKey).(*Session)
	if !ok {
		return nil, fmt.Errorf("can't cast value to session type")
	}

	return sess, nil
}

func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	res := make([]byte, 0, 20)
	for i := 0; i < n; i++ {
		res[i] = charSet[rand.Intn(len(charSet))]
	}

	return string(res)
}
