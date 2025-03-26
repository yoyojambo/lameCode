package config

import (
	"flag"
)

func WasFlagPassed(name string) bool {
    found := false
    flag.Visit(func(f *flag.Flag) {
        if f.Name == name {
            found = true
        }
    })
    return found
}

const DEFAULT_DOT_ENV = ".env"

const DEFAULT_SQLITE_DB_FILE = "database.db"

const db_file_usage = "File with persistent SQLite database (defaults to " + DEFAULT_SQLITE_DB_FILE + ")"

var DbFile = flag.String("DB_FILE", DEFAULT_SQLITE_DB_FILE, db_file_usage)

const env_file_usage = "File with application configuration. Values set in env file OVERRIDE values set by other flags. Defaults to " + DEFAULT_DOT_ENV + ". If not found and "
var envFile = flag.String("env", DEFAULT_DOT_ENV, env_file_usage)
