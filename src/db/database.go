package db

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
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

func OpenDbInstance() *sql.DB {
	c, err := mysql.ParseDSN(config.Cfg.DbString)
	if err != nil {
		log.Fatal(err.Error())
	}
	c.ParseTime = true
	c.MultiStatements = true
	db, err := sql.Open("mysql", c.FormatDSN())
	if err != nil {
		log.Fatal(err.Error())
	}
	return db
}

func InitDb() {
	log.Println("Initialising database...")
	if config.Cfg.DebugMode {
		log.Printf("Logging in with %s...", config.Cfg.DbString)
	}
	Db = OpenDbInstance()
	log.Println("Database initialised.")
	log.Println("Initialising repositories...")

	UsersRepo = NewMySQLUsersRepository()
	PollsRepo = NewMySQLPollsRepository()
	VotesRepo = NewMySQLVotesRepository()
	UsersRepo.Init(Db)
	PollsRepo.Init(Db)
	VotesRepo.Init(Db)
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
