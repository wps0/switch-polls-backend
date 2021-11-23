package db

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"switch-polls-backend/config"
	"switch-polls-backend/utils"
	"time"
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
	create_date BIGINT NOT NULL DEFAULT UNIX_TIMESTAMP(),
	INDEX ix_users_email(email)
);`
	CREATE_TABLE_POLLS_QUERY = `
CREATE TABLE IF NOT EXISTS ` + "`" + TABLE_POLLS + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	title VARCHAR(256) NOT NULL,
	description VARCHAR(2048) NULL,
	create_date BIGINT NOT NULL DEFAULT UNIX_TIMESTAMP()
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
	confirmed_at BIGINT NULL,
	create_date BIGINT DEFAULT UNIX_TIMESTAMP(),
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
	create_date BIGINT DEFAULT UNIX_TIMESTAMP(),
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
	usersRepo = NewMySQLUserRepository()
	usersRepo.Init(config.Cfg)
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
	defer res.Close()

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
	defer res.Close()

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
	var poll Poll
	err := Db.QueryRow("SELECT * FROM "+TABLE_POLLS+" WHERE id = ?;", id).Scan(&poll.Id, &poll.Title, &poll.Description, &poll.CreateTime)
	if err != nil {
		return nil, err
	}

	poll.Options, err = GetOptionsByPollId(id)
	if err != nil {
		log.Println("Get options by poll id error: ", err)
		return nil, nil
	}

	return &poll, nil
}

func GetPollIdByOptionId(id int) (int, error) {
	var pollId int
	err := Db.QueryRow("SELECT poll_id FROM "+TABLE_OPTIONS+" WHERE id = ?;", id).Scan(&pollId)
	if err != nil {
		return 0, err
	}
	return pollId, nil
}

func GetVoteById(id int) (*PollVote, error) {
	var vote PollVote
	err := Db.QueryRow("SELECT * FROM "+TABLE_VOTES+" WHERE id = ?;", id).Scan(&vote.Id, &vote.UserId, &vote.OptionId, &vote.Confirmed, &vote.ConfirmedAt, &vote.CreateDate)
	if err != nil {
		return nil, err
	}

	return &vote, nil
}

func GetConfirmationByToken(token string) (*Confirmation, error) {
	var cnf Confirmation
	err := Db.QueryRow("SELECT * FROM "+TABLE_CONFIRMATIONS+" WHERE token = '"+token+"';").Scan(&cnf.Token, &cnf.VoteId, &cnf.CreateDate)
	if err != nil {
		return nil, err
	}

	return &cnf, nil
}

func PrepareResultsSummary(pollId int) (*ResultsSummary, error) {
	res, err := Db.Query(`
SELECT O.id, O.content, COUNT(*) 
FROM `+TABLE_VOTES+` V INNER JOIN `+TABLE_OPTIONS+` O ON V.option_id = O.id 
WHERE O.poll_id = ? AND confirmed = 1 GROUP BY O.id;`, pollId)
	if err != nil {
		log.Println("prepare results error", err)
		return nil, err
	}
	defer res.Close()
	var summary = make([]VoteResult, 0)
	for res.Next() {
		var result VoteResult
		err = res.Scan(&result.Id, &result.Content, &result.Count)
		if err != nil {
			return nil, err
		}
		summary = append(summary, result)
	}
	return &ResultsSummary{summary}, nil
}

func CheckIfUserHasAlreadyVoted(userEmail string, pollId int) (bool, error) {
	if !utils.IsAlphaWithAtAndDot(userEmail) {
		return false, errors.New("invalid email format")
	}
	user, err := usersRepo.GetUser(User{
		Email: userEmail,
	})
	if err != nil {
		return false, err
	}
	return CheckIfUserHasAlreadyVotedById(user.Id, pollId)
}

func CheckIfUserHasAlreadyVotedById(userId int, pollId int) (bool, error) {
	res, err := Db.Query(`
SELECT
	V.confirmed
FROM `+TABLE_VOTES+` V INNER JOIN `+TABLE_USERS+`
		U ON V.user_id = U.id
	INNER JOIN `+TABLE_OPTIONS+` O ON
		V.option_id = O.id
WHERE O.poll_id = ? AND V.confirmed = 1 AND U.id = ?;`, pollId, userId)
	if err != nil {
		log.Printf("error when checking if user `%d` has already voted on poll `%d`: %v", userId, pollId, err)
		return false, err
	}
	defer res.Close()
	return res.Next(), res.Err()
}

func CheckIfUserExists(email string) bool {
	res, err := Db.Query("SELECT id FROM " + TABLE_USERS + " WHERE email = '" + email + "';")
	if err != nil {
		log.Println("check if user exists error", err)
		return false
	}
	defer res.Close()
	return res.Next()
}

func InsertVote(userEmail string, optId int) (int, error) {
	var user *User
	var err error
	if !CheckIfUserExists(userEmail) {
		user, err = usersRepo.CreateUser(User{Email: userEmail})
		if err != nil {
			return 0, err
		}
	} else {
		user, err = usersRepo.GetUser(User{Email: userEmail})
		if err != nil {
			return 0, err
		}
	}
	res, err := Db.Exec("INSERT INTO "+TABLE_VOTES+"(user_id, option_id) VALUES (?, ?);", user.Id, optId)
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

func ChangeConfirmationStatus(voteId int, newStatus bool) error {
	res, err := Db.Exec("UPDATE "+TABLE_VOTES+" SET confirmed = ?, confirmed_at = ? WHERE id = ?;", newStatus, time.Now().Unix(), voteId)
	if err != nil {
		return err
	}
	if rows, err := res.RowsAffected(); err != nil || rows != 1 {
		return err
	}
	return err
}
