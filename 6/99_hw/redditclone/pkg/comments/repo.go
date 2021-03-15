package comments

import "sync"

type MemoryCommentsRepo struct {
	lastID uint64
	data   []*Comment
	mu     *sync.Mutex
}

func NewRepo() *MemoryCommentsRepo {
	return &MemoryCommentsRepo{data: make([]*Comment, 0, 10), mu: &sync.Mutex{}}
}

func (repo *MemoryCommentsRepo) GetAll() ([]*Comment, error) {
	return repo.data, nil
}

func (repo *MemoryCommentsRepo) GetByID(id uint64) (*Comment, error) {
	return nil, nil
}

func (repo *MemoryCommentsRepo) GetByPostID(id interface{}) ([]*Comment, error) {
	res := make([]*Comment, 0)
	for _, c := range repo.data {
		if c.PostID == id {
			res = append(res, c)
		}
	}

	return res, nil
}

func (repo *MemoryCommentsRepo) Add(comment *Comment) (uint64, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.lastID++
	id := repo.lastID
	comment.ID = id
	repo.data = append(repo.data, comment)
	return id, nil
}

func (repo *MemoryCommentsRepo) Update(comment *Comment) (bool, error) {
	return false, nil
}

func (repo *MemoryCommentsRepo) Delete(commentID uint64) (bool, error) {
	// repo.mu.Lock()
	// defer repo.mu.Unlock()
	// for i, p := range repo.data {
	// 	if p.AuthorID == postID && p.ID == commentID {
	// 		repo.data[i] = repo.data[len(repo.data)-1]
	// 		repo.data = repo.data[:len(repo.data)-1]
	// 		return true, nil
	// 	}
	// }

	return false, nil
}
