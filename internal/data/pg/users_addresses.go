package pg

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	usersAddressesTable = "users_addresses"

	usersAddressesUserID    = "user_id"
	usersAddressesAddressID = "address_id"
)

type usersAddressesQ struct {
	db       *pgdb.DB
	inserter squirrel.InsertBuilder
}

func (u *usersAddressesQ) New() data.UsersAddressesQ {
	return NewUsersAddressesQ(u.db.Clone())
}

func (u *usersAddressesQ) Insert(userAddress data.UserAddress) error {
	exists, err := u.Exists(userAddress)
	if err != nil {
		return errors.Wrap(err, "failed to check user_address existence before inserting")
	}
	if exists {
		return data.ErrAlreadyExists
	}

	query := u.inserter.
		SetMap(map[string]interface{}{
			usersAddressesUserID:    userAddress.UserID,
			usersAddressesAddressID: userAddress.AddressID,
		}).Suffix("ON CONFLICT DO NOTHING")

	err = u.db.Exec(query)
	if err != nil {
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			return data.ErrAlreadyExists
		}
		return errors.Wrap(err, "failed to exec raw db query")
	}

	return nil
}

func (u *usersAddressesQ) Exists(userAddress data.UserAddress) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1 AND %s=$2)",
		usersAddressesTable,
		usersAddressesUserID,
		usersAddressesAddressID,
	)

	var ok bool
	err := u.db.RawDB().
		QueryRow(query, userAddress.UserID, userAddress.AddressID).
		Scan(&ok)
	if err != nil {
		return false, errors.Wrap(err, "failed to scan results of raw db query")
	}

	return ok, nil
}

func (u *usersAddressesQ) GetAddresses(userID int) ([]int, error) {
	query := squirrel.
		Select(usersAddressesAddressID).
		From(usersAddressesTable).
		Where(squirrel.Eq{
			usersAddressesUserID: userID,
		})

	var addressID []int
	err := u.db.Select(&addressID, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from db")
	}

	return addressID, nil
}

func NewUsersAddressesQ(db *pgdb.DB) data.UsersAddressesQ {
	return &usersAddressesQ{
		db:       db.Clone(),
		inserter: squirrel.Insert(usersAddressesTable),
	}
}
