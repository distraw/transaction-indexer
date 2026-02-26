package pg

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/core/bitcoin"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type storage struct {
	db *pgdb.DB

	users          data.UsersQ
	usersAddresses data.UsersAddressesQ
	addresses      data.AddressesQ
	transactions   data.TransactionsQ
	outs           data.OutsQ
	ins            data.InsQ

	blocks data.BlocksQ
}

func (s *storage) Users() data.UsersQ {
	return s.users
}

func (s *storage) UsersAddresses() data.UsersAddressesQ {
	return s.usersAddresses
}

func (s *storage) Addresses() data.AddressesQ {
	return s.addresses
}

func (s *storage) Blocks() data.BlocksQ {
	return s.blocks
}

func (s *storage) Transactions() data.TransactionsQ {
	return s.transactions
}

func (s *storage) Outs() data.OutsQ {
	return s.outs
}

func (s *storage) Ins() data.InsQ {
	return s.ins
}

func (s *storage) AddAddress(userID int, address data.Address) error {
	addressId, err := s.addresses.Insert(address)
	if err != nil && !errors.Is(err, data.ErrAlreadyExists) {
		return errors.Wrap(err, "failed to insert address into db")
	}

	userAddress := data.UserAddress{
		UserID:    userID,
		AddressID: *addressId,
	}

	err = s.usersAddresses.Insert(userAddress)
	if err != nil {
		if errors.Is(err, data.ErrAlreadyExists) {
			return errors.Wrap(err, "proposed row already exists")
		}

		return errors.Wrap(err, "failed to insert new relation user_address into db")
	}

	return nil
}

func (s *storage) GetAddresses(userID int) ([]data.Address, error) {
	exists, err := s.users.Exists(userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if user with given ID exists")
	}
	if !exists {
		return nil, data.ErrNotFound
	}

	addressesID, err := s.usersAddresses.GetAddresses(userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all relations user_address from db for given userID")
	}

	addresses, err := s.addresses.SelectAddresses(addressesID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select addresses by theirs ids")
	}

	return addresses, nil
}

func (s *storage) IsTracking(userID int, address string) (bool, error) {
	spk, err := bitcoin.ToScriptPubKey(address)
	if err != nil {
		return false, errors.Wrap(err, "failed to convert address to scriptpubkey")
	}

	dbAddress, err := s.addresses.GetByScriptPubKey(spk)
	if errors.Is(err, data.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "failed to get address by scriptpubkey")
	}

	exists, err := s.usersAddresses.Exists(data.UserAddress{
		UserID:    userID,
		AddressID: dbAddress.ID,
	})
	if err != nil {
		return false, errors.Wrap(err, "failed to check if relation {user.ID;address.ID} exists in db")
	}

	return exists, nil
}

func (s *storage) GetBalance(addr string) (*float64, error) {
	query := fmt.Sprintf(`SELECT COALESCE(SUM(%s.%s), 0) AS balance
	FROM %s
	JOIN %s ON %s.%s = %s.%s
	WHERE %s.%s = $1 AND %s IS NULL`,
		outsTable, outsValue,
		outsTable,
		addressesTable, addressesTable, addressesAddr, outsTable, outsAddress,
		addressesTable, addressesAddr, outsSpentInBlockHeight,
	)

	var balance float64
	err := s.db.GetRaw(&balance, query, addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get balance for given address")
	}

	return &balance, nil
}

func Union(sql1 squirrel.SelectBuilder, sql2 squirrel.SelectBuilder) (rawQuery string, arguments []interface{}, e error) {
	receivedSQL, receivedArgs, err := sql1.ToSql()
	if err != nil {
		return "", nil, err
	}

	spentSQL, spentArgs, err := sql2.ToSql()
	if err != nil {
		return "", nil, err
	}

	query := fmt.Sprintf(
		"%s UNION %s",
		receivedSQL,
		spentSQL,
	)

	args := append(receivedArgs, spentArgs...)

	return query, args, nil
}

func (s *storage) GetTxs(addr string) ([]data.Transaction, error) {
	received := squirrel.
		Select(fmt.Sprintf("%s.%s", transactionsTable, transactionsTxid), transactionsLocktime, transactionsTimestamp).
		From(addressesTable).
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s", outsTable, addressesTable, addressesAddr, outsTable, outsAddress)).
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s", transactionsTable, outsTable, outsTransactionID, transactionsTable, "id")).
		Where(squirrel.Eq{addressesAddr: addr})

	spent := squirrel.
		Select(fmt.Sprintf("%s.%s", transactionsTable, transactionsTxid), transactionsLocktime, transactionsTimestamp).
		From(addressesTable).
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s", insTable, addressesTable, addressesAddr, insTable, insFromAddress)).
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s", transactionsTable, insTable, insTransactionID, transactionsTable, "id")).
		Where(squirrel.Eq{addressesAddr: addr})

	query, args, err := Union(received, spent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to union sql queries")
	}

	var txs []data.Transaction
	err = s.db.SelectRaw(&txs, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute sql query")
	}

	return txs, nil
}

func (s *storage) GetUtxos(addr string) ([]data.Out, error) {
	query := squirrel.Select(outsTxid, outsVout, outsValue, outsSpentInBlockHeight, outsAddress, outsTransactionID).
		From(outsTable).
		JoinClause(
			fmt.Sprintf("JOIN %s ON %s.%s = %s.%s",
				addressesTable,
				addressesTable, addressesAddr, outsTable, outsAddress,
			),
		).
		Where(
			squirrel.Eq{
				fmt.Sprintf("%s.%s", addressesTable, addressesAddr): addr,
				outsSpentInBlockHeight:                              nil,
			},
		)

	var outs []data.Out
	err := s.db.Select(&outs, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from utxos table")
	}

	return outs, nil
}

func (s *storage) GetOutputsInTransaction(txid string) ([]data.Out, error) {
	query := squirrel.Select(fmt.Sprintf("%s.%s", outsTable, outsTxid), outsVout, outsValue, outsSpentInBlockHeight, outsAddress, outsTransactionID).
		From(outsTable).
		JoinClause(
			fmt.Sprintf("JOIN %s ON %s.%s = %s.%s",
				transactionsTable,
				transactionsTable, transactionsTxid, outsTable, outsTxid,
			),
		).
		Where(
			squirrel.Eq{fmt.Sprintf("%s.%s", transactionsTable, transactionsTxid): txid},
		)

	var outs []data.Out
	err := s.db.Select(&outs, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from outs table")
	}

	return outs, nil
}

func (s *storage) GetInputsInTransaction(txid string) ([]data.In, error) {
	query := squirrel.Select(insFromAddress, insValue, insTransactionID, insVin).
		From(insTable).
		JoinClause(
			fmt.Sprintf("JOIN %s ON %s.%s = %s.%s",
				transactionsTable,
				transactionsTable, "id", insTable, insTransactionID,
			),
		).
		Where(
			squirrel.Eq{fmt.Sprintf("%s.%s", transactionsTable, transactionsTxid): txid},
		)

	var ins []data.In
	err := s.db.Select(&ins, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from ins table")
	}

	return ins, nil
}

func (s *storage) New() data.Storage {
	return NewStorage(s.db.Clone())
}

func NewStorage(db *pgdb.DB) data.Storage {
	return &storage{
		db: db,

		users:          NewUsersQ(db),
		usersAddresses: NewUsersAddressesQ(db),
		transactions:   NewTransactionsQ(db),
		addresses:      NewAddressesQ(db),
		outs:           NewOutsQ(db),
		ins:            NewInsQ(db),

		blocks: NewBlocksQ(db),
	}
}
