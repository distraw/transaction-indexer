package pg

import (
	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type storage struct {
	db *pgdb.DB

	users          data.UsersQ
	usersAddresses data.UsersAddressesQ
	addresses      data.AddressesQ
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

func (s *storage) AddAddress(userID int, address data.Address) error {
	exists, err := s.users.Exists(userID)
	if err != nil {
		return err
	}
	if !exists {
		return data.ErrNotFound
	}

	addressId, err := s.addresses.Insert(address)
	if err != nil {
		return err
	}

	userAddress := data.UserAddress{
		UserID:    userID,
		AddressID: addressId,
	}

	exists, err = s.usersAddresses.Exists(userAddress)
	if err != nil {
		return err
	}
	if exists {
		return data.ErrAlreadyExists
	}

	err = s.usersAddresses.Insert(userAddress)
	if err != nil {
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

func (s *storage) New() data.Storage {
	return NewStorage(s.db.Clone())
}

func NewStorage(db *pgdb.DB) data.Storage {
	return &storage{
		db: db,

		users:          NewUsersQ(db),
		usersAddresses: NewUsersAddressesQ(db),
		addresses:      NewAddressesQ(db),
	}
}
