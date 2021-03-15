package session

import "github.com/dgrijalva/jwt-go"

type key int

const (
	SessionKey key = 1
)

type Session struct {
	User *User `json:"user"`
	jwt.StandardClaims
}

type User struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
}
