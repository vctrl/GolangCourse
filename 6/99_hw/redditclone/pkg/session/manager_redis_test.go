package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/elliotchance/redismock/v8"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
)

var token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjozNCwidXNlcm5hbWUiOiJ2ZWN0b3JlYWwifSwiU2Vzc2lvbklEIjoiNDgwZjA4ODYtYmJiYi00MGU4LTljMmItYTQ3ZThhYTdhNjY2IiwiZXhwIjozMjQ5OTg2NjA5OH0.VA5mS8vKfpRJ2ZaNgWxYWFKw5HlAxu9B9EyEheLUcp5E-om8dfJEN-Q020bKPPLiw5QgjDhuSD_wX9ONa2h2v_uVYqdsfcWoBHtmgyHP5HOWixwwFXtm3_verX0Ip59SPr85kUEUDgPQqacUYrObCru1Q8tJsfKDDRnwREaaxMQ"
var sessID = "480f0886-bbbb-40e8-9c2b-a47e8aa7a666"
var expiresAt = time.Date(2999, 11, 17, 20, 34, 58, 651387237, time.UTC)
var u = &User{Username: "vectoreal", ID: 34}

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	jwtMock := NewMockSessionManager(ctrl)

	ctx := context.Background()
	w := httptest.NewRecorder()

	jwtMock.EXPECT().Create(ctx, w, u, sessID, expiresAt.Unix()).Return(token, nil)

	mock := redismock.NewMock()
	sm := NewSessionManagerRedis(mock, jwtMock)

	mock.On("Set", ctx, sessID, u.ID, time.Duration(0)).Return(redis.NewStatusCmd(ctx, "set", sessID, u.ID))
	mock.On("SAdd", ctx, strconv.FormatInt(u.ID, 10), []interface{}{sessID}).Return(redis.NewIntCmd(ctx, "sadd", strconv.FormatInt(u.ID, 10), sessID))

	fact, err := sm.Create(ctx, w, u, sessID, expiresAt.Unix())
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}

	if fact != token {
		t.Errorf("expected %v but was %v", token, fact)
	}
}

func TestCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	jwtMock := NewMockSessionManager(ctrl)

	mock := redismock.NewMock()

	sm := NewSessionManagerRedis(mock, jwtMock)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	sess := &Session{
		User:           &User{ID: 34, Username: "vectoreal"},
		SessionID:      "480f0886-bbbb-40e8-9c2b-a47e8aa7a666",
		StandardClaims: jwt.StandardClaims{ExpiresAt: 32499866098},
	}

	jwtMock.EXPECT().Check(ctx, r).Return(sess, nil)
	mock.On("Get", ctx, sessID).Return(redis.NewStringResult(strconv.FormatInt(u.ID, 10), nil))

	fact, err := sm.Check(ctx, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}

	if fact != sess {
		t.Errorf("expected %v but was %v", sess, fact)
	}
}

func TestDestroy(t *testing.T) {
	ctrl := gomock.NewController(t)
	jwtMock := NewMockSessionManager(ctrl)
	sess := &Session{
		User:           &User{ID: 34, Username: "vectoreal"},
		SessionID:      "480f0886-bbbb-40e8-9c2b-a47e8aa7a666",
		StandardClaims: jwt.StandardClaims{ExpiresAt: 32499866098},
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), SessionKey, sess)
	r = r.WithContext(ctx)

	mock := redismock.NewMock()
	sm := NewSessionManagerRedis(mock, jwtMock)
	w := httptest.NewRecorder()

	jwtMock.EXPECT().Destroy(ctx, w, r).Return(nil)
	mock.On("Del", ctx, []string{sessID}).Return(redis.NewIntResult(1, nil))
	err := sm.Destroy(ctx, w, r)

	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func TestDestroyAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	jwtMock := NewMockSessionManager(ctrl)
	sess := &Session{
		User:           &User{ID: 34, Username: "vectoreal"},
		SessionID:      "480f0886-bbbb-40e8-9c2b-a47e8aa7a666",
		StandardClaims: jwt.StandardClaims{ExpiresAt: 32499866098},
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), SessionKey, sess)
	mock := redismock.NewMock()
	sm := NewSessionManagerRedis(mock, jwtMock)

	mock.On("SMembers", ctx, strconv.FormatInt(u.ID, 10)).Return(redis.NewStringSliceResult([]string{sessID}, nil))
	mock.On("Del", ctx, []string{sessID}).Return(redis.NewIntResult(1, nil))

	err := sm.DestroyAll(ctx, u)

	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}
