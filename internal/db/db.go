package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func New(databaseUrl string) (*sql.DB, error) {
	return sql.Open("postgres", databaseUrl)
}
