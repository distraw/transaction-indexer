package pg

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	addressesTable = "addresses"

	addressesAddr = "addr"
)

type addressesQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
	inserter squirrel.InsertBuilder
}

func (a *addressesQ) New() data.AddressesQ {
	return NewAddressesQ(a.db.Clone())
}

func (a *addressesQ) Insert(address data.Address) (int, error) {
	query := a.inserter.SetMap(map[string]interface{}{
		addressesAddr: address.Addr,
	}).
		Suffix(fmt.Sprintf("ON CONFLICT (%s) DO NOTHING", addressesAddr)).
		Suffix("RETURNING id")

	var id int
	err := a.db.Get(&id, query)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (a *addressesQ) Get(addr string) (*data.Address, error) {
	query := a.selector.Where(squirrel.Eq{addressesAddr: addr})

	var address data.Address
	err := a.db.Get(&address, query)
	if err != nil {
		return nil, err
	}

	return &address, nil
}

func (a *addressesQ) Exists(addr string) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1)",
		usersTable,
		usersUsername,
	)

	var ok bool
	err := a.db.RawDB().
		QueryRow(query, addr).
		Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func NewAddressesQ(db *pgdb.DB) data.AddressesQ {
	return &addressesQ{
		db:       db.Clone(),
		selector: squirrel.Select("*").From(addressesTable),
		inserter: squirrel.Insert(addressesTable),
	}
}
