package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"redditclone/pkg/comments"
	"redditclone/pkg/posts"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type PostHandler struct {
	Sm           session.SessionManager
	PostsRepo    PostsRepo
	UsersRepo    UsersRepo
	CommentsRepo CommentsRepo
	Logger       *zap.SugaredLogger
}

type PostsRepo interface {
	GetAll(context.Context) ([]*posts.Post, error)
	GetByID(context.Context, interface{}) (*posts.Post, error)
	GetByCategory(context.Context, string) ([]*posts.Post, error)
	GetByAuthorID(context.Context, interface{}) ([]*posts.Post, error)
	Add(context.Context, *posts.Post) (interface{}, error)
	Delete(context.Context, interface{}) (bool, error)
	Upvote(context.Context, interface{}, int64) (*posts.Post, error)
	DownVote(context.Context, interface{}, int64) (*posts.Post, error)
	Unvote(context.Context, interface{}, int64) (*posts.Post, error)

	ParseID(string) (interface{}, error)
}

type PostResponse struct {
	Score            int                `json:"score"`
	Views            uint64             `json:"views"`
	Type             posts.PostType     `json:"type"`
	Title            string             `json:"title"`
	Author           *Author            `json:"author"`
	Category         posts.PostCategory `json:"category"`
	URL              string             `json:"url,omitempty"`
	Text             string             `json:"text,omitempty"`
	Votes            []*posts.Vote      `json:"votes"`
	Comments         []*CommentResponse `json:"comments"`
	Created          time.Time          `json:"created"`
	UpvotePercentage uint8              `json:"upvotePercentage"`
	ID               interface{}        `json:"id"`
}

type CommentResponse struct {
	Created time.Time   `json:"created"`
	Author  *Author     `json:"author"`
	Body    string      `json:"body"`
	ID      interface{} `json:"id"`
}

type TokenDecoded struct {
	Payload *Payload `json:"payload"`
}

type Payload struct {
	User *user.User `json:"user"`
}

type Author struct {
	Username string `json:"username"`
	ID       int64  `json:"id"`
}

func mapToCommentsResponse(comments []*comments.Comment, usersRepo UsersRepo) ([]*CommentResponse, error) {
	result := make([]*CommentResponse, 0, len(comments))
	for _, c := range comments {
		author, err := usersRepo.GetByID(c.AuthorID)
		if err != nil {
			return nil, err
		}
		mapped := &CommentResponse{Created: c.Created, Author: &Author{Username: author.Username, ID: author.ID}, Body: c.Body, ID: c.ID}
		result = append(result, mapped)
	}

	return result, nil
}

type CreatePostReq struct {
	Category *string
	Type     *posts.PostType
	Title    *string
	URL      *string
	Text     *string
}

func (p *CreatePostReq) validate() []*CustomError {
	title := &Validator{value: p.Title, location: "body", field: "title"}
	titleErr := func() *CustomError {
		err := title.Required()
		if err != nil {
			return err
		}
		err = title.Empty()
		if err != nil {
			return err
		}
		err = title.MaxLength(100)
		if err != nil {
			return err
		}
		return title.Custom(func(value string) bool {
			return strings.TrimSpace(value) == value
		}, "cannot start or end with whitespace")
	}()

	var urlOrTextErr *CustomError
	if *p.Type == posts.PostType("text") {
		text := &Validator{value: p.Text, location: "body", field: "title"}
		urlOrTextErr = func() *CustomError {
			err := text.Required()
			if err != nil {
				return err
			}
			return text.MinLength(4)
		}()
	} else {
		url := &Validator{value: p.URL, location: "body", field: "title"}
		urlOrTextErr = func() *CustomError {
			err := url.Required()
			if err != nil {
				return err
			}
			return url.URL()
		}()
	}

	category := &Validator{value: p.Category, location: "body", field: "title"}
	categoryErr := func() *CustomError {
		err := category.Required()
		if err != nil {
			return err
		}
		return category.Empty()
	}()

	return mergeErrors(titleErr, urlOrTextErr, categoryErr)
}

