package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/comments"
	"redditclone/pkg/posts"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	reflect "reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var postIDs = []primitive.ObjectID{primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()}
var userIDs = []int64{int64(1), int64(2), int64(3)}
var commentIDs = []primitive.ObjectID{primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()}

var dateFormat = "2006-01-02 15:04:05.999999999 -0700 MST"
var t, _ = time.Parse(dateFormat, strings.Split(time.Now().String(), " m=")[0])
var testPostData = []*posts.Post{
	{
		ID:       postIDs[0],
		Score:    3,
		Views:    123,
		Type:     posts.Text,
		Title:    "Fashion title 1",
		AuthorID: userIDs[0],
		Category: posts.Fashion,
		Text:     "test",
		Created:  t,
		Votes:    map[int64]posts.VoteValue{int64(1): posts.VoteValue(1), int64(2): posts.VoteValue(1), int64(3): posts.VoteValue(1)},
	},
	{
		ID:       postIDs[1],
		Score:    0,
		Views:    456,
		Type:     posts.Text,
		Title:    "Funny title 2",
		AuthorID: userIDs[0],
		Category: posts.Funny,
		Text:     "test",
		Created:  t,
		Votes:    map[int64]posts.VoteValue{},
	},
	{
		ID:       postIDs[2],
		Score:    0,
		Views:    789,
		Type:     posts.Text,
		Title:    "Programming title 3",
		AuthorID: userIDs[1],
		Category: posts.Programming,
		Text:     "test",
		Created:  t,
		Votes:    map[int64]posts.VoteValue{},
	},
	{
		ID:       postIDs[3],
		Score:    0,
		Views:    789,
		Type:     posts.Text,
		Title:    "Programming title 3",
		AuthorID: userIDs[1],
		Category: posts.Programming,
		Text:     "test",
		Created:  t,
		Votes:    map[int64]posts.VoteValue{},
	},
	{
		ID:       postIDs[4],
		Score:    0,
		Views:    789,
		Type:     posts.Link,
		Title:    "News title 3",
		AuthorID: userIDs[2],
		Category: posts.News,
		URL:      "https://news.mail.ru/",
		Created:  t,
		Votes:    map[int64]posts.VoteValue{},
	},
}

var upvotePercentages = []uint8{100, 0, 0, 0, 0}

var testUserData = []*user.User{
	{ID: userIDs[0], Username: "test1"},
	{ID: userIDs[1], Username: "test2"},
	{ID: userIDs[2], Username: "test3"},
}

var testCommentData = []*comments.Comment{
	{ID: commentIDs[0], Created: t, AuthorID: userIDs[0], Body: "test comment 1", PostID: postIDs[0]},
	{ID: commentIDs[1], Created: t, AuthorID: userIDs[0], Body: "test comment 2", PostID: postIDs[0]},
	{ID: commentIDs[2], Created: t, AuthorID: userIDs[1], Body: "test comment 3", PostID: postIDs[0]},
	{ID: commentIDs[3], Created: t, AuthorID: userIDs[2], Body: "test comment 4", PostID: postIDs[1]},
	{ID: commentIDs[3], Created: t, AuthorID: userIDs[2], Body: "test comment 5", PostID: postIDs[2]},
}

var newPostID = primitive.NewObjectID()
var newPost = &posts.Post{
	ID:       newPostID,
	Score:    0,
	Views:    0,
	Type:     posts.Link,
	Title:    "News title 3",
	AuthorID: userIDs[0],
	Category: posts.News,
	URL:      "https://news.mail.ru/",
	Created:  t,
	Votes:    map[int64]posts.VoteValue{int64(1): posts.Upvote},
}

var newPostResponse = &PostResponse{
	Score:    newPost.Score,
	Views:    newPost.Views,
	Type:     newPost.Type,
	Title:    newPost.Title,
	Author:   &Author{Username: testUserData[0].Username, ID: testUserData[0].ID},
	Category: newPost.Category,
	URL:      newPost.URL,
	Text:     newPost.Text,
	Votes:    []*posts.Vote{},
	// Votes:    []*posts.Vote{{User: int64(1), Vote: posts.Upvote}},
	Comments: []*CommentResponse{},
	Created:  time.Now(),
	// todo calculate upvote percentage test
	UpvotePercentage: 0,
	ID:               newPostID.Hex(),
}

