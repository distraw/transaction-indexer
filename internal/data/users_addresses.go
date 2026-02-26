package data

type UsersAddressesQ interface {
	New() UsersAddressesQ

	Insert(userAddress UserAddress) error
	Exists(userAddress UserAddress) (ok bool, err error)

	GetAddresses(userID int) (addressID []int, err error)
}

type UserAddress struct {
	UserID    int `structs:"user_id" db:"user_id"`
	AddressID int `structs:"address_id" db:"address_id"`
}
