package posts

import (
	"time"
)

type PostType string

const (
	Text PostType = "text"
	Link          = ""
)

type PostCategory string

const (
	All         PostCategory = "all"
	Music                    = "music"
	Funny                    = "funny"
	Videos                   = "videos"
	Programming              = "programming"
	News                     = "news"
	Fashion                  = "fashion"
)

type Post struct {
	ID       interface{}         `bson:"_id,omitempty"`
	Score    int                 `bson:"score"`
	Views    uint64              `bson:"views"`
	Type     PostType            `bson:"type"`
	Title    string              `bson:"title"`
	AuthorID int64               `bson:"authorID"`
	Category PostCategory        `bson:"category"`
	Text     string              `bson:"text"`
	URL      string              `bson:"URL"`
	Created  time.Time           `bson:"created"`
	Votes    map[int64]VoteValue `bson:"votes"`
}

type Votes struct {
	Votes []*Vote `json:"votes"`
}

type Vote struct {
	User int64     `json:"user"`
	Vote VoteValue `json:"vote"`
}

type VoteValue int8

const (
	Downvote VoteValue = iota - 1
	Unvote
	Upvote
)
