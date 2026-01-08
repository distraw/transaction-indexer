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

// Insert inserts new user with given credentials into DB.
//
// Does not return ID because, even though it exists as a primary key in DB,
// username should be used as a key instead
//
// For proper functioning chosen DB MUST ensure uniqueness of the username
func (u *usersQ) Insert(user data.User) error {
	stmt := squirrel.
		Insert(usersTable).
		SetMap(map[string]interface{}{
			usersUsername: user.Username,
			usersPassword: user.Password,
		})

	if err := u.db.Exec(stmt); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			err = data.ErrAlreadyExists
		}

		return err
	}

	return nil
}

func (u *usersQ) Get(username string) (*data.User, error) {
	var user data.User

	query := squirrel.
		Select(usersUsername, usersPassword).
		From(usersTable).
		Where(squirrel.Eq{
			usersUsername: username,
		})

	err := u.db.Get(&user, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrUserNotFound
	}

	return &user, err
}

func NewUsersQ(db *pgdb.DB) data.UsersQ {
	return &usersQ{
		db: db.Clone(),

		// Although ID (unique primary key) exists, it is not used.
		// Username, which uniqueness MUST be forced by DB, is used instead
		selector: squirrel.Select(usersUsername, usersPassword).From(usersTable),
	}
}
