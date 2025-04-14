package data

import (
	"database/sql"
	"lameCode/platform/config"
	"log"
	"strings"
	"sync"

	_ "embed"

	_ "modernc.org/sqlite"
)

// Creates database connection based on application configuration
var loadDB = sync.OnceValue(func() *sql.DB {
	db, err := sql.Open("sqlite", config.DbFile())
	if err != nil {
		panic(err)
	}

	log.Println("[data/client] Initialized SQL conn to", config.DbFile())

	return db
})

func DB() *sql.DB {
	return loadDB()
}

// Creates and saves a *Queries object from the configured database connection.
var loadRepo = sync.OnceValue(func() *Queries {
	return New(DB())
})

func Repository() *Queries {
	return loadRepo()
}

//go:embed schema.sql
var schemaContent string

var GetSchemaStatements = sync.OnceValue(
	func() []string {
		statements := make([]string, 0, 5)
		for _, s := range strings.Split(schemaContent, ";") {
			statements = append(statements, strings.TrimSpace(s))
		}

		return statements
	})

func LoadSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmts := GetSchemaStatements()
	for i := range stmts {
		_, err := tx.Exec(stmts[i])
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	
	tx.Commit()
	return nil
}
