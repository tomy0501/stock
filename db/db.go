package db

import (
	"github.com/jmoiron/sqlx"
)

func NewUserDB(db *sqlx.DB) UserDB {
	return UserDB{
		UserDB: db,
	}
}

type UserDB struct {
	UserDB *sqlx.DB
}
