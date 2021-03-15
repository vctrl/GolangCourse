package user

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type getByFieldTestCase struct {
	getBy func(*UserRepoSQL, interface{}) (*User, error)
	param interface{}
}

var id = int64(25)
var u = &User{ID: id, Username: "vectoreal", Password: []byte("secretPASSW0rd")}

var cases = []getByFieldTestCase{
	{
		getBy: func(r *UserRepoSQL, id interface{}) (*User, error) {
			return r.GetByID(id.(int64))
		},
		param: u.ID,
	},
	{
		getBy: func(r *UserRepoSQL, username interface{}) (*User, error) {
			return r.GetByUsername(username.(string))
		},
		param: u.Username,
	},
}

func TestGetByID(t *testing.T) {
	for _, tc := range cases {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		defer db.Close()

		repo := NewUserRepoSQL(db)

		rows := sqlmock.NewRows([]string{"id", "username", "password"}).
			AddRow(id, u.Username, u.Password)

		mock.
			ExpectQuery("SELECT `id`, `username`, `password` FROM users WHERE").
			WithArgs(tc.param).
			WillReturnRows(rows)

		res, err := tc.getBy(repo, tc.param)
		if err != nil {
			t.Fatalf("unexpected error: %v", err.Error())
		}

		if !reflect.DeepEqual(u, res) {
			t.Fatalf("expected %v, but was %v", u, res)
		}

		// error
		mock.
			ExpectQuery("SELECT `id`, `username`, `password` FROM users WHERE").
			WithArgs(tc.param).
			WillReturnError(errors.New("db_error"))

		res, err = tc.getBy(repo, tc.param)

		if res != nil {
			t.Fatalf("unexpected result: %v", res)
		}

		if err == nil {
			t.Fatalf("expected error but was nil")
		}

		// no rows
		mock.
			ExpectQuery("SELECT `id`, `username`, `password` FROM users WHERE").
			WithArgs(tc.param).
			WillReturnError(sql.ErrNoRows)

		res, err = tc.getBy(repo, tc.param)

		if res != nil || err != nil {
			t.Fatalf("wrong result, expected both nil but was %v, %v", res, err)
		}
	}
}

func TestAdd(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepoSQL(db)
	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(u.Username, u.Password).
		WillReturnResult(sqlmock.NewResult(u.ID, int64(1)))

	id, err := repo.Add(u)
	if err != nil {
		t.Fatalf("unexpected error while adding user: %v", err.Error())
	}
	if id != u.ID {
		t.Fatalf("expected %v but was %v", u.ID, id)
	}

	// error
	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(u.Username, u.Password).
		WillReturnError(errors.New("db_error"))

	_, err = repo.Add(u)

	if err == nil {
		t.Fatalf("expected error but was nil")
	}
	if err.Error() != "db_error" {
		t.Fatalf("unexpected error: %v", err.Error())
	}

	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(u.Username, u.Password).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("db_error")))

	_, err = repo.Add(u)
	if err == nil {
		t.Fatalf("expected error but was nil")
	}
	if err.Error() != "db_error" {
		t.Fatalf("unexpected error: %v", err.Error())
	}
}
