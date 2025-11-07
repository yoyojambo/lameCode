package data

import (
	"database/sql"
	"fmt"
	"lameCode/internal/platform/config"
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

	if !config.LocalDB() && config.DbAuthToken() != ""{
		l.Println("with separate token", config.DbAuthToken())
	}
	
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

	if config.ApplySchema() {
		l.Println("Applying schema...")
		err := LoadSchema(db)
		if err != nil {
			l.Fatalln("Failed to apply schema:\n", err)
		}
	}

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


// This function is ONLY splitting by ";\n\n", any double spaces with
// ";" inside a trigger or anything else will break it!
// TODO: A less naive approach that actually is aware of BEGIN/END keywords
var GetSchemaStatements = sync.OnceValue(
	func() []string {
		statements := make([]string, 0, 5)
		// double line-break for trigger with ';' inside, just a hack
		for _, s := range strings.Split(schemaContent, ";\n\n") { 
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
	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("Error running SQL statement: %v \n\n STATEMENT\n>>>\n%s\n<<<", err, stmt)
		}
	}
	
	tx.Commit()
	return nil
}
