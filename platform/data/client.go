package data

import (
	"database/sql"
	"lameCode/platform/config"
	"log"
	"os"
	"strings"
	"sync"

	_ "embed"

	"net/url"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

// Creates database connection based on application configuration
var loadDB = sync.OnceValue(func() *sql.DB {
	var db *sql.DB
	var err error
	
	l := log.New(os.Stdout, "[data/client] ", log.LstdFlags | log.Lmsgprefix)
	l.Println("Initializing connection to", config.DbUrl())
	l.Println("with separate token", config.DbAuthToken())

	if config.LocalDB() {
		db, err = sql.Open("sqlite", config.DbUrl())
	} else {
		u, err := url.Parse(config.DbUrl())
		if err != nil {
			l.Fatalf("Could not parse url for remote database connection: %v", err)
		}
		q := u.Query()

		if u.Scheme != "libsql" {
			u.Scheme = "libsql"
		}
		// Set token if in auth flag
		// Overrides if it was already in the url (on purpose)
		if config.DbAuthToken() == "" {
			if !q.Has("authToken") {
				l.Fatalln("No auth token found in URL or --auth flag")
			}
		} else {
			if q.Has("authToken") {
				l.Println("Overriding auth token in database URL")
			}
			q.Set("authToken", config.DbAuthToken())
			u.RawQuery = q.Encode()
		}

		l.Println("Connecting with finished URL")
		db, err = sql.Open("libsql", u.String())
	}

	// Handle error in sql.Open (remote OR local)
	if err != nil {
		panic(err)
	}

	l.Println("Initialized SQL conn to", config.DbUrl())

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
