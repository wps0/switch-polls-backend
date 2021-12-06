package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"switch-polls-backend/config"
)

var Db *sql.DB

const (
	TablePrefix        = "spolls_"
	TablePolls         = TablePrefix + "polls"
	TableVotes         = TablePrefix + "votes"
	TableUsers         = TablePrefix + "users"
	TableOptions       = TablePrefix + "options"
	TableExtras        = TablePrefix + "extras"
	TableConfirmations = TablePrefix + "confirmations"
)

const (
	CreateTableUsersQuery = `
CREATE TABLE IF NOT EXISTS ` + "`" + TableUsers + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	email VARCHAR(128) NOT NULL,
	create_date BIGINT NOT NULL DEFAULT UNIX_TIMESTAMP(),
	INDEX ix_users_email(email)
);`
	CreateTablePollsQuery = `
CREATE TABLE IF NOT EXISTS ` + "`" + TablePolls + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	title VARCHAR(256) NOT NULL,
	description VARCHAR(2048) NULL,
	create_date BIGINT NOT NULL DEFAULT UNIX_TIMESTAMP()
);`
	CreateTableOptionsQuery = `
CREATE TABLE IF NOT EXISTS ` + "`" + TableOptions + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	poll_id INT,
	content VARCHAR(1024) NOT NULL,
	INDEX fk_options_poll_ix(poll_id),
	FOREIGN KEY fk_options_poll_ix(poll_id)
        REFERENCES ` + TablePolls + `(id)
        ON DELETE CASCADE
		ON UPDATE CASCADE
);`
	CreateTableExtrasQuery = `
CREATE TABLE IF NOT EXISTS ` + "`" + TableExtras + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	option_id INT NULL,
	type VARCHAR(64) NOT NULL,
	content VARCHAR(2048) NULL,
	INDEX fk_extras_opt_ix (option_id),
	FOREIGN KEY fk_extras_opt_ix(option_id)
        REFERENCES ` + TableOptions + `(id)
        ON DELETE CASCADE
		ON UPDATE CASCADE
);`
	CreateTableVotesQuery = `
CREATE TABLE IF NOT EXISTS ` + "`" + TableVotes + "`" + ` (
	id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
	user_id INT NULL,
	option_id INT NULL,
	confirmed_at BIGINT NULL,
	create_date BIGINT DEFAULT UNIX_TIMESTAMP(),
	INDEX fk_votes_usr_ix (user_id),
	INDEX fk_votes_opt_ix (option_id),
	FOREIGN KEY fk_votes_usr_ix(user_id)
        REFERENCES ` + TableUsers + `(id)
        ON DELETE CASCADE
		ON UPDATE CASCADE,
	FOREIGN KEY fk_votes_opt_ix(option_id)
        REFERENCES ` + TableOptions + `(id)
        ON DELETE CASCADE
		ON UPDATE CASCADE
);`
	CreateTableVoteConfirmationsQuery = `
CREATE TABLE IF NOT EXISTS ` + "`" + TableConfirmations + "`" + ` (
	token VARCHAR(192) NOT NULL PRIMARY KEY,
	vote_id INT NOT NULL,
	create_date BIGINT DEFAULT UNIX_TIMESTAMP(),
	INDEX fk_confirmations_vote_id_ix (vote_id),
	FOREIGN KEY fk_confirmations_vote_id_ix(vote_id)
        REFERENCES ` + TableVotes + `(id)
        ON DELETE CASCADE
		ON UPDATE CASCADE
);`
)

func InitDb() {
	log.Println("Initialising database...")
	db, err := sql.Open("mysql", config.Cfg.DbString)
	if err != nil {
		panic(err.Error())
	}
	Db = db

	_, err = Db.Exec(CreateTableUsersQuery)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec(CreateTablePollsQuery)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec(CreateTableOptionsQuery)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec(CreateTableExtrasQuery)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec(CreateTableVotesQuery)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec(CreateTableVoteConfirmationsQuery)
	if err != nil {
		panic(err)
	}
	log.Println("Database initialised.")
	log.Println("Initialising repositories...")

	UsersRepo = NewMySQLUsersRepository()
	PollsRepo = NewMySQLPollsRepository()
	VotesRepo = NewMySQLVotesRepository()
	UsersRepo.Init(db)
	PollsRepo.Init(db)
	VotesRepo.Init(db)
	log.Println("Repositories initialised.")
}

func GetConfirmationByToken(token string) (*Confirmation, error) {
	var cnf Confirmation
	err := Db.QueryRow("SELECT * FROM "+TableConfirmations+" WHERE token = ?;", token).Scan(&cnf.Token, &cnf.VoteId, &cnf.CreateDate)
	if err != nil {
		return nil, err
	}

	return &cnf, nil
}

func PrepareResultsSummary(pollId int) (*ResultsSummary, error) {
	res, err := Db.Query(`
SELECT O.id, O.content, COUNT(*) 
FROM `+TableVotes+` V INNER JOIN `+TableOptions+` O ON V.option_id = O.id 
WHERE O.poll_id = ? AND confirmed_at IS NOT NULL GROUP BY O.id;`, pollId)
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

func CheckIfUserHasAlreadyVotedById(userId int, pollId int) (bool, error) {
	res, err := Db.Query(`
SELECT
	V.confirmed_at
FROM `+TableVotes+` V INNER JOIN `+TableUsers+`
		U ON V.user_id = U.id
	INNER JOIN `+TableOptions+` O ON
		V.option_id = O.id
WHERE O.poll_id = ? AND V.confirmed_at IS NOT NULL AND U.id = ?;`, pollId, userId)
	if err != nil {
		log.Printf("error when checking if user `%d` has already voted on poll `%d`: %v", userId, pollId, err)
		return false, err
	}
	defer res.Close()
	return res.Next(), res.Err()
}

func InsertToken(token string, voteId int) error {
	res, err := Db.Exec("INSERT INTO "+TableConfirmations+"(token, vote_id) VALUES (?, ?);", token, voteId)
	if err != nil {
		return err
	}
	if rows, err := res.RowsAffected(); err != nil || rows != 1 {
		return err
	}
	return err
}

func ChangeConfirmationStatus(voteId int, confirmedAt int64) error {
	res, err := Db.Exec("UPDATE "+TableVotes+" SET confirmed_at = ? WHERE id = ?;", confirmedAt, voteId)
	if err != nil {
		return err
	}
	if rows, err := res.RowsAffected(); err != nil || rows != 1 {
		return err
	}
	return err
}
