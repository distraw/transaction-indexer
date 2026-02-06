package pg

import (
	"fmt"
	"strings"

	"github.com/distraw/transaction-indexer/internal/core/bitcoin"
	"github.com/distraw/transaction-indexer/internal/data"
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
	if err != nil {
		return err
	}

	userAddress := data.UserAddress{
		UserID:    userID,
		AddressID: addressId,
	}

	err = s.usersAddresses.Insert(userAddress)
	if err != nil {
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			return data.ErrAlreadyExists
		}

		return err
	}

	return nil
}

func (s *storage) GetAddresses(userID int) ([]data.Address, error) {
	exists, err := s.users.Exists(userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, data.ErrNotFound
	}

	addressID, err := s.usersAddresses.GetAddresses(userID)
	if err != nil {
		return nil, err
	}

	addresses, err := s.addresses.SelectAddresses(addressID)
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

func (s *storage) IsTracking(userID int, addr string) (bool, error) {
	spk, err := bitcoin.ToScriptPubKey(addr)
	if err != nil {
		return false, err
	}

	address, err := s.addresses.GetByScriptPubKey(spk)
	if err != nil {
		return false, err
	}

	return address.ID == userID, nil
}

func (s *storage) GetBalance(addr string) (float64, error) {
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
		return -1, err
	}

	return balance, nil
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
