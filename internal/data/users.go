package data

import "errors"

var ErrAlreadyExists = errors.New("user with providen username already exists")
var ErrUserNotFound = errors.New("user with given username and password does not exist")

type UsersQ interface {
	New() UsersQ
	Insert(user User) (id int64, err error)
	Get(username string) (*User, error)
}

type User struct {
	ID       int64  `structs:"-" db:"id"`
	Username string `structs:"username" db:"username"`
	Password []byte `structs:"password" db:"password"`
}
