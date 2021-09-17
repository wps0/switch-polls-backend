package db

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"switch-polls-backend/config"
	"switch-polls-backend/utils"
)

var Db *sql.DB

const (
	TABLE_PREFIX        = "spolls_"
	TABLE_POLLS         = TABLE_PREFIX + "polls"
	TABLE_VOTES         = TABLE_PREFIX + "votes"
	TABLE_USERS         = TABLE_PREFIX + "users"
	TABLE_OPTIONS       = TABLE_PREFIX + "options"
	TABLE_EXTRAS        = TABLE_PREFIX + "extras"
	TABLE_CONFIRMATIONS = TABLE_PREFIX + "confirmations"
)

const (
	CREATE_TABLE_USERS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_USERS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	email VARCHAR(128) NOT NULL,
	create_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
	INDEX ix_users_email(email)
);`
	CREATE_TABLE_POLLS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_POLLS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	title VARCHAR(256) NOT NULL,
	description VARCHAR(2048) NULL,
	create_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP()
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
	confirmed_at TIMESTAMP NULL,
	create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP(),
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
	CREATE_TABLE_VOTE_CONFIRMATIONS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_CONFIRMATIONS + "`" + ` (
	token VARCHAR(192) NOT NULL PRIMARY KEY,
	vote_id INT NOT NULL,
	create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP(),
	INDEX fk_confirmations_vote_id_ix (vote_id),
	FOREIGN KEY fk_confirmations_vote_id_ix(vote_id)
        REFERENCES ` + TABLE_VOTES + `(id)
        ON DELETE CASCADE
		ON UPDATE CASCADE
);`
)

func InitDb() {
	log.Println("Initialising database")
	db, err := sql.Open("mysql", config.Cfg.DbString)
	if err != nil {
		panic(err.Error())
	}
	Db = db

	_, err = Db.Exec(CREATE_TABLE_USERS_QUERY)
	if err != nil {
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
	_, err = Db.Exec(CREATE_TABLE_VOTE_CONFIRMATIONS_QUERY)
	if err != nil {
		panic(err)
	}
	log.Println("Database initialised")
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

func GetPollIdByOptionId(id int) (int, error) {
	res, err := Db.Query("SELECT poll_id FROM "+TABLE_OPTIONS+" WHERE id = ?;", id)
	if err != nil {
		return 0, err
	}
	if !res.Next() {
		return 0, errors.New("option id " + strconv.Itoa(id) + " not found")
	}
	var pollId int
	err = res.Scan(&pollId)
	if err != nil {
		return 0, err
	}
	return pollId, nil
}

func GetUserIdByEmail(email string) (int, error) {
	res, err := Db.Query("SELECT id FROM " + TABLE_USERS + " WHERE email = '" + email + "';")
	if err != nil {
		log.Println("get user by id error", err)
		return 0, err
	}
	if !res.Next() {
		log.Println("get user by id user does not exist")
		return 0, errors.New("user with the specified email was not found")
	}
	var id int
	err = res.Scan(&id)
	if err != nil {
		log.Println("get user id scan error", err)
		return 0, err
	}
	return id, nil
}

func CheckIfUserHasAlreadyVoted(userEmail string, pollId int) (bool, error) {
	if !utils.IsAlphaWithAtAndDot(userEmail) {
		return false, errors.New("invalid email format")
	}
	res, err := Db.Query(`
SELECT
	V.confirmed
FROM `+TABLE_VOTES+` V INNER JOIN `+TABLE_USERS+`
		U ON V.user_id = U.id
	INNER JOIN `+TABLE_OPTIONS+` O ON
		V.option_id = O.id
WHERE O.poll_id = ? AND V.confirmed = 1 AND U.email = '`+userEmail+`';`, pollId)
	if err != nil {
		log.Printf("error when checking if user `%s` has already voted on poll `%d`: %v", userEmail, pollId, err)
		return false, err
	}
	return res.Next(), nil
}

func CheckIfUserExists(email string) bool {
	res, err := Db.Query("SELECT id FROM " + TABLE_USERS + " WHERE email = '" + email + "';")
	if err != nil {
		log.Println("check if user exists error", err)
		return false
	}
	return res.Next()
}

func InsertUser(email string) (int, error) {
	res, err := Db.Exec("INSERT INTO " + TABLE_USERS + "(email) VALUES ('" + email + "');")
	if err != nil {
		return 0, err
	}
	if rows, err := res.RowsAffected(); err != nil || rows != 1 {
		return 0, err
	}
	return GetUserIdByEmail(email)
}

func InsertVote(userEmail string, optId int) (int, error) {
	var userId int
	var err error
	if !CheckIfUserExists(userEmail) {
		userId, err = InsertUser(userEmail)
		if err != nil {
			return 0, err
		}
	} else {
		userId, err = GetUserIdByEmail(userEmail)
		if err != nil {
			return 0, err
		}
	}
	res, err := Db.Exec("INSERT INTO "+TABLE_VOTES+"(user_id, option_id) VALUES (?, ?);", userId, optId)
	if err != nil {
		return 0, err
	}
	var insertId int64
	if insertId, err = res.LastInsertId(); err != nil {
		return 0, err
	}
	return int(insertId), nil
}

func InsertToken(token string, voteId int) error {
	res, err := Db.Exec("INSERT INTO "+TABLE_CONFIRMATIONS+"(token, vote_id) VALUES ('"+token+"', ?);", voteId)
	if err != nil {
		return err
	}
	if rows, err := res.RowsAffected(); err != nil || rows != 1 {
		return err
	}
	return err
}
