package db

import "database/sql"

// TODO: change the naming to correspond to the one in db
type OptionExtras struct {
	Id       int    `json:"-"`
	OptionId int    `json:"-"`
	Type     string `json:"type"`
	Value    string `json:"value"`
}

type PollOption struct {
	Id      int            `json:"id"`
	PollId  int            `json:"-"`
	Content string         `json:"content"`
	Extras  []OptionExtras `json:"extras"`
}

type Poll struct {
	Id          int          `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Options     []PollOption `json:"options"`
	CreateTime  int64        `json:"-"`
}

type PollVote struct {
	Id          int
	UserId      int
	OptionId    int
	Confirmed   bool
	ConfirmedAt sql.NullInt64
	CreateDate  int64
}

type Confirmation struct {
	Token      string
	VoteId     int
	CreateDate int64
}

type ResultsSummary struct {
	Summary []VoteResult `json:"summary"`
}

type VoteResult struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
	Count   int    `json:"count"`
}

type User struct {
	Id         int    `json:"id"`
	Email      string `json:"email"`
	CreateDate int64  `json:"-"`
}
