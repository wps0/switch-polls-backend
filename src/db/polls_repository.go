package db

import "database/sql"

type MySQLPollsRepository struct {
	Db *sql.DB
}

func NewMySQLPollsRepository() MySQLPollsRepository {
	return MySQLPollsRepository{}
}
