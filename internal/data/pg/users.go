package pg

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	usersTable = "users"

	usersUsername = "username"
	usersPassword = "password"
)

type usersQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
	inserter squirrel.InsertBuilder
}

func (u *usersQ) New() data.UsersQ {
	return NewUsersQ(u.db.Clone())
}

func (u *usersQ) Insert(user data.User) (int, error) {
	stmt := u.inserter.
		SetMap(map[string]interface{}{
			usersUsername: user.Username,
			usersPassword: user.Password,
		}).Suffix("RETURNING id")

	var id int
	if err := u.db.Get(&id, stmt); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			err = data.ErrAlreadyExists
		}

		return -1, err
	}

	return id, nil
}

func (u *usersQ) Get(username string) (*data.User, error) {
	query := u.selector.Where(squirrel.Eq{usersUsername: username})

	var user data.User
	err := u.db.Get(&user, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *usersQ) Exists(username string) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1)",
		usersTable,
		usersUsername,
	)

	var ok bool
	err := u.db.RawDB().
		QueryRow(query, username).
		Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func NewUsersQ(db *pgdb.DB) data.UsersQ {
	return &usersQ{
		db:       db.Clone(),
		selector: squirrel.Select("*").From(usersTable),
		inserter: squirrel.Insert(usersTable),
	}
}
