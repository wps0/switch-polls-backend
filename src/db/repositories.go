package db

var UsersRepo MySQLUsersRepository
var PollsRepo MySQLPollsRepository

type UsersRepository interface {
	GetUser(user User, createIfDoesNotExist bool) (*User, error)
	CreateUser(user User) (*User, error)
	UpdateUser(user User) (*User, error)
}

type PollsRepository interface {
	GetPoll(poll Poll, recursiveMode bool) (*Poll, error)
	GetPollOption(option PollOption, recursiveMode bool) (PollOption, error)
	CreatePoll(poll Poll) (*Poll, error)
	UpdatePoll(poll Poll) (*Poll, error)
}

type VotesRepository interface {
	GetVote(vote PollVote) (*PollVote, error)
	CreateVote(vote PollVote) (*PollVote, error)
	UpdateVote(vote PollVote) (*PollVote, error)
}
