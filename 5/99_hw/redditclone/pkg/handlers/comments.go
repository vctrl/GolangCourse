package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"redditclone/pkg/comments"
	"redditclone/pkg/posts"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"time"

	"go.uber.org/zap"
)

type CommentHandler struct {
	CommentsRepo *comments.MemoryCommentsRepo
	PostsRepo    *posts.MemoryPostsRepo
	UsersRepo    *user.MemoryUsersRepo
	Logger       *zap.SugaredLogger
}

type AddCommentRequest struct {
	Comment string `json:"comment"`
}

func (h *CommentHandler) GetByPostId(id uint64) ([]*comments.Comment, error) {
	comments, _ := h.CommentsRepo.GetByPostID(id)
	return comments, nil
}

func (h *CommentHandler) Add(w http.ResponseWriter, r *http.Request) {
	postID, err := ParseUintParam(r, "post_id")
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

	sess, err := session.FromContext(r.Context())
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

	_, err = h.CommentsRepo.Add(comment)

	post, err := h.PostsRepo.GetById(postID)
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
	postID, err := ParseUintParam(r, "post_id")
	if err != nil {
		h.Logger.Errorf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	commentID, err := ParseUintParam(r, "comment_id")
	if err != nil {
		h.Logger.Errorf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok, err := h.CommentsRepo.Delete(postID, commentID)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !ok {
		WriteResponse(w, "Not found", http.StatusNotFound)
		return
	}

	post, err := h.PostsRepo.GetById(postID)
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

	w.WriteHeader(http.StatusCreated)
	w.Write(respBytes)
}
