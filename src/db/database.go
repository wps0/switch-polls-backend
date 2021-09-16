package db

import (
	"database/sql"
	"switch-polls-backend/config"
)

var Db *sql.DB

const (
	TABLE_PREFIX = "spolls_"
	TABLE_POLLS = TABLE_PREFIX + "polls"
	TABLE_VOTES = TABLE_PREFIX + "votes"
	TABLE_USERS = TABLE_PREFIX + "users"
	TABLE_OPTIONS = TABLE_PREFIX + "options"
	TABLE_EXTRAS = TABLE_PREFIX + "extras"
)

const (
	CREATE_TABLE_USERS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_USERS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	email VARCHAR(128) NOT NULL,
	create_date INT NOT NULL DEFAULT CURRENT_TIMESTAMP
`
	CREATE_TABLE_POLLS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_POLLS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	title VARCHAR(256) NOT NULL,
	description VARCHAR(2048) NULL,
	create_date INT NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
	CREATE_TABLE_OPTIONS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_OPTIONS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	content VARCHAR(1024) NOT NULL
);`
	CREATE_TABLE_EXTRAS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_EXTRAS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	option_id INT NULL,
	type VARCHAR(64) NOT NULL,
	content VARCHAR(2048) NULL,
	INDEX fk_opt_ix (option_id),
	FOREIGN KEY fk_opt_ix(option_id)
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
	create_date INT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	INDEX fk_usr_ix (user_id),
	INDEX fk_opt_ix (option_id),
	FOREIGN KEY fk_usr_ix(user_id)
        REFERENCES ` + TABLE_USERS + `(id)
        ON DELETE SET NULL
		ON UPDATE CASCADE,
	FOREIGN KEY fk_opt_ix(option_id)
        REFERENCES ` + TABLE_OPTIONS + `(id)
        ON DELETE SET NULL
		ON UPDATE CASCADE
);`
)

func InitDb() {
	//"user:password@/dbname"
	db, err := sql.Open("mysql", config.Cfg.DbString)
	if err != nil {
		panic(err.Error())
	}
	Db = db
}

func GetPollById(id int) (*Poll, error) {
	res, err := Db.Query("SELECT * FROM ? WHERE id = ?; ", TABLE_POLLS, id)
	if err != nil {
		return nil, err
	}

	var poll Poll
	err = res.Scan(&poll)
	if err != nil {
		return nil, err
	}
	return &poll, nil
}