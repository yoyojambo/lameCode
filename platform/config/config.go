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

type Config struct {
	debug  bool
	db_url string // Either local (path) or remote (turso)
}

const (
	DEFAULT_DOT_ENV        = ".env"        // Default .env location
	DEFAULT_SQLITE_DB_FILE = "database.db" // Default database location
)

// Usage messages
const (
	db_file_usage    = "File with persistent SQLite database. (defaults to " + DEFAULT_SQLITE_DB_FILE + ")"
	env_file_usage   = "File with application configuration. Values set in env file OVERRIDE values set by other flags. (defaults to " + DEFAULT_DOT_ENV + ")"
	create_db_usage  = "Try to apply schema, will not crash if it can't."
	debug_usage      = "Runs server in debug mode, with more logs. (defaults to false)"
	remote_usage     = "The provided database path is a Turso database URL. (defaults to false)"
	auth_token_usage = "Token to access turso database. NOT RECOMMENDED, USE ENVIRONMENT VARIABLES INSTEAD."
)

// Flags that can be set from command line
var (
	db_URL     string
	create     bool
	envFile    string
	debug      bool = true
	remote     bool
	turso_auth string
)

// Activate flags designated for server configuration.
// TODO: Maybe implement this as flagsets instead? This is getting ugly
func LoadServerFlags() {
	flag.StringVar(&db_URL, "db-url", DEFAULT_SQLITE_DB_FILE, db_file_usage)
	flag.StringVar(&turso_auth, "token", "", auth_token_usage)
	flag.BoolVar(&create, "create-db", false, create_db_usage)              
	flag.StringVar(&envFile, "env", DEFAULT_DOT_ENV, env_file_usage)         
	flag.BoolVar(&debug, "debug", false, debug_usage)
	flag.BoolVar(&remote, "remote", false, remote_usage)                    
}

func DbUrl() string {
	return db_URL
}

func DbAuthToken() string {
	return turso_auth
}

func Debug() bool {
	return debug
}

// Currently only local database, Turso later on
func LocalDB() bool {
	return !remote
}
