package db

import (
	"database/sql"
	"fmt"
)

type MySQLVotesRepository struct {
	db *sql.DB
}

func NewMySQLVotesRepository() MySQLVotesRepository {
	return MySQLVotesRepository{}
}

func (m *MySQLVotesRepository) Init(Db *sql.DB) {
	m.db = Db
}

func (m *MySQLVotesRepository) GetVote(vote PollVote) (*PollVote, error) {
	condition, args := ObjectToSQLCondition(AND, vote, false)
	row := Db.QueryRow("SELECT * FROM "+TableVotes+" WHERE "+condition+";", args...)
	var resVote PollVote
	if err := row.Scan(&resVote.Id, &resVote.UserId, &resVote.OptionId, &resVote.Confirmed, &resVote.ConfirmedAt, &resVote.CreateDate); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("GetVote: vote %v not found %v", vote, err)
		}
		return nil, fmt.Errorf("GetVote %v: %v", vote, err)
	}
	return &resVote, nil
}

func (m *MySQLVotesRepository) CreateVote(vote PollVote) (*PollVote, error) {
	res, err := Db.Exec("INSERT INTO "+TableVotes+"(user_id, option_id) VALUES (?, ?);", vote.UserId, vote.OptionId)
	if err != nil {
		return nil, fmt.Errorf("CreateVote %v: %v", vote, err)
	}
	var insertId int64
	if insertId, err = res.LastInsertId(); err != nil {
		return nil, fmt.Errorf("CreateVote %v - failed to get the id of the inserted row: %v", vote, err)
	}
	insertedVote, err := m.GetVote(PollVote{Id: int(insertId)})
	if err != nil {
		return nil, fmt.Errorf("CreateVote %v - failed to get the inserted row: %v", vote, err)
	}
	return insertedVote, err
}

func (m *MySQLVotesRepository) UpdateVote(poll PollVote) (*PollVote, error) {
	panic("implement me")
}
