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

	var ok bool
	err := u.db.RawDB().
		QueryRow(query, id).
		Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (u *usersQ) AddAddress(userID int, address data.Address) error {
	exists, err := u.Exists(userID)
	if err != nil {
		return err
	}
	if !exists {
		return data.ErrNotFound
	}

	addressId, err := u.addressesQueryer.Insert(address)
	if err != nil {
		return err
	}

	userAddress := data.UserAddress{
		UserID:    userID,
		AddressID: addressId,
	}

	exists, err = u.usersAddressesQueryer.Exists(userAddress)
	if err != nil {
		return err
	}
	if exists {
		return data.ErrAlreadyExists
	}

	err = u.usersAddressesQueryer.Insert(userAddress)
	if err != nil {
		return err
	}

	return nil
}

func (u *usersQ) GetAddresses(userID int) ([]data.Address, error) {
	// дістати з addresses всі адреси
	exists, err := u.Exists(userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, data.ErrNotFound
	}

	addressID, err := u.usersAddressesQueryer.GetAddresses(userID)
	if err != nil {
		return nil, err
	}

	addresses, err := u.addressesQueryer.SelectAddresses(addressID)
	if err != nil {
		return nil, err
	}

	return addresses, nil
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
