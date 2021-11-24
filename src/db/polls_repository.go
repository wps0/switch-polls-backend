package db

import (
	"database/sql"
	"fmt"
)

type MySQLPollsRepository struct {
	Db *sql.DB
}

func NewMySQLPollsRepository() MySQLPollsRepository {
	return MySQLPollsRepository{}
}

func (m *MySQLPollsRepository) Init(db *sql.DB) {
	m.Db = db
}

//GetPoll does not support default values!
// recursiveMode - return the whole poll structure, together with pollOptions and optionExtras
func (m *MySQLPollsRepository) GetPoll(cond Poll, recursiveMode bool) (*Poll, error) {
	condition, values := ObjectToSQLCondition(AND, cond, false)
	row := m.Db.QueryRow("SELECT * FROM "+TABLE_POLLS+" WHERE "+condition+";", values...)

	var poll Poll
	if err := row.Scan(&poll.Id, &poll.Title, &poll.Description, &poll.CreateTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("GetPoll: poll %v not found", cond)
		}
		return nil, fmt.Errorf("GetPoll %v: %v", cond, err)
	}
	if recursiveMode {
		options, err := m.GetPollOptions(poll.Id, true)
		if err != nil {
			return &poll, fmt.Errorf("GetPoll %v: failed build the whole object relations %v", cond, err)
		}
		poll.Options = options
	}

	return &poll, nil
}

func (m *MySQLPollsRepository) GetPollOption(cond PollOption, recursiveMode bool) (PollOption, error) {
	conditionString, args := ObjectToSQLCondition(AND, cond, false)
	row := m.Db.QueryRow("SELECT * FROM "+TABLE_OPTIONS+" WHERE "+conditionString, args...)
	var option PollOption
	if err := row.Scan(&option.Id, &option.PollId, &option.Content); err != nil {
		return PollOption{}, fmt.Errorf("GetPollOption %v: %v", cond, err)
	}
	if recursiveMode {
		var err error
		option.Extras, err = m.GetOptionExtras(option.Id)
		if err != nil {
			return PollOption{}, fmt.Errorf("GetPollOption cannot get option extras %v: %v", cond, err)
		}
	}

	if err := row.Err(); err != nil {
		return PollOption{}, fmt.Errorf("GetPollOption %v: %v", cond, err)
	}
	return option, nil
}

func (m *MySQLPollsRepository) GetPollOptions(pollId int, recursiveMode bool) ([]PollOption, error) {
	rows, err := m.Db.Query("SELECT * FROM "+TABLE_OPTIONS+" AS O WHERE O.poll_id = ?;", pollId)
	if err != nil {
		return nil, fmt.Errorf("GetPollOptions %d: %v", pollId, err)
	}
	defer rows.Close()
	options := make([]PollOption, 0)
	for rows.Next() {
		var opt PollOption
		if err = rows.Scan(&opt.Id, &opt.PollId, &opt.Content); err != nil {
			return nil, fmt.Errorf("GetPollOptions %d: %v", pollId, err)
		}

		if recursiveMode {
			opt.Extras, err = m.GetOptionExtras(opt.Id)
			if err != nil {
				return nil, fmt.Errorf("GetPollOptions cannot get option (%v) extras %d: %v", opt, pollId, err)
			}
		}

		options = append(options, opt)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPollOptions %d: %v", pollId, err)
	}
	return options, nil
}

func (m *MySQLPollsRepository) GetOptionExtras(optionId int) ([]OptionExtras, error) {
	res, err := Db.Query("SELECT * FROM "+TABLE_EXTRAS+" WHERE option_id = ?;", optionId)
	if err != nil {
		return make([]OptionExtras, 0), fmt.Errorf("GetOptionExtras %d: %v", optionId, err)
	}
	defer res.Close()

	extras := make([]OptionExtras, 0)
	for res.Next() {
		var extra OptionExtras
		var tmp int
		err = res.Scan(&extra.Id, &tmp, &extra.Type, &extra.Value)

		if err != nil {
			return make([]OptionExtras, 0), fmt.Errorf("GetOptionExtras %d: %v", optionId, err)
		}
		extras = append(extras, extra)
	}
	return extras, nil
}

func (m *MySQLPollsRepository) CreatePoll(poll Poll) (*Poll, error) {
	panic("implement me")
}

func (m *MySQLPollsRepository) UpdatePoll(poll Poll) (*Poll, error) {
	panic("implement me")
}
