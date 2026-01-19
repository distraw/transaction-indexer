package data

import "errors"

var (
	ErrAlreadyExists = errors.New("proposed row already exists")
	ErrNotFound      = errors.New("requested row does not exist")
)
