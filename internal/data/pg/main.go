package pg

import (
	"database/sql"
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
	outputs        data.OutputsQ
	inputs         data.InputsQ

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

func (s *storage) Outputs() data.OutputsQ {
	return s.outputs
}

func (s *storage) Inputs() data.InputsQ {
	return s.inputs
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
		outputsTable, outputsValue,
		outputsTable,
		addressesTable, addressesTable, addressesAddr, outputsTable, outputsAddress,
		addressesTable, addressesAddr, outputsSpentInTransactionID,
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
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s", outputsTable, addressesTable, addressesAddr, outputsTable, outputsAddress)).
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s", transactionsTable, outputsTable, outputsTransactionID, transactionsTable, "id")).
		Where(squirrel.Eq{addressesAddr: addr})

	spent := squirrel.
		Select(fmt.Sprintf("%s.%s", transactionsTable, transactionsTxid), transactionsLocktime, transactionsTimestamp).
		From(addressesTable).
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s", inputsTable, addressesTable, addressesAddr, inputsTable, inputsFromAddress)).
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s", transactionsTable, inputsTable, inputsTransactionID, transactionsTable, "id")).
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

func (s *storage) GetUtxos(addr string) ([]data.Output, error) {
	query := squirrel.Select(outputsTxid, outputsVout, outputsValue, outputsSpentInTransactionID, outputsAddress, outputsTransactionID).
		From(outputsTable).
		JoinClause(
			fmt.Sprintf("JOIN %s ON %s.%s = %s.%s",
				addressesTable,
				addressesTable, addressesAddr, outputsTable, outputsAddress,
			),
		).
		Where(
			squirrel.Eq{
				fmt.Sprintf("%s.%s", addressesTable, addressesAddr): addr,
				outputsSpentInTransactionID:                         nil,
			},
		)

	var outputs []data.Output
	err := s.db.Select(&outputs, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from utxos table")
	}

	return outputs, nil
}

func (s *storage) GetBlockOnDepth(depth int) (*data.Block, error) {
	current, err := s.blocks.GetTip()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tip of local chain")
	}

	for i := 0; i < depth; i++ {
		if current.PreviousBlockID == nil {
			return current, nil
		}

		current, err = s.blocks.GetByID(*current.PreviousBlockID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get block by id in local chain")
		}
	}

	return current, nil
}

func (s *storage) GetOutputsInTransaction(txid string) ([]data.Output, error) {
	query := squirrel.Select(fmt.Sprintf("%s.%s", outputsTable, outputsTxid), outputsVout, outputsValue, outputsSpentInTransactionID, outputsAddress, outputsTransactionID).
		From(outputsTable).
		JoinClause(
			fmt.Sprintf("JOIN %s ON %s.%s = %s.%s",
				transactionsTable,
				transactionsTable, transactionsTxid, outputsTable, outputsTxid,
			),
		).
		Where(
			squirrel.Eq{fmt.Sprintf("%s.%s", transactionsTable, transactionsTxid): txid},
		)

	var outputs []data.Output
	err := s.db.Select(&outputs, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from outs table")
	}

	return outputs, nil
}

func (s *storage) GetInputsInTransaction(txid string) ([]data.Input, error) {
	query := squirrel.Select(inputsFromAddress, inputsValue, inputsTransactionID, inputsVin).
		From(inputsTable).
		JoinClause(
			fmt.Sprintf("JOIN %s ON %s.%s = %s.%s",
				transactionsTable,
				transactionsTable, "id", inputsTable, inputsTransactionID,
			),
		).
		Where(
			squirrel.Eq{fmt.Sprintf("%s.%s", transactionsTable, transactionsTxid): txid},
		)

	var inputs []data.Input
	err := s.db.Select(&inputs, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from ins table")
	}

	return inputs, nil
}

func (s *storage) New() data.Storage {
	return NewStorage(s.db.Clone())
}

func (s *storage) DeleteBlocksAfter(afterHash string) error {
	getIDQuery := squirrel.
		Select("id").
		From(blocksTable).
		Where(squirrel.Eq{
			blocksHash: afterHash,
		})

	var block data.Block
	err := s.db.Get(&block, getIDQuery)
	if errors.Is(err, sql.ErrNoRows) {
		return data.ErrNotFound
	}

	deleteQuery := squirrel.
		Delete(blocksTable).
		Where(squirrel.Eq{
			blocksPreviousBlockID: block.ID,
		})

	err = s.db.Exec(deleteQuery)
	if err != nil {
		return errors.Wrap(err, "failed to exec db query")
	}

	return nil
}

func NewStorage(db *pgdb.DB) data.Storage {
	return &storage{
		db: db,

		users:          NewUsersQ(db),
		usersAddresses: NewUsersAddressesQ(db),
		transactions:   NewTransactionsQ(db),
		addresses:      NewAddressesQ(db),
		outputs:        NewOutputsQ(db),
		inputs:         NewInsQ(db),

		blocks: NewBlocksQ(db),
	}
}
