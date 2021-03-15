package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/comments"
	"redditclone/pkg/posts"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type CommentCase struct {
	name        string
	execHandler func(h *CommentHandler, w http.ResponseWriter, r *http.Request)
	method      string
	status      int
	body        map[string]string
}

var commentCases = []CommentCase{
	{
		name: "AddCommentHappyCase",
		execHandler: func(h *CommentHandler, w http.ResponseWriter, r *http.Request) {
			h.Add(w, r)
		},
		method: http.MethodPost,
		status: http.StatusCreated,
		body:   map[string]string{"comment": "comment 1"},
	},
	{
		name: "DeleteCommentHappyCase",
		execHandler: func(h *CommentHandler, w http.ResponseWriter, r *http.Request) {
			h.Delete(w, r)
		},
		method: http.MethodDelete,
		status: http.StatusOK,
		body:   nil,
	},
}

func TestCommentHandler(t *testing.T) {
	for i, tc := range commentCases {
		ctrl := gomock.NewController(t)

		repo := NewMockCommentsRepo(ctrl)
		postsRepo := NewMockPostsRepo(ctrl)
		usersRepo := NewMockUsersRepo(ctrl)

		commentID := uuid.New()
		postID := uuid.New()

		layout := "2006-01-02T15:04:05.000Z"
		timeCreated, _ := time.Parse(layout, "2020-10-10T18:22:22.222Z")
		commentText := "comment 1"
		comment := &comments.Comment{Created: timeCreated,
			AuthorID: int64(1),
			Body:     commentText,
			ID:       commentID.String(),
			PostID:   postID}

		post := &posts.Post{
			ID:       postID,
			Score:    1,
			Views:    1,
			Type:     posts.Text,
			Title:    "test",
			AuthorID: int64(1),
			Category: posts.Fashion,
			Text:     "test some test some test",
			Votes:    make(map[int64]posts.VoteValue),
			Created:  timeCreated,
		}

		userID := int64(1)
		user := &user.User{
			Username: "vectoreal",
			Password: []byte("somepass"),
			ID:       userID,
		}

		repo.EXPECT().GetByPostID(gomock.Any(), postID).Return([]*comments.Comment{comment}, nil)
		repo.EXPECT().ParseID(commentID.String()).Return(commentID, nil)
		repo.EXPECT().Delete(gomock.Any(), commentID).Return(true, nil)
		repo.EXPECT().Add(gomock.Any(), gomock.AssignableToTypeOf(comment)).Return(gomock.Any(), nil)
		postsRepo.EXPECT().ParseID(postID.String()).Return(postID, nil)
		postsRepo.EXPECT().GetByID(gomock.Any(), postID).Return(post, nil)
		usersRepo.EXPECT().GetByID(userID).Return(user, nil).AnyTimes()

		w := httptest.NewRecorder()

		var r *http.Request
		if tc.body != nil {
			body, _ := json.Marshal(tc.body)
			r = httptest.NewRequest(tc.method, "/", bytes.NewBuffer(body))
		} else {
			r = httptest.NewRequest(tc.method, "/", nil)
		}

		r = r.WithContext(context.WithValue(r.Context(), session.SessionKey, &session.Session{User: &session.User{ID: userID, Username: user.Username}}))

		h := &CommentHandler{
			CommentsRepo: repo,
			PostsRepo:    postsRepo,
			UsersRepo:    usersRepo,
			Logger:       zap.NewNop().Sugar(),
		}

		vars := map[string]string{
			"post_id":    postID.String(),
			"comment_id": commentID.String(),
		}
		r = mux.SetURLVars(r, vars)

		tc.execHandler(h, w, r)

		if w.Result().StatusCode != tc.status {
			t.Fatalf("test case %d %s status code check failed: expected %d but was %d", i, tc.name, tc.status, w.Result().StatusCode)
		}

		res, _ := ioutil.ReadAll(w.Result().Body)

		expected := []byte(fmt.Sprintf(`{"score":0,"views":1,"type":"text","title":"test","author":{"username":"vectoreal","id":1},"category":"fashion","text":"test some test some test","votes":[],"comments":[{"created":"2020-10-10T18:22:22.222Z","author":{"username":"vectoreal","id":1},"body":"comment 1","id":"%s"}],"created":"2020-10-10T18:22:22.222Z","upvotePercentage":0,"id":"%s"}`,
			commentID.String(), postID.String()))
		if !reflect.DeepEqual(res, expected) {
			t.Fatalf("test case %d %s failed: expected %s but was %s", i, tc.name, expected, res)
		}
	}
}
