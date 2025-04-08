package data

import (
	"database/sql"
	"lameCode/platform/config"
	"log"
	"sync"

	_ "modernc.org/sqlite"
)

// Creates database connection based on application configuration
var loadDB = sync.OnceValue(func () *sql.DB {
	db, err := sql.Open("sqlite", *config.DbFile)
	if err != nil {
		panic(err)
	}

	log.Println("[data/client] Initialized SQL conn to", *config.DbFile)

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
