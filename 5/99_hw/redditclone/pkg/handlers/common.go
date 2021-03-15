package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"redditclone/pkg/comments"
	"redditclone/pkg/posts"
	"redditclone/pkg/user"
	"strconv"

	"github.com/gorilla/mux"
)

type Response struct {
	Message string `json:"message"`
}

type CustomError struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Value    string `json:"value"`
	Msg      string `json:"msg"`
}

type ErrorsResponse struct {
	Errors []*CustomError `json:"errors"`
}

func WriteResponse(w http.ResponseWriter, msg string, status int) {
	resp := &Response{Message: msg}
	res, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(status)
		return
	}

	w.WriteHeader(status)
	w.Write(res)
}

func writeErrorsResponse(w http.ResponseWriter, errors []*CustomError, status int) {
	errorsJSON, err := json.Marshal(&ErrorsResponse{Errors: errors})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(status)
	w.Write(errorsJSON)
}

func getPostData(p *posts.Post, ur *user.MemoryUsersRepo, cr *comments.MemoryCommentsRepo) (*PostResponse, error) {
	author, err := ur.GetByID(p.AuthorID)
	if err != nil {
		return nil, err
	}
	comments, err := cr.GetByPostID(p.ID)
	if err != nil {
		return nil, err
	}

	return MapToPostResponse(p, author.ID, author.Username, comments, ur)
}

func MapToPostResponse(post *posts.Post, authorID uint64, authorName string, comments []*comments.Comment, usersRepo *user.MemoryUsersRepo) (*PostResponse, error) {
	var userVote map[uint64]posts.VoteValue
	err := json.Unmarshal(post.Votes, &userVote)

	if err != nil {
		return nil, err
	}

	score := 0
	votes := make([]*posts.Vote, 0, len(userVote))
	for u, v := range userVote {
		vote := &posts.Vote{User: u, Vote: v}
		votes = append(votes, vote)
		score += int(v)
	}

	commentsResp, err := mapToCommentsResponse(comments, usersRepo)
	if err != nil {
		return nil, err
	}

	resp := &PostResponse{
		ID:               post.ID,
		Score:            score,
		Views:            post.Views,
		Type:             post.Type,
		Title:            post.Title,
		Author:           &Author{Username: authorName, ID: authorID},
		Category:         post.Category,
		Votes:            votes,
		Comments:         commentsResp,
		Created:          post.Created,
		UpvotePercentage: calculateUpvotePercentage(votes),
	}

	if resp.Type == posts.PostType("text") {
		resp.Text = post.Text
	} else {
		resp.URL = post.URL
	}

	return resp, nil
}

func ParseUintParam(r *http.Request, name string) (uint64, error) {
	vars := mux.Vars(r)
	varStr := vars[name]
	val, err := strconv.ParseUint(varStr, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("wrong id value: %v", varStr)
	}

	return val, nil
}