func prepareTestData(ctrl *gomock.Controller) *PostHandler {
	postsRepoMock := NewMockPostsRepo(ctrl)
	commentsRepoMock := NewMockCommentsRepo(ctrl)
	usersRepoMock := NewMockUsersRepo(ctrl)
	// ctx := context.Background()
	h := &PostHandler{
		Sm:           session.NewMockSessionManager(ctrl),
		PostsRepo:    postsRepoMock,
		UsersRepo:    usersRepoMock,
		CommentsRepo: commentsRepoMock,
		Logger:       zap.NewNop().Sugar(),
	}

	// GetAll result
	postsRepoMock.EXPECT().GetAll(gomock.Any()).Return(testPostData, nil)

	// GetByID result
	for i := 0; i < len(postIDs); i++ {
		postsRepoMock.EXPECT().GetByID(gomock.Any(), postIDs[i]).Return(testPostData[i], nil)
	}

	// GetByAuthorID result

	postsRepoMock.EXPECT().GetByAuthorID(gomock.Any(), userIDs[0]).Return([]*posts.Post{testPostData[0], testPostData[1]}, nil)
	usersRepoMock.EXPECT().GetByUsername(testUserData[0].Username).Return(testUserData[0], nil)
	// GetByCategory result
	// categories := []posts.PostCategory{posts.Fashion, posts.Funny, posts.Programming, posts.News, posts.Music}

	postsRepoMock.EXPECT().GetByCategory(gomock.Any(), posts.Programming).Return([]*posts.Post{testPostData[2], testPostData[3]}, nil)

	for i := 0; i < len(postIDs); i++ {
		postsRepoMock.EXPECT().ParseID(postIDs[i].Hex()).Return(postIDs[i], nil)
	}

	for i := 0; i < len(postIDs); i++ {
		commentsByPostID := func(postID interface{}) []*comments.Comment {
			res := make([]*comments.Comment, 0)
			for _, c := range testCommentData {
				if c.PostID == postID {
					res = append(res, c)
				}
			}
			return res
		}(postIDs[i])
		commentsRepoMock.EXPECT().GetByPostID(gomock.Any(), postIDs[i]).Return(commentsByPostID, nil).AnyTimes()
	}

	for i := 0; i < len(userIDs); i++ {
		usersRepoMock.EXPECT().GetByID(userIDs[i]).Return(testUserData[i], nil).AnyTimes()
	}

	postsRepoMock.EXPECT().Add(gomock.Any(), gomock.AssignableToTypeOf(newPost)).Return(newPostID.Hex(), nil)
	postsRepoMock.EXPECT().Delete(gomock.Any(), newPostID).Return(true, nil)
	postsRepoMock.EXPECT().ParseID(newPostID.Hex()).Return(newPostID, nil)

	postsRepoMock.EXPECT().Upvote(gomock.Any(), postIDs[0], userIDs[0]).
		Return(testPostData[0], nil)
	postsRepoMock.EXPECT().DownVote(gomock.Any(), postIDs[0], userIDs[0]).
		Return(testPostData[0], nil)
	postsRepoMock.EXPECT().Unvote(gomock.Any(), postIDs[0], userIDs[0]).
		Return(testPostData[0], nil)

	return h
}

func getExpectedResult(data []*posts.Post, filter func(*posts.Post) bool) []*PostResponse {
	getVotes := func(votesDB map[int64]posts.VoteValue) []*posts.Vote {
		res := make([]*posts.Vote, 0, len(votesDB))
		for userID, vote := range votesDB {
			res = append(res, &posts.Vote{User: userID, Vote: vote})
		}

		return res
	}

	getAuthor := func(authorID int64) *Author {
		var res *Author
		for _, u := range testUserData {
			if u.ID == authorID {
				return &Author{ID: u.ID, Username: u.Username}
			}
		}

		return res
	}

	getComments := func(commentDB []*comments.Comment, postID interface{}) []*CommentResponse {
		res := make([]*CommentResponse, 0, len(commentDB))
		for _, c := range commentDB {
			if c.PostID == postID {
				res = append(res, &CommentResponse{Created: c.Created, Author: getAuthor(c.AuthorID), Body: c.Body, ID: c.ID.(primitive.ObjectID).Hex()})
			}
		}

		return res
	}

	resp := make([]*PostResponse, 0, len(data))
	for i, d := range data {
		if !filter(d) {
			continue
		}
		r := &PostResponse{Score: d.Score, Views: d.Views, Type: d.Type, Title: d.Title, Category: d.Category, URL: d.URL, Text: d.Text,
			Votes: getVotes(d.Votes), Author: getAuthor(d.AuthorID), Comments: getComments(testCommentData, d.ID), Created: d.Created, UpvotePercentage: upvotePercentages[i],
			ID: d.ID.(primitive.ObjectID).Hex()}
		resp = append(resp, r)
	}

	return resp
}

