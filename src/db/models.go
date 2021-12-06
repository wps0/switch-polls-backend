package db

import "database/sql"

type OptionExtras struct {
	Id       int    `json:"-" db:"id"`
	OptionId int    `json:"-" db:"option_id"`
	Type     string `json:"type" db:"type"`
	Value    string `json:"value" db:"content"`
}

type PollOption struct {
	Id      int            `json:"id" db:"id"`
	PollId  int            `json:"-" db:"poll_id"`
	Content string         `json:"content" db:"content"`
	Extras  []OptionExtras `json:"extras" db:"-"`
}

type Poll struct {
	Id          int          `json:"id" db:"id"`
	Title       string       `json:"title" db:"title"`
	Description string       `json:"description" db:"description"`
	Options     []PollOption `json:"options" db:"-"`
	CreateTime  int64        `json:"-" db:"create_time"`
}

type PollVote struct {
	Id          int           `db:"id"`
	UserId      int           `db:"user_id"`
	OptionId    int           `db:"option_id"`
	ConfirmedAt sql.NullInt64 `db:"confirmed_at"`
	CreateDate  int64         `db:"create_date"`
}

type Confirmation struct {
	Token      string `db:"token"`
	VoteId     int    `db:"vote_id"`
	CreateDate int64  `db:"create_date"`
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
	Id         int    `json:"id" db:"id"`
	Email      string `json:"email" db:"email"`
	CreateDate int64  `json:"-" db:"create_date"`
}
