package data

type Storage interface {
	Users() UsersQ
	UsersAddresses() UsersAddressesQ
	Addresses() AddressesQ
	Transactions() TransactionsQ
	Blocks() BlocksQ
	Outs() OutsQ
	Ins() InsQ

	AddAddress(userID int, address Address) (err error)
	GetAddresses(userID int) ([]Address, error)
	IsTracking(userID int, addr string) (bool, error)

	GetBalance(addr string) (*float64, error)
	GetTxs(addr string) ([]Transaction, error)
	GetUtxos(addr string) ([]Out, error)

	GetOutputsInTransaction(txid string) ([]Out, error)
	GetInputsInTransaction(txid string) ([]In, error)

	New() Storage
}
