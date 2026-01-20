package data

type AddressesQ interface {
	New() AddressesQ

	Insert(address Address) (id int, err error)
	Get(addr string) (*Address, error)
	Exists(addr string) (ok bool, err error)

	SelectAddresses(ids []int) ([]Address, error)
}

type Address struct {
	ID   int    `structs:"id" db:"id"`
	Addr string `structs:"addr" db:"addr"`
}
