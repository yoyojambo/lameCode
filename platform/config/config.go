// config sets up configuration for all other modules, as far as user
// configuration is concerned. It exposes functions, flags and
// constants that configure behaviour of other modules.
package config

import (
	"flag"
)

// WasFlagPassed checks if a specific flag (name)
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

var dbFile = flag.String("db-file", DEFAULT_SQLITE_DB_FILE, db_file_usage)

func DbFile() string {
	return *dbFile
}

const env_file_usage = "File with application configuration. Values set in env file OVERRIDE values set by other flags. Defaults to " + DEFAULT_DOT_ENV + ". If not found and "

var envFile = flag.String("env", DEFAULT_DOT_ENV, env_file_usage)

var debug = flag.Bool("debug", false, "Runs server in debug mode, with more logs.")

func Debug() bool {
	return *debug
}

// Currently only local database, Turso later on
func LocalDB() bool {
	return true
}

