package data

import (
	"database/sql"
	"lameCode/platform/config"
	"sync"

	_ "modernc.org/sqlite"
)


var loadDB = sync.OnceValue(func () *sql.DB {
	db, err := sql.Open("sqlite", *config.DbFile)
	if err != nil {
		panic(err)
	}

	return db
})

func DB() *sql.DB {
	return loadDB()
}
