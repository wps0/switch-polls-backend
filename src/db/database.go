package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"switch-polls-backend/config"
)

var Db *sql.DB

const (
	TABLE_PREFIX  = "spolls_"
	TABLE_POLLS   = TABLE_PREFIX + "polls"
	TABLE_VOTES   = TABLE_PREFIX + "votes"
	TABLE_USERS   = TABLE_PREFIX + "users"
	TABLE_OPTIONS = TABLE_PREFIX + "options"
	TABLE_EXTRAS  = TABLE_PREFIX + "extras"
)

const (
	CREATE_TABLE_USERS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_USERS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	email VARCHAR(128) NOT NULL,
	create_date INT NOT NULL DEFAULT CURRENT_TIMESTAMP()
);`
	CREATE_TABLE_POLLS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_POLLS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	title VARCHAR(256) NOT NULL,
	description VARCHAR(2048) NULL,
	create_date INT NOT NULL DEFAULT CURRENT_TIMESTAMP()
);`
	CREATE_TABLE_OPTIONS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_OPTIONS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	poll_id INT,
	content VARCHAR(1024) NOT NULL,
	INDEX fk_options_poll_ix(poll_id),
	FOREIGN KEY fk_options_poll_ix(poll_id)
        REFERENCES ` + TABLE_POLLS + `(id)
        ON DELETE CASCADE
		ON UPDATE CASCADE

);`
	CREATE_TABLE_EXTRAS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_EXTRAS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	option_id INT NULL,
	type VARCHAR(64) NOT NULL,
	content VARCHAR(2048) NULL,
	INDEX fk_extras_opt_ix (option_id),
	FOREIGN KEY fk_extras_opt_ix(option_id)
        REFERENCES ` + TABLE_OPTIONS + `(id)
        ON DELETE CASCADE
		ON UPDATE CASCADE
);`
	CREATE_TABLE_VOTES_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_VOTES + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	user_id INT NULL,
	option_id INT NULL,
	confirmed BOOLEAN DEFAULT false,
	create_date INT NOT NULL DEFAULT CURRENT_TIMESTAMP(),
	INDEX fk_votes_usr_ix (user_id),
	INDEX fk_votes_opt_ix (option_id),
	FOREIGN KEY fk_votes_usr_ix(user_id)
        REFERENCES ` + TABLE_USERS + `(id)
        ON DELETE SET NULL
		ON UPDATE CASCADE,
	FOREIGN KEY fk_votes_opt_ix(option_id)
        REFERENCES ` + TABLE_OPTIONS + `(id)
        ON DELETE SET NULL
		ON UPDATE CASCADE
);`
)

func InitDb() {
	db, err := sql.Open("mysql", config.Cfg.DbString)
	if err != nil {
		panic(err.Error())
	}
	Db = db

	_, err = Db.Exec(CREATE_TABLE_USERS_QUERY)
	if err != nil {
		fmt.Println("users")
		panic(err)
	}
	_, err = Db.Exec(CREATE_TABLE_POLLS_QUERY)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec(CREATE_TABLE_OPTIONS_QUERY)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec(CREATE_TABLE_EXTRAS_QUERY)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec(CREATE_TABLE_VOTES_QUERY)
	if err != nil {
		panic(err)
	}
}

func GetOptionExtrasByOptionId(optId int) ([]OptionExtras, error) {
	res, err := Db.Query("SELECT * FROM "+TABLE_EXTRAS+" WHERE option_id = ?;", optId)
	if err != nil {
		return make([]OptionExtras, 0), err
	}

	var extras = make([]OptionExtras, 0)
	for res.Next() {
		var opt OptionExtras
		var tmp int
		err = res.Scan(&tmp, &tmp, &opt.Type, &opt.Value)

		if err != nil {
			log.Println("option extras scan error (optionId:", optId, "): ", err)
			return make([]OptionExtras, 0), err
		}
		extras = append(extras, opt)
	}
	return extras, nil
}

func GetOptionsByPollId(pollId int) ([]PollOption, error) {
	res, err := Db.Query("SELECT * FROM "+TABLE_OPTIONS+" WHERE poll_id = ?;", pollId)
	if err != nil {
		return make([]PollOption, 0), err
	}

	var options = make([]PollOption, 0)
	for res.Next() {
		var opt PollOption
		var tmp int
		err = res.Scan(&opt.Id, &tmp, &opt.Content)

		if err != nil {
			log.Println("option scan error (pollId:", pollId, "): ", err)
			return make([]PollOption, 0), err
		}
		opt.Extras, _ = GetOptionExtrasByOptionId(opt.Id)
		options = append(options, opt)
	}
	return options, nil
}

func GetPollById(id int) (*Poll, error) {
	res, err := Db.Query("SELECT * FROM "+TABLE_POLLS+" WHERE id = ?;", id)
	if err != nil {
		return nil, err
	}

	var poll Poll
	if res.Next() {
		err = res.Scan(&poll.Id, &poll.Title, &poll.Description, &poll.CreateTime)
		if err != nil {
			return nil, err
		}
	}
	poll.Options, err = GetOptionsByPollId(id)
	if err != nil {
		log.Println("Get options by poll id error: ", err)
		return nil, nil
	}

	return &poll, nil
}
