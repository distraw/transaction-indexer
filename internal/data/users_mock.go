package data

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

func (u *UsersQMock) Get(username string) (*User, error) {
	if username == MockedUser {
		return &User{
			ID:       MockedID,
			Username: MockedUser,
			Password: []byte(MockedPassword),
		}, nil
	}

	return nil, ErrUserNotFound
}
