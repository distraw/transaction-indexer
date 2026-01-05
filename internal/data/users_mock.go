package data

import "errors"

const (
	MockedID       = 1
	MockedUser     = "mocked_user"
	MockedPassword = "mocked_password"
)

type UsersQMock struct{}

func (u *UsersQMock) New() UsersQ {
	return u
}

func (u *UsersQMock) Insert(user User) (int64, error) {
	return 1, nil
}

func (u *UsersQMock) Get(id int64) (*User, error) {
	if id < 1 {
		return nil, errors.New("id must not be negative and must not equal to zero")
	}

	return &User{
		ID:       1,
		Username: MockedUser,
		Password: []byte(MockedPassword),
	}, nil
}
