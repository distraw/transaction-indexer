package data

import "maps"

type UsersQMock struct {
	Data map[string]User
}

func (u *UsersQMock) New() UsersQ {
	return u
}

func (u *UsersQMock) Insert(user User) error {
	if _, ok := u.Data[user.Username]; ok {
		return ErrAlreadyExists
	}

	u.Data[user.Username] = user
	return nil
}

func (u *UsersQMock) Get(username string) (*User, error) {
	if user, ok := u.Data[username]; ok {
		return &User{
			Username: user.Username,
			Password: user.Password,
		}, nil
	}

	return nil, ErrUserNotFound
}

func (u *UsersQMock) Exists(username string) (bool, error) {
	_, ok := u.Data[username]
	return ok, nil
}

// WithUser returns deep copy of UsersQMock containing providen user
// for testing purposes
func (u UsersQMock) WithUser(user User) UsersQMock {
	dataCopy := maps.Clone(u.Data)
	dataCopy[user.Username] = user
	return UsersQMock{Data: dataCopy}
}

func NewUsersQMock() UsersQMock {
	return UsersQMock{
		Data: make(map[string]User, 0),
	}
}
