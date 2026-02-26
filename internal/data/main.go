package data

type Storage interface {
	Users() UsersQ
	UsersAddresses() UsersAddressesQ
	Addresses() AddressesQ
	Transactions() TransactionsQ
	Blocks() BlocksQ
	Outputs() OutputsQ
	Inputs() InputsQ

	AddAddress(userID int, address Address) (err error)
	GetAddresses(userID int) ([]Address, error)
	IsTracking(userID int, addr string) (bool, error)

	GetBalance(addr string) (*float64, error)
	GetTxs(addr string) ([]Transaction, error)
	GetUtxos(addr string) ([]Output, error)

	// GetBlockOnDepth returns block on given depth in storage
	//
	// If local chain is shorter than depth, first block in chain would be returned
	GetBlockOnDepth(depth int) (*Block, error)

	GetOutputsInTransaction(txid string) ([]Output, error)
	GetInputsInTransaction(txid string) ([]Input, error)

	DeleteBlocksAfter(afterHash string) error

	New() Storage
}
