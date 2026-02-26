package pg

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	addressesTable = "addresses"

	addressesAddr         = "addr"
	addressesScriptPubKey = "scriptpubkey"
)

type addressesQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
	inserter squirrel.InsertBuilder
}

func (a *addressesQ) New() data.AddressesQ {
	return NewAddressesQ(a.db.Clone())
}

func (a *addressesQ) Insert(address data.Address) (*int, error) {
	query := a.inserter.SetMap(map[string]interface{}{
		addressesAddr:         address.Addr,
		addressesScriptPubKey: address.ScriptPubKey,
	}).
		Suffix(fmt.Sprintf(
			"ON CONFLICT (%s) DO UPDATE SET %s=EXCLUDED.%s",
			addressesAddr, addressesAddr, addressesAddr)).
		Suffix("RETURNING id")

	var id int
	err := a.db.Get(&id, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute db query")
	}

	return &id, nil
}

func (a *addressesQ) GetAllScriptPubKeys() ([]string, error) {
	query := squirrel.Select(addressesScriptPubKey).From(addressesTable)

	var addresses []string
	err := a.db.Select(&addresses, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute db query")
	}
	if len(addresses) == 0 {
		return nil, data.ErrNotFound
	}

	return addresses, nil
}

func (a *addressesQ) GetByScriptPubKey(scriptPubKey string) (*data.Address, error) {
	query := a.selector.Where(squirrel.Eq{
		addressesScriptPubKey: scriptPubKey,
	})

	var address data.Address
	err := a.db.Get(&address, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute db query")
	}

	return &address, nil
}

func (a *addressesQ) Exists(addr string) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1)",
		addressesTable,
		addressesAddr,
	)

	var ok bool
	err := a.db.RawDB().
		QueryRow(query, addr).
		Scan(&ok)
	if err != nil {
		return false, errors.Wrap(err, "failed to scan raw db query row")
	}

	return ok, nil
}

func (a *addressesQ) ExistsByScriptPubKey(spk string) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1)",
		addressesTable,
		addressesScriptPubKey,
	)

	var ok bool
	err := a.db.RawDB().
		QueryRow(query, spk).
		Scan(&ok)
	if err != nil {
		return false, errors.Wrap(err, "failed to execute raw db query")
	}

	return ok, nil
}

func (a *addressesQ) SelectAddresses(ids []int) ([]data.Address, error) {
	if len(ids) == 0 {
		return []data.Address{}, nil
	}

	query := a.selector.Where(squirrel.Eq{
		"id": ids,
	})

	var addrs []data.Address
	err := a.db.Select(&addrs, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from db")
	}

	return addrs, nil
}

func NewAddressesQ(db *pgdb.DB) data.AddressesQ {
	return &addressesQ{
		db:       db.Clone(),
		selector: squirrel.Select("*").From(addressesTable),
		inserter: squirrel.Insert(addressesTable),
	}
}
