package posts

import "time"

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
	ID       uint64
	Score    int
	Views    uint64
	Type     PostType
	Title    string
	AuthorID uint64
	Category PostCategory
	Text     string
	URL      string
	Created  time.Time
	Votes    []byte
}

type Votes struct {
	Votes []*Vote `json:"votes"`
}

type Vote struct {
	User uint64    `json:"user"`
	Vote VoteValue `json:"vote"`
}

type VoteValue int8

const (
	Downvote VoteValue = iota - 1
	Unvote
	Upvote
)
