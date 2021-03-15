package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

var username = "vectoreal"
var password = "secret_password"
var token = "test_token"
var passwordDB = HashPass(getSalt(), password)

// todo дублирование кода
func getSalt() []byte {
	salt := make([]byte, 8)
	rand.Read(salt)
	return salt
}

type Case struct {
	name             string
	expectedRepoUser *user.User
	execHandler      func(h *UserHandler, w http.ResponseWriter, r *http.Request)
	expectedResponse []byte
	expectedStatus   int
}

var cases = []Case{
	{
		name:             "LoginHappyCase",
		expectedRepoUser: &user.User{Username: username, Password: passwordDB, ID: int64(1)},
		execHandler: func(h *UserHandler, w http.ResponseWriter, r *http.Request) {
			h.Login(w, r)
		},
		expectedResponse: []byte(`{"token":"test_token"}`),
		expectedStatus:   http.StatusOK,
	},
	{
		name:             "LoginUserNotExistCase",
		expectedRepoUser: nil,
		execHandler: func(h *UserHandler, w http.ResponseWriter, r *http.Request) {
			h.Login(w, r)
		},
		expectedResponse: []byte(`{"message":"user not found"}`),
		expectedStatus:   http.StatusUnauthorized,
	},
	{
		name:             "RegisterHappyCase",
		expectedRepoUser: nil,
		execHandler: func(h *UserHandler, w http.ResponseWriter, r *http.Request) {
			h.Register(w, r)
		},
		expectedResponse: []byte(`{"token":"test_token"}`),
		expectedStatus:   http.StatusCreated,
	},
	{
		name:             "RegisterUserAlreadyExistCase",
		expectedRepoUser: &user.User{Username: username, Password: passwordDB, ID: int64(1)},
		execHandler: func(h *UserHandler, w http.ResponseWriter, r *http.Request) {
			h.Register(w, r)
		},
		expectedResponse: []byte(`{"errors":[{"location":"body","param":"username","value":"vectoreal","msg":"already exists"}]}`),
		expectedStatus:   http.StatusUnprocessableEntity,
	},
}

func TestLogin(t *testing.T) {
	for _, tc := range cases {
		ctrl := gomock.NewController(t)
		repo := NewMockUsersRepo(ctrl)
		sm := session.NewMockSessionManager(ctrl)
		h := &UserHandler{Sm: sm, Repo: repo, Logger: zap.NewNop().Sugar()}
		w := httptest.NewRecorder()

		body := map[string]string{"username": username, "password": password}
		bodyBytes, _ := json.Marshal(body)
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(bodyBytes))

		sm.EXPECT().
			Create(gomock.Any(),
				w, &session.User{ID: int64(1), Username: username},
				gomock.Any(), gomock.Any()).
			Return(token, nil)

		repo.EXPECT().GetByUsername(username).Return(tc.expectedRepoUser, nil)
		repo.EXPECT().Add(gomock.Any()).Return(int64(1), nil)

		tc.execHandler(h, w, r)

		if w.Result().StatusCode != tc.expectedStatus {
			t.Fatalf("wrong status code: %d, but expected %d", w.Result().StatusCode, tc.expectedStatus)
		}

		res, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatalf("unexpected error while reading response body: %s", err.Error())
		}

		if !reflect.DeepEqual(res, tc.expectedResponse) {
			t.Fatalf("unexpected response: %s but expected %s", res, tc.expectedResponse)
		}
	}

}
