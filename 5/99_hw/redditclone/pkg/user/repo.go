package user

import "sync"

type MemoryUsersRepo struct {
	lastID uint64
	data   []*User
	mu     *sync.Mutex
}

func NewRepo() *MemoryUsersRepo {
	return &MemoryUsersRepo{data: make([]*User, 0, 10), mu: &sync.Mutex{}}
}

func (repo *MemoryUsersRepo) GetAll() ([]*User, error) {
	return repo.data, nil
}

func (repo *MemoryUsersRepo) GetByID(id uint64) (*User, error) {
	for _, u := range repo.data {
		if u.ID == id {
			return u, nil
		}
	}

	return nil, nil
}

func (repo *MemoryUsersRepo) GetByUsername(username string) (*User, error) {
	for _, u := range repo.data {
		if u.Username == username {
			return u, nil
		}
	}

	return nil, nil
}

func (repo *MemoryUsersRepo) Add(user *User) (uint64, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.lastID++
	user.ID = repo.lastID
	repo.data = append(repo.data, user)
	return repo.lastID, nil
}

func (repo *MemoryUsersRepo) Update(user *User) (bool, error) {
	return false, nil
}

func (repo *MemoryUsersRepo) Delete(user *User) (bool, error) {
	return false, nil
}
