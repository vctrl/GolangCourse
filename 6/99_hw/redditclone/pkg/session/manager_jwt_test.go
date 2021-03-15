package session

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var expectedToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjozNCwidXNlcm5hbWUiOiJ2ZWN0b3JlYWwifSwiU2Vzc2lvbklEIjoiNDgwZjA4ODYtYmJiYi00MGU4LTljMmItYTQ3ZThhYTdhNjY2IiwiZXhwIjozMjQ5OTg2NjA5OH0.VA5mS8vKfpRJ2ZaNgWxYWFKw5HlAxu9B9EyEheLUcp5E-om8dfJEN-Q020bKPPLiw5QgjDhuSD_wX9ONa2h2v_uVYqdsfcWoBHtmgyHP5HOWixwwFXtm3_verX0Ip59SPr85kUEUDgPQqacUYrObCru1Q8tJsfKDDRnwREaaxMQ"
var testTime = time.Date(2999, 11, 17, 20, 34, 58, 651387237, time.UTC)
var testTimeExpired = time.Date(1999, 11, 17, 20, 34, 58, 651387237, time.UTC)
var expiredToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjozNCwidXNlcm5hbWUiOiJ2ZWN0b3JlYWwifSwiU2Vzc2lvbklEIjoiNDgwZjA4ODYtYmJiYi00MGU4LTljMmItYTQ3ZThhYTdhNjY2IiwiZXhwIjo5NDI4NzA4OTh9.N_plMSerRZAWHDIPLV1mfBCeiTyKZ5VQKcKTquDl7WFTnxRsEwhWfjY3y3ciMobwJS6B3aGYPnoFQ-fOkOOwp1XfAjNvAHEL2T8RvDFzEChBtI9U9DMGtTKVW0HydeBldoY42MDUWjib6yLbyC72tAjfjs_Ws6R3FQNNm6pUwoQ"

func TestCreateJWT(t *testing.T) {
	sm, err := NewTestSessionManager()
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}

	ctx := context.Background()
	w := httptest.NewRecorder()
	u := &User{Username: "vectoreal", ID: 34}
	sessID := "480f0886-bbbb-40e8-9c2b-a47e8aa7a666"

	token, err := sm.Create(ctx, w, u, sessID, testTime.Unix())

	if token != expectedToken {
		t.Errorf("test fail, expected token: %v, but was: %v", expectedToken, token)
	}
}

func TestCheckJWT(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+expectedToken)
	sm, err := NewTestSessionManager()

	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}

	ctx := context.Background()
	sess, err := sm.Check(ctx, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}

	expected := &Session{User: &User{ID: 34, Username: "vectoreal"}, SessionID: "480f0886-bbbb-40e8-9c2b-a47e8aa7a666", StandardClaims: jwt.StandardClaims{ExpiresAt: 32499866098}}
	if !reflect.DeepEqual(sess, expected) {
		t.Errorf("test fail, expected %v but was %v", expected, sess)
	}
}

func TestCheckJWTExpired(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+expiredToken)
	sm, err := NewTestSessionManager()

	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}

	ctx := context.Background()
	_, err = sm.Check(ctx, r)
	if err == nil {
		t.Fatal("expected expired token error, but was nil")
	}

	verr, ok := err.(*jwt.ValidationError)
	if !ok {
		t.Fatalf("expected jwt validation error, but was %v", err)
	}

	if verr.Errors&jwt.ValidationErrorExpired != jwt.ValidationErrorExpired {
		t.Fatalf("expected jwt expired error, but was %v", verr.Errors)
	}
}

func NewTestSessionManager() (*SessionManagerJWT, error) {
	testPrivateKeyBytes, err := ioutil.ReadFile("test_key.rsa")
	if err != nil {
		return nil, err
	}

	testPublicKeyBytes, err := ioutil.ReadFile("test_key.rsa.pub")
	if err != nil {
		return nil, err
	}

	sm, err := NewSessionsJWTManager(testPrivateKeyBytes, testPublicKeyBytes)
	if err != nil {
		return nil, err
	}

	return sm, nil
}
