package session

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type Cmdable interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
}

type SessionManagerRedis struct {
	rdb    Cmdable
	jwt    SessionManager
	logger *zap.SugaredLogger
}

func NewSessionManagerRedis(rdb Cmdable, jwt SessionManager) *SessionManagerRedis {
	return &SessionManagerRedis{rdb: rdb, jwt: jwt}
}

func (sm *SessionManagerRedis) Create(ctx context.Context, w http.ResponseWriter, u *User, sessID string, expiresAt int64) (string, error) {
	token, err := sm.jwt.Create(ctx, w, u, sessID, expiresAt)
	if err != nil {
		return "", err
	}

	err = sm.rdb.Set(ctx, sessID, u.ID, 0).Err()
	if err != nil {
		return "", err
	}

	err = sm.rdb.SAdd(ctx, strconv.FormatInt(u.ID, 10), sessID).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (sm *SessionManagerRedis) Check(ctx context.Context, r *http.Request) (*Session, error) {
	sess, err := sm.jwt.Check(ctx, r)
	if err != nil {
		return nil, err
	}
	userIDStr, err := sm.rdb.Get(ctx, sess.SessionID).Result()
	if err != nil {
		return nil, err
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 0)
	if err != nil {
		return nil, err
	}
	if userID != sess.User.ID {
		return nil, errors.New("wrong user")
	}

	return sess, nil
}

func (sm *SessionManagerRedis) Destroy(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	sess, err := SessionFromContext(r.Context())
	if err != nil {
		return err
	}

	err = sm.rdb.Del(ctx, sess.SessionID).Err()
	if err != nil {
		return err
	}

	return nil
}

func (sm *SessionManagerRedis) DestroyAll(ctx context.Context, user *User) error {
	sessionIDs, err := sm.rdb.SMembers(ctx, strconv.FormatInt(user.ID, 10)).Result()
	if err != nil {
		return err
	}

	err = sm.rdb.Del(ctx, sessionIDs...).Err()
	if err != nil {
		return err
	}

	return nil
}
