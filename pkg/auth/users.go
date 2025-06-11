package auth

import (
	"github.com/google/uuid"
	"github.com/null-create/logger"
)

type UserKey string // context key for the parsed claims

type User struct {
	name  string
	token string
}

func (u *User) Name() string  { return u.name }
func (u *User) Token() string { return u.token }
func (u *User) AddToken(tok string) {
	if u.token == "" {
		u.token = tok
	}
}

type UsersManager struct {
	log   *logger.Logger
	users map[string]*User
}

func NewUsersManager() UsersManager {
	return UsersManager{
		log:   logger.NewLogger("USERS_MANAGER", uuid.NewString()),
		users: make(map[string]*User),
	}
}

func (u *UsersManager) HasUser(name string) bool {
	_, exists := u.users[name]
	return exists
}

func (u *UsersManager) AddUser(name string) {
	if !u.HasUser(name) {
		u.users[name] = &User{name: name}
		u.log.Info("user '%s' registered", name)
	}
}

func (u *UsersManager) AddToken(userName, token string) error {
	if !u.HasUser(userName) {
		return ErrUnauthorized
	}
	u.users[userName].AddToken(token)
	return nil
}

func (u *UsersManager) GetUsers() []*User {
	var users []*User
	for _, usr := range u.users {
		users = append(users, usr)
	}
	return users
}
