package db

var usersRepo MySQLUsersRepository

type UsersRepository interface {
	GetUser(user User, createIfDoesNotExist bool) (*User, error)
	CreateUser(user User) (*User, error)
	UpdateUser(user User) (*User, error)
}

type PollsRepository interface {
	GetPoll(poll Poll) (*Poll, error)
	CreatePoll(poll Poll) (*Poll, error)
	UpdatePoll(poll Poll) (*Poll, error)
}
