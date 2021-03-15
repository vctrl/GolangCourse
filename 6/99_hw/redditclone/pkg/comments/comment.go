package comments

import "time"

type Comment struct {
	Created  time.Time   `bson:"created"`
	AuthorID int64       `bson:"authorID"`
	Body     string      `bson:"body"`
	ID       interface{} `bson:"_id,omitempty"`
	PostID   interface{} `bson:"postID"`
}
