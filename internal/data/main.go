package data

type Storage interface {
	Users() UsersQ
	UsersAddresses() UsersAddressesQ
	Addresses() AddressesQ

	AddAddress(userID int, address Address) (err error)
	GetAddresses(userID int) ([]Address, error)

	New() Storage
}
