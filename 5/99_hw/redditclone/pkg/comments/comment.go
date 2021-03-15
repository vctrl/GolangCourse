package comments

import "time"

type Comment struct {
	Created  time.Time
	AuthorID uint64
	Body     string
	ID       uint64
	PostID   uint64
}
