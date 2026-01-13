package data

import "errors"

var ErrAlreadyExists = errors.New("user with provided username already exists")
var ErrUserNotFound = errors.New("user with given username and password does not exist")

type UsersQ interface {
	New() UsersQ
	Insert(user User) error
	Get(username string) (*User, error)
	Exists(username string) (ok bool)
}

type User struct {
	Username string `structs:"username" db:"username"`
	Password []byte `structs:"password" db:"password"`
}