func (h *PostHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	postsDb, err := h.PostsRepo.GetAll(ctx)
	postsResp, err := h.getPostsWithData(postsDb)

	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	postsBytes, err := json.Marshal(postsResp)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(postsBytes)
}

func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.PostsRepo.ParseID(mux.Vars(r)["id"])
	if err != nil {
		h.Logger.Error(err.Error())
		WriteResponse(w, "invalid post id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	post, err := h.PostsRepo.GetByID(ctx, id)
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

	postBytes, err := json.Marshal(postWithData)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(postBytes)
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteResponse(w, "bad request", http.StatusBadRequest)
		return
	}

	var req CreatePostReq
	err = json.Unmarshal(body, &req)
	if err != nil {
		WriteResponse(w, "bad request", http.StatusBadRequest)
		return
	}

	validationErrors := req.validate()

	if len(validationErrors) > 0 {
		writeErrorsResponse(w, validationErrors, http.StatusUnprocessableEntity)
		return
	}

	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	post := &posts.Post{Views: 0,
		Score: 0, Type: *req.Type,
		Title:    *req.Title,
		AuthorID: sess.User.ID,
		Category: posts.PostCategory(*req.Category),
		Created:  time.Now(),
	}

	if *req.Type == posts.Text {
		post.Text = *req.Text
	} else {
		post.URL = *req.URL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id, err := h.PostsRepo.Add(ctx, post)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	post.ID = id

	u, err := h.UsersRepo.GetByID(sess.User.ID)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	postResp, err := MapToPostResponse(post, u.ID, u.Username, []*comments.Comment{}, h.UsersRepo)

	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respBytes, err := json.Marshal(postResp)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(respBytes)
}

func (h *PostHandler) GetPostsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	posts, err := h.PostsRepo.GetByCategory(ctx, category)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	postsWithData, err := h.getPostsWithData(posts)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respBytes, err := json.Marshal(postsWithData)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func (h *PostHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	user, err := h.UsersRepo.GetByUsername(username)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	posts, err := h.PostsRepo.GetByAuthorID(ctx, user.ID)

	postsWithData, err := h.getPostsWithData(posts)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	postsBytes, err := json.Marshal(postsWithData)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(postsBytes)
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := h.PostsRepo.ParseID(mux.Vars(r)["id"])
	if err != nil {
		h.Logger.Error(err.Error())
		WriteResponse(w, "invalid post id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ok, err := h.PostsRepo.Delete(ctx, id)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if ok {
		WriteResponse(w, "success", http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
	WriteResponse(w, "post not found", http.StatusNotFound)
}

func (h *PostHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	h.vote(w, r, h.PostsRepo.Upvote)
}

func (h *PostHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	h.vote(w, r, h.PostsRepo.DownVote)
}

func (h *PostHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	h.vote(w, r, h.PostsRepo.Unvote)
}

func calculateUpvotePercentage(postVotes []*posts.Vote) uint8 {
	if len(postVotes) == 0 {
		return uint8(0)
	}

	upvoteCnt := 0
	for _, v := range postVotes {
		if v.Vote == posts.Upvote {
			upvoteCnt++
		}
	}

	return uint8(math.Round((float64(upvoteCnt) / float64(len(postVotes))) * 100))
}

func (h *PostHandler) getPostsWithData(postsDb []*posts.Post) ([]*PostResponse, error) {
	result := make([]*PostResponse, 0, len(postsDb))
	for _, p := range postsDb {
		postWithData, err := getPostData(p, h.UsersRepo, h.CommentsRepo)
		if err != nil {
			return nil, err
		}

		result = append(result, postWithData)
	}

	return result, nil
}

func (h *PostHandler) vote(w http.ResponseWriter, r *http.Request,
	voteRepo func(context.Context, interface{}, int64) (*posts.Post, error)) {
	id, err := h.PostsRepo.ParseID(mux.Vars(r)["post_id"])

	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	post, err := voteRepo(ctx, id, sess.User.ID)

	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	comments, err := h.CommentsRepo.GetByPostID(ctx, post.ID)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	u, err := h.UsersRepo.GetByID(sess.User.ID)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := MapToPostResponse(post, u.ID, u.Username, comments, h.UsersRepo)

	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		h.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resBytes)
}
