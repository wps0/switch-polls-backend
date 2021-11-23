package db

var usersRepo MySQLUserRepository

type UsersRepository interface {
	GetUser(user User) (*User, error)
	CreateUser(user User) (*User, error)
	UpdateUser(user User) (*User, error)
}