type getTestCase struct {
	name     string
	handler  func(*PostHandler, http.ResponseWriter, *http.Request)
	method   string
	status   int
	vars     map[string]string
	needAuth bool
	body     map[string]string

	expected       []*PostResponse
	expectedOne    *PostResponse
	expectedCustom map[string]string
}

var getTestCases = []getTestCase{
	{
		name:   "GetAll",
		status: http.StatusOK,
		handler: func(ph *PostHandler, rw http.ResponseWriter, r *http.Request) {
			ph.GetAll(rw, r)
		},
		expected: getExpectedResult(testPostData, func(*posts.Post) bool {
			return true
		}),
		method: http.MethodGet,
		vars:   nil,
	},
	{
		name:   "GetByID",
		status: http.StatusOK,
		handler: func(ph *PostHandler, rw http.ResponseWriter, r *http.Request) {
			ph.GetByID(rw, r)
		},
		expectedOne: getExpectedResult(testPostData, func(p *posts.Post) bool {
			return p.ID == postIDs[0]
		})[0],
		method: http.MethodGet,
		vars: map[string]string{
			"id": postIDs[0].Hex(),
		},
	},
	{
		name:   "GetPostsByCategory",
		status: http.StatusOK,
		handler: func(ph *PostHandler, rw http.ResponseWriter, r *http.Request) {
			ph.GetPostsByCategory(rw, r)
		},
		expected: getExpectedResult(testPostData, func(p *posts.Post) bool {
			return p.Category == posts.Programming
		}),
		method: http.MethodGet,
		vars: map[string]string{
			"category": posts.Programming,
		},
	},
	{
		name:   "GetByAuthorID",
		status: http.StatusOK,
		handler: func(ph *PostHandler, rw http.ResponseWriter, r *http.Request) {
			ph.GetByUser(rw, r)
		},
		expected: getExpectedResult(testPostData, func(p *posts.Post) bool {
			return p.AuthorID == userIDs[0]
		}),
		method: http.MethodGet,
		vars: map[string]string{
			"username": testUserData[0].Username,
		},
	},
	{
		name:     "Create",
		needAuth: true,
		status:   http.StatusCreated,
		handler: func(ph *PostHandler, rw http.ResponseWriter, r *http.Request) {
			ph.Create(rw, r)
		},
		expectedOne: newPostResponse,
		body: map[string]string{
			"category": string(newPost.Category),
			"type":     string(newPost.Type),
			"title":    newPost.Title,
			"text":     newPost.Text,
			"url":      newPost.URL,
		},
		method: http.MethodPost,
		vars: map[string]string{
			"username": testUserData[0].Username,
		},
	},
	{
		name:     "Delete",
		needAuth: true,
		status:   http.StatusOK,
		handler: func(ph *PostHandler, rw http.ResponseWriter, r *http.Request) {
			ph.Delete(rw, r)
		},
		expectedCustom: map[string]string{"message": "success"},
		method:         http.MethodDelete,
		vars: map[string]string{
			"id": newPostID.Hex(),
		},
	},
	{
		name:     "Upvote",
		needAuth: true,
		status:   http.StatusOK,
		handler: func(ph *PostHandler, rw http.ResponseWriter, r *http.Request) {
			ph.Upvote(rw, r)
		},
		expectedOne: getExpectedResult(testPostData, func(p *posts.Post) bool {
			return p.ID == postIDs[0]
		})[0],
		method: http.MethodGet,
		vars: map[string]string{
			"post_id": postIDs[0].Hex(),
		},
	},
	{
		name:     "Downvote",
		needAuth: true,
		status:   http.StatusOK,
		handler: func(ph *PostHandler, rw http.ResponseWriter, r *http.Request) {
			ph.Downvote(rw, r)
		},
		expectedOne: getExpectedResult(testPostData, func(p *posts.Post) bool {
			return p.ID == postIDs[0]
		})[0],
		method: http.MethodGet,
		vars: map[string]string{
			"post_id": postIDs[0].Hex(),
		},
	},
	{
		name:     "Unvote",
		needAuth: true,
		status:   http.StatusOK,
		handler: func(ph *PostHandler, rw http.ResponseWriter, r *http.Request) {
			ph.Unvote(rw, r)
		},
		expectedOne: getExpectedResult(testPostData, func(p *posts.Post) bool {
			return p.ID == postIDs[0]
		})[0],
		method: http.MethodGet,
		vars: map[string]string{
			"post_id": postIDs[0].Hex(),
		},
	},
}

