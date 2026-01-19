package pg

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	usersAddressesTable = "users_addresses"

	usersAddressesUserID    = "user_id"
	usersAddressesAddressID = "address_id"
)

type usersAddressesQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
	inserter squirrel.InsertBuilder
}

func (u *usersAddressesQ) New() data.UsersAddressesQ {
	return NewUsersAddressesQ(u.db.Clone())
}

func (u *usersAddressesQ) Insert(userAddress data.UserAddress) error {
	query := u.inserter.
		SetMap(map[string]interface{}{
			usersAddressesUserID:    userAddress.UserID,
			usersAddressesAddressID: userAddress.AddressID,
		}).Suffix("ON CONFLICT DO NOTHING")

	if err := u.db.Exec(query); err != nil {
		return err
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
		return false, err
	}

	return ok, nil
}

func NewUsersAddressesQ(db *pgdb.DB) data.UsersAddressesQ {
	return &usersAddressesQ{
		db:       db.Clone(),
		selector: squirrel.Select("*").From(usersAddressesTable),
		inserter: squirrel.Insert(usersAddressesTable),
	}
}
