package db

import (
	"database/sql"
	"fmt"
)

type MySQLUsersRepository struct {
	Db *sql.DB
}

func NewMySQLUsersRepository() MySQLUsersRepository {
	return MySQLUsersRepository{}
}

func (m *MySQLUsersRepository) Init(db *sql.DB) {
	m.Db = db
}

// GetUser Does not support empty values!
func (m *MySQLUsersRepository) GetUser(conditions User, createIfDoesNotExist bool) (*User, error) {
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

func (m *MySQLUsersRepository) CreateUser(user User) (*User, error) {
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

func (m *MySQLUsersRepository) UpdateUser(user User) (*User, error) {
	panic("implement me")
}