func TestPostCases(t *testing.T) {
	for i, tc := range getTestCases {
		ctrl := gomock.NewController(t)
		h := prepareTestData(ctrl)
		w := httptest.NewRecorder()

		var r *http.Request

		if tc.body != nil {
			bodyBytes, _ := json.Marshal(tc.body)
			body := bytes.NewBuffer(bodyBytes)
			r = httptest.NewRequest(tc.method, "/", body)
		} else {
			r = httptest.NewRequest(tc.method, "/", nil)
		}

		if tc.needAuth {
			r = r.WithContext(context.WithValue(r.Context(), session.SessionKey, &session.Session{User: &session.User{ID: testUserData[0].ID, Username: testUserData[0].Username}}))
		}
		if tc.vars != nil {
			r = mux.SetURLVars(r, tc.vars)
		}

		tc.handler(h, w, r)
		if w.Code != tc.status {
			t.Fatalf("test case %d %s wrong response code, expected %v but was %v", i, tc.name, tc.status, w.Code)
		}
		resBytes, err := ioutil.ReadAll(w.Result().Body)
		if err != nil {
			t.Fatalf("unexpected error occured: %v", err.Error())
		}

		if tc.expected != nil {
			var res []*PostResponse
			err := json.Unmarshal(resBytes, &res)
			if err != nil {
				t.Fatalf("test case %d %s can't get expected result, error occured: %v", i, tc.name, err.Error())
			}
			for i := 0; i < len(res); i++ {
				PostsTestEquals(t, res[i], tc.expected[i])
			}
		}
		if tc.expectedOne != nil {
			var res *PostResponse
			err := json.Unmarshal(resBytes, &res)
			if err != nil {
				t.Fatalf("can't get expected result, error occured: %v", err.Error())
			}
			PostsTestEquals(t, res, tc.expectedOne)
		}
		if tc.expectedCustom != nil {
			res := map[string]string{}
			err := json.Unmarshal(resBytes, &res)
			if err != nil {
				t.Fatalf("can't get expected result, error occured: %v", err.Error())
			}

			if !reflect.DeepEqual(tc.expectedCustom, res) {
				t.Errorf("test fail, votes not equal. expected: %v, but was: %v", tc.expectedCustom, res)
			}
		}
	}
}

func PostsTestEquals(t *testing.T, p1 *PostResponse, p2 *PostResponse) {
	if !func() bool {
		m1 := make(map[int64]posts.VoteValue)
		for _, v := range p1.Votes {
			m1[v.User] = v.Vote
		}
		m2 := make(map[int64]posts.VoteValue)
		for _, v := range p2.Votes {
			m2[v.User] = v.Vote
		}

		return reflect.DeepEqual(m1, m2)
	}() {
		t.Errorf("test fail, votes not equal. expected: %v, but was: %v", p2.Votes, p1.Votes)
	}

	p1.Votes = p2.Votes
	p1.Created = p2.Created
	// fmt.Println(reflect.DeepEqual(p1.Author, p2.Author))
	// fmt.Println(reflect.DeepEqual(p1.Score, p2.Score))
	// fmt.Println(reflect.DeepEqual(p1.Views, p2.Views))
	// fmt.Println(reflect.DeepEqual(p1.Type, p2.Type))
	// fmt.Println(reflect.DeepEqual(p1.Title, p2.Title))
	// fmt.Println(reflect.DeepEqual(p1.Author, p2.Author))
	// fmt.Println(reflect.DeepEqual(p1.Category, p2.Category))
	// fmt.Println(reflect.DeepEqual(p1.URL, p2.URL))
	// fmt.Println(reflect.DeepEqual(p1.Text, p2.Text))
	// fmt.Println(reflect.DeepEqual(p1.Votes, p2.Votes))
	// fmt.Println(reflect.DeepEqual(p1.Comments, p2.Comments))
	// fmt.Println(reflect.DeepEqual(p1.Created, p2.Created))
	// fmt.Println(reflect.DeepEqual(p1.UpvotePercentage, p2.UpvotePercentage))
	// fmt.Println(reflect.DeepEqual(p1.ID, p2.ID))

	if !reflect.DeepEqual(p1, p2) {
		t.Errorf("test fail, expected: %v, but was: %v", p2, p1)
	}
}
