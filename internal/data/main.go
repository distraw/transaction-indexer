package data

type Storage interface {
	Users() UsersQ
	UsersAddresses() UsersAddressesQ
	Addresses() AddressesQ
	Blocks() BlocksQ

	AddAddress(userID int, address Address) (err error)
	GetAddresses(userID int) ([]Address, error)

	New() Storage
}
