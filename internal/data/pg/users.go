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

	addressesQueryer      data.AddressesQ
	usersAddressesQueryer data.UsersAddressesQ
}

func (u *usersQ) New() data.UsersQ {
	return NewUsersQ(u.db.Clone())
}

func (u *usersQ) Insert(user data.User) (int, error) {
	query := u.inserter.
		SetMap(map[string]interface{}{
			usersUsername: user.Username,
			usersPassword: user.Password,
		}).Suffix("RETURNING id")

	var id int
	if err := u.db.Get(&id, query); err != nil {
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			return 0, data.ErrAlreadyExists
		}

		return 0, err
	}

	return id, nil
}

func (u *usersQ) Get(username string) (*data.User, error) {
	query := u.selector.Where(squirrel.Eq{usersUsername: username})

	var user data.User
	err := u.db.Get(&user, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *usersQ) Exists(id int) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1)",
		usersTable,
		"id",
	)

	var exists bool
	err := u.db.RawDB().
		QueryRow(query, id).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (u *usersQ) ExistsByUsername(username string) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1)",
		usersTable,
		usersUsername,
	)

	var exists bool
	err := u.db.RawDB().
		QueryRow(query, username).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func NewUsersQ(db *pgdb.DB) data.UsersQ {
	return &usersQ{
		db:       db.Clone(),
		selector: squirrel.Select("*").From(usersTable),
		inserter: squirrel.Insert(usersTable),

		addressesQueryer:      NewAddressesQ(db.Clone()),
		usersAddressesQueryer: NewUsersAddressesQ(db.Clone()),
	}
}
