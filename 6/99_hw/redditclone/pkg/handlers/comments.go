package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"redditclone/pkg/comments"
	"redditclone/pkg/session"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type CommentHandler struct {
	CommentsRepo CommentsRepo
	PostsRepo    PostsRepo
	UsersRepo    UsersRepo
	Logger       *zap.SugaredLogger
}

type AddCommentRequest struct {
	Comment string `json:"comment"`
}

type CommentsRepo interface {
	GetByPostID(context.Context, interface{}) ([]*comments.Comment, error)
	GetByID(context.Context, interface{}) (*comments.Comment, error)
	Add(context.Context, *comments.Comment) (interface{}, error)
	Delete(context.Context, interface{}) (bool, error)

	ParseID(in string) (interface{}, error)
}

func (h *CommentHandler) GetByPostID(id uint64) ([]*comments.Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	comments, _ := h.CommentsRepo.GetByPostID(ctx, id)
	return comments, nil
}

func (h *CommentHandler) Add(w http.ResponseWriter, r *http.Request) {
	postID, err := h.PostsRepo.ParseID(mux.Vars(r)["post_id"])
	if err != nil {
		h.Logger.Errorf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var req AddCommentRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	comment := &comments.Comment{
		Created:  time.Now(),
		AuthorID: sess.User.ID,
		Body:     req.Comment,
		PostID:   postID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = h.CommentsRepo.Add(ctx, comment)

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	post, err := h.PostsRepo.GetByID(ctx, postID)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	postWithData, err := getPostData(post, h.UsersRepo, h.CommentsRepo)

	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respBytes, err := json.Marshal(postWithData)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(respBytes)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	postID, err := h.PostsRepo.ParseID(mux.Vars(r)["post_id"])
	if err != nil {
		h.Logger.Errorf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	commentID, err := h.CommentsRepo.ParseID(mux.Vars(r)["comment_id"])
	if err != nil {
		h.Logger.Errorf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ok, err := h.CommentsRepo.Delete(ctx, commentID)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !ok {
		WriteResponse(w, "Not found", http.StatusNotFound)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	post, err := h.PostsRepo.GetByID(ctx, postID)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	postWithData, err := getPostData(post, h.UsersRepo, h.CommentsRepo)

	respBytes, err := json.Marshal(postWithData)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}
