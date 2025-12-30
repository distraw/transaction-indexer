package pg

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	usersTable = "users"

	usersID       = "id"
	usersUsername = "username"
	usersPassword = "password"
)

type usersQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
}

func (u *usersQ) New() data.UsersQ {
	return NewUsersQ(u.db.Clone())
}

func (u *usersQ) Insert(user data.User) (int64, error) {
	stmt := squirrel.
		Insert(usersTable).
		SetMap(map[string]interface{}{
			usersUsername: user.Username,
			usersPassword: user.Password,
		}).
		Suffix("RETURNING id")

	var id int64
	if err := u.db.Get(&id, stmt); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			err = data.ErrAlreadyExists
		}

		return id, err
	}

	return id, nil
}

func (u *usersQ) Get(id int64) (*data.User, error) {
	var user data.User
	err := u.db.Get(&user, u.selector.Where(squirrel.Eq{
		usersID: id,
	}))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &user, err
}

func NewUsersQ(db *pgdb.DB) data.UsersQ {
	return &usersQ{
		db:       db.Clone(),
		selector: squirrel.Select("*").From(usersTable),
	}
}
