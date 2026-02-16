package pg

import (
	"fmt"
	"strings"

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
	utxos          data.UtxosQ

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

func (s *storage) Utxos() data.UtxosQ {
	return s.utxos
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
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			return data.ErrAlreadyExists
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
		utxosTable, utxosValue,
		utxosTable,
		addressesTable, addressesTable, "id", utxosTable, utxosAddressID,
		addressesTable, addressesAddr, utxosSpentInBlockHeight,
	)

	var balance float64
	err := s.db.GetRaw(&balance, query, addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get balance for given address")
	}

	return &balance, nil
}

func (s *storage) GetUtxos(addr string) ([]data.Utxo, error) {
	query := squirrel.Select(utxosTxid, utxosVout, utxosValue, utxosSpentInBlockHeight, utxosAddressID, utxosBlockID).
		From(utxosTable).
		JoinClause(
			fmt.Sprintf("JOIN %s ON %s.id = %s.%s",
				addressesTable,
				addressesTable, utxosTable, utxosAddressID,
			),
		).
		Where(
			squirrel.Eq{fmt.Sprintf("%s.%s", addressesTable, addressesAddr): addr},
		)

	var utxos []data.Utxo
	err := s.db.Select(&utxos, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from utxos table")
	}

	return utxos, nil
}

func (s *storage) New() data.Storage {
	return NewStorage(s.db.Clone())
}

func NewStorage(db *pgdb.DB) data.Storage {
	return &storage{
		db: db,

		users:          NewUsersQ(db),
		usersAddresses: NewUsersAddressesQ(db),
		addresses:      NewAddressesQ(db),
		utxos:          NewUtxosQ(db),

		blocks: NewBlocksQ(db),
	}
}
