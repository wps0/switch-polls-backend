package db

import (
	"database/sql"
	"fmt"
	"log"
	"switch-polls-backend/config"
)

type MySQLUserRepository struct {
	Db *sql.DB
}

func NewMySQLUserRepository() MySQLUserRepository {
	return MySQLUserRepository{}
}

func (m *MySQLUserRepository) Init(cfg *config.Configuration) {
	m.Db = m.InitDb(cfg.DbString)
}

func (m *MySQLUserRepository) InitDb(dbString string) *sql.DB {
	log.Println("Initialising database")
	db, err := sql.Open("mysql", dbString)
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec(CREATE_TABLE_USERS_QUERY)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(CREATE_TABLE_POLLS_QUERY)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(CREATE_TABLE_OPTIONS_QUERY)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(CREATE_TABLE_EXTRAS_QUERY)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(CREATE_TABLE_VOTES_QUERY)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(CREATE_TABLE_VOTE_CONFIRMATIONS_QUERY)
	if err != nil {
		panic(err)
	}

	log.Println("Database initialised")
	return db
}

// GetUser Does not support empty values!
func (m *MySQLUserRepository) GetUser(conditions User, createIfDoesNotExist bool) (*User, error) {
	condition, values := ObjectToSQLCondition(AND, conditions, false)
	row := m.Db.QueryRow("SELECT * FROM "+TABLE_USERS+" WHERE "+condition+";", values...)

	var user User
	if err := row.Scan(&user.Id, &user.Email, &user.CreateDate); err != nil {
		if err == sql.ErrNoRows {
			if createIfDoesNotExist {
				return m.CreateUser(conditions)
			}
			return &user, fmt.Errorf("GetUser: user %v not found", conditions)
		}
		return nil, fmt.Errorf("GetUser %v: %v", conditions, err)
	}
	return &user, nil
}

func (m *MySQLUserRepository) CreateUser(user User) (*User, error) {
	res, err := m.Db.Exec("INSERT INTO "+TABLE_USERS+" (`email`) VALUES (?);", user.Email)
	if err != nil {
		return nil, fmt.Errorf("CreateUser %v: %v", user, err)
	}
	affected, err := res.RowsAffected()
	if affected != 1 || err != nil {
		return nil, fmt.Errorf("CreateUser %v: rows affected count other than 1 (err: %v)", user, err)
	}
	id, err := res.LastInsertId()
	if id == 0 || err != nil {
		return nil, fmt.Errorf("CreateUser %v: cannot get last inserted user's id (err: %v) though the query was successful", user, err)
	}
	return m.GetUser(User{Id: int(id)}, false)
}

func (m *MySQLUserRepository) UpdateUser(user User) (*User, error) {
	panic("implement me")
}
