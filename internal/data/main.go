package data

type Storage interface {
	Users() UsersQ
	UsersAddresses() UsersAddressesQ
	Addresses() AddressesQ
	Blocks() BlocksQ
	Utxos() UtxosQ

	AddAddress(userID int, address Address) (err error)
	GetAddresses(userID int) ([]Address, error)
	IsTracking(userID int, addr string) (bool, error)

	GetBalance(addr string) (float64, error)

	New() Storage
}
