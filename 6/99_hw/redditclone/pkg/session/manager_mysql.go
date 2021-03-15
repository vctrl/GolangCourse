package session

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type SessionDB struct {
	ID     string
	UserID int64
}

type SessionManagerSQL struct {
	db  *sql.DB
	jwt *SessionManagerJWT
}

func NewSessionManagerSQL(db *sql.DB) (*SessionManagerSQL, error) {
	return &SessionManagerSQL{db: db}, nil
}

func (sm *SessionManagerSQL) Create(w http.ResponseWriter, u *User, expiresAt int64) (string, error) {
	sessID := uuid.New().String()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	token, err := sm.jwt.Create(ctx, w, u, sessID, expiresAt)
	if err != nil {
		return "", err
	}

	_, err = sm.db.Exec("INSERT INTO sessions (`id`, `user_id`) VALUES (?, ?)", sessID, u.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (sm *SessionManagerSQL) Check(r *http.Request) (*Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sess, err := sm.jwt.Check(ctx, r)
	if err != nil {
		return nil, err
	}

	sessDB := &SessionDB{}

	err = sm.db.QueryRow("SELECT `id`, `user_id` FROM sessions WHERE id = ?", sess.SessionID).Scan(&sessDB.ID, &sessDB.UserID)
	if err != nil {
		return nil, err
	}

	if sessDB.UserID != sess.User.ID {
		return nil, errors.New("wrong user")
	}

	return sess, nil
}

func (sm *SessionManagerSQL) Destroy(ctx context.Context, r *http.Request) error {
	sess, err := SessionFromContext(r.Context())
	if err != nil {
		return err
	}

	_, err = sm.db.Exec("DELETE FROM sessions WHERE id = ?", sess.SessionID)
	if err != nil {
		return err
	}

	return nil
}

func (sm *SessionManagerSQL) DestroyAll(ctx context.Context, user *User) error {
	_, err := sm.db.Exec("DELETE FROM sessions WHERE user_id = ?", user.ID)
	if err != nil {
		return err
	}

	return nil
}
