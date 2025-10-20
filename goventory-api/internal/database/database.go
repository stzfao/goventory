package database

import (
	"database/sql"
	_ "embed"
	
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

// create and return a new sqlite dbconn instancce
func NewDB(dataSourceName string) (*sql.DB, error) {
	// open dbconn. create if doesnt exist
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	// DB ping! :D
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Execute the schema for thefirst run
	if _, err = db.Exec(ddl); err != nil {
		return nil, err
	}

	return db, nil
}
