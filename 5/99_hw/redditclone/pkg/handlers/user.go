package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
)

type UserHandler struct {
	Sm     *session.SessionManagerJWT
	Repo   *user.MemoryUsersRepo
	Logger *zap.SugaredLogger
}

type AuthReq struct {
	Password *string `json:"password"`
	Username *string `json:"username"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (r *AuthReq) validate() []*CustomError {
	usr := &Validator{value: r.Username, location: "body", field: "username"}
	// todo refactoring
	// в общий функционал немного сложно выносить, т.к. функции имеют разные сигнатуры
	usrErr := func() *CustomError {
		err := usr.Required()
		if err != nil {
			return err
		}
		err = usr.Empty()
		if err != nil {
			return err
		}
		err = usr.MaxLength(32)
		if err != nil {
			return err
		}
		err = usr.Custom(func(value string) bool {
			return strings.TrimSpace(value) == value
		}, "cannot start or end with whitespace")

		if err != nil {
			return err
		}

		return usr.Matches("^[a-zA-Z0-9_-]+$")
	}()

	pwd := &Validator{value: r.Password, location: "body", field: "password"}
	pwdErr := func() *CustomError {
		err := pwd.Required()
		if err != nil {
			return err
		}
		err = pwd.Empty()
		if err != nil {
			return err
		}
		err = pwd.MinLength(8)
		if err != nil {
			return err
		}
		return pwd.MaxLength(72)
	}()

	return mergeErrors(usrErr, pwdErr)
}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		u.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var authReq AuthReq
	err = json.Unmarshal(body, &authReq)
	if err != nil {
		WriteResponse(w, "bad request", http.StatusBadRequest)
		return
	}

	validationErrors := authReq.validate()

	if len(validationErrors) > 0 {
		writeErrorsResponse(w, validationErrors, http.StatusUnprocessableEntity)
		return
	}

	user, err := u.Repo.GetByUsername(*authReq.Username)

	if err != nil {
		u.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user == nil {
		WriteResponse(w, "user not found", http.StatusUnauthorized)
		return
	}

	if !checkPass(user.Password, *authReq.Password) {
		WriteResponse(w, "invalid password", http.StatusUnauthorized)
		return
	}

	u.writeAuthResponse(w, user, http.StatusOK)
}

func (u *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		u.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var authReq AuthReq
	err = json.Unmarshal(body, &authReq)

	if err != nil {
		WriteResponse(w, "bad request", http.StatusBadRequest)
		return
	}

	validationErrors := authReq.validate()
	if len(validationErrors) > 0 {
		writeErrorsResponse(w, validationErrors, http.StatusUnprocessableEntity)
		return
	}

	existUser, err := u.Repo.GetByUsername(*authReq.Username)
	if err != nil {
		u.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if existUser != nil {
		validationError := &CustomError{Location: "body", Param: "username", Value: *authReq.Username, Msg: "already exists"}
		writeErrorsResponse(w, []*CustomError{validationError}, http.StatusUnprocessableEntity)
		return
	}

	salt := make([]byte, 8)
	rand.Read(salt)

	passHash := hashPass(salt, *authReq.Password)

	user := &user.User{
		Username: *authReq.Username,
		Password: passHash,
	}

	_, err = u.Repo.Add(user)
	if err != nil {
		u.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	u.writeAuthResponse(w, user, http.StatusCreated)
}

func hashPass(salt []byte, plainPassword string) []byte {
	hashedPass := argon2.IDKey([]byte(plainPassword), []byte(salt), 1, 64*1024, 4, 32)
	return append(salt, hashedPass...)
}

func checkPass(passHash []byte, plainPassword string) bool {
	salt := passHash[0:8]
	newSalt := make([]byte, len(salt))
	copy(newSalt, salt)
	usersPassHash := hashPass(newSalt, plainPassword)
	return bytes.Equal(usersPassHash, passHash)
}

func (u *UserHandler) writeAuthResponse(w http.ResponseWriter, user *user.User, status int) {
	token, err := u.Sm.Create(w, user)
	if err != nil {
		u.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := &AuthResponse{Token: token}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		u.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respBytes)
}
