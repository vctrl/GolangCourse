package user

import (
	"database/sql"
)

type UserRepoSQL struct {
	db *sql.DB
}

func NewUserRepoSQL(db *sql.DB) *UserRepoSQL {
	return &UserRepoSQL{db: db}
}

func (repo *UserRepoSQL) GetByID(id int64) (*User, error) {
	query := "SELECT `id`, `username`, `password` FROM users WHERE id = ?"
	r := repo.db.QueryRow(query, id)

	u := User{}
	err := r.Scan(&u.ID, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (repo *UserRepoSQL) GetByUsername(username string) (*User, error) {
	query := "SELECT `id`, `username`, `password` FROM users WHERE username = ?"
	r := repo.db.QueryRow(query, username)

	u := User{}
	err := r.Scan(&u.ID, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (repo *UserRepoSQL) Add(user *User) (int64, error) {
	query := "INSERT INTO users (`username`, `password`) VALUES (?, ?)"
	r, err := repo.db.Exec(query, user.Username, user.Password)

	if err != nil {
		return 0, err
	}

	lastID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastID, nil
}
