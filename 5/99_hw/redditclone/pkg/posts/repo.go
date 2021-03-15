package posts

import (
	"encoding/json"
	"sync"
)

type MemoryPostsRepo struct {
	mu     *sync.Mutex
	lastID uint64
	data   []*Post
}

func NewRepo() *MemoryPostsRepo {
	return &MemoryPostsRepo{data: make([]*Post, 0, 10), mu: &sync.Mutex{}}
}

func (repo *MemoryPostsRepo) GetAll() ([]*Post, error) {
	return repo.data, nil
}

func (repo *MemoryPostsRepo) GetByCategory(category string) ([]*Post, error) {
	return repo.getPosts(func(p *Post) bool { return p.Category == PostCategory(category) })
}

func (repo *MemoryPostsRepo) GetByAuthorID(authorID uint64) ([]*Post, error) {
	return repo.getPosts(func(p *Post) bool { return p.AuthorID == authorID })
}

func (repo *MemoryPostsRepo) GetById(id uint64) (*Post, error) {
	res, err := repo.getPosts(func(p *Post) bool { return p.ID == id })
	if err != nil {
		return nil, err
	}

	if len(res) > 0 {
		repo.mu.Lock()
		defer repo.mu.Unlock()
		res[0].Views++
		return res[0], nil
	}

	return nil, nil
}

func (repo *MemoryPostsRepo) Add(post *Post) (uint64, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.lastID++
	post.ID = repo.lastID
	repo.data = append(repo.data, post)
	return post.ID, nil
}

func (repo *MemoryPostsRepo) Update(post *Post) (bool, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for i, p := range repo.data {
		if p.ID == post.ID {
			repo.data[i].AuthorID = post.ID
			repo.data[i].Category = post.Category
			repo.data[i].Text = post.Text
			repo.data[i].Title = post.Title
			repo.data[i].Type = post.Type
			repo.data[i].Views = post.Views
			return true, nil
		}
	}

	return false, nil
}

func (repo *MemoryPostsRepo) Delete(id uint64) (bool, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for i, p := range repo.data {
		if p.ID == id {
			repo.data[i] = repo.data[len(repo.data)-1]
			repo.data = repo.data[:len(repo.data)-1]
			return true, nil
		}
	}

	return false, nil
}

func (repo *MemoryPostsRepo) Upvote(postID, userID uint64) (*Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, p := range repo.data {
		if p.ID == postID {
			return vote(p, userID, func(a map[uint64]VoteValue, b uint64) { a[b] = 1 })
		}
	}

	return nil, nil
}

func (repo *MemoryPostsRepo) DownVote(postID, userID uint64) (*Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, p := range repo.data {
		if p.ID == postID {
			return vote(p, userID, func(a map[uint64]VoteValue, b uint64) { a[b] = -1 })
		}
	}

	return nil, nil
}

func (repo *MemoryPostsRepo) Unvote(postID, userID uint64) (*Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, p := range repo.data {
		if p.ID == postID {
			return vote(p, userID, func(a map[uint64]VoteValue, b uint64) { delete(a, b) })
		}
	}

	return nil, nil
}

func (repo *MemoryPostsRepo) getPosts(filter func(*Post) bool) ([]*Post, error) {
	res := make([]*Post, 0, 10)
	for _, p := range repo.data {
		if filter(p) {
			res = append(res, p)
		}
	}

	return res, nil
}

func vote(p *Post, userID uint64, vote func(map[uint64]VoteValue, uint64)) (*Post, error) {
	userVote := make(map[uint64]VoteValue)
	err := json.Unmarshal(p.Votes, &userVote)
	if err != nil {
		return nil, err
	}

	vote(userVote, userID)

	res := &Votes{Votes: make([]*Vote, 0, len(userVote))}

	for u, v := range userVote {
		res.Votes = append(res.Votes, &Vote{User: u, Vote: v})
	}

	newVotes, err := json.Marshal(userVote)
	if err != nil {
		return nil, err
	}

	p.Votes = newVotes
	return p, nil
}
