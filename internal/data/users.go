package data

import "errors"

var ErrAlreadyExists = errors.New("user with providen username already exists")

type UsersQ interface {
	New() UsersQ
	Insert(user User) (id int64, err error)
	Get(id int64) (user *User, err error)
}

type User struct {
	ID       int64  `structs:"-" db:"id"`
	Username string `structs:"username" db:"username"`
	Password []byte `structs:"password" db:"password"`
}
