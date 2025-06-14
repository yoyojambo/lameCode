// config sets up configuration for all other modules, as far as user
// configuration is concerned. It exposes functions, flags and
// constants that configure behaviour of other modules.
package config

import (
	"flag"
	"lameCode/platform/session"
)

const (
	DEFAULT_DOT_ENV        = ".env"        // Default .env location
	DEFAULT_SQLITE_DB_FILE = "database.db" // Default database location
)

// Usage messages
const (
	db_file_usage        = "File with persistent SQLite database. (defaults to " + DEFAULT_SQLITE_DB_FILE + ")"
	env_file_usage       = "File with application configuration. Values set in env file OVERRIDE values set by other flags. (defaults to " + DEFAULT_DOT_ENV + ")"
	create_db_usage      = "Try to apply schema, will not crash if it can't."
	debug_usage          = "Runs server in debug mode, with more logs. (defaults to false)"
	remote_usage         = "The provided database path is a Turso database URL. (defaults to false)"
	auth_token_usage     = "Token to access turso database. NOT RECOMMENDED, USE ENVIRONMENT VARIABLES INSTEAD."
	install_wasmer_usage = "Indicates whether to run wasmer.io installer script if no WASM runtime is found."
	jwt_secret_usage     = "Sets a secret to sign and verify JWT tokens for user sessions. Multiplexing requests between multiple instances naively will require to set this to a shared secret."
)

// Flags that can be set from command line
var (
	db_URL         string
	create         bool
	envFile        string
	debug          bool = true
	remote         bool
	turso_auth     string
	install_wasmer bool
	jwt_secret     string
)

// Activate flags designated for server configuration.
// Separating declaration and calling flag.*Var functions allows using
// the config package without polluting command's flags.
// TODO: Maybe implement this as flagsets instead? This is getting ugly
func LoadServerFlags() {
	flag.StringVar(&db_URL, "db-url", DEFAULT_SQLITE_DB_FILE, db_file_usage)
	flag.StringVar(&turso_auth, "token", "", auth_token_usage)
	flag.BoolVar(&create, "create-db", false, create_db_usage)
	flag.StringVar(&envFile, "env", DEFAULT_DOT_ENV, env_file_usage)
	flag.BoolVar(&debug, "debug", false, debug_usage)
	flag.BoolVar(&remote, "remote", false, remote_usage)
	flag.BoolVar(&install_wasmer, "install-wasmer", false, install_wasmer_usage)
	flag.StringVar(&jwt_secret, "jwt_secret", session.GenerateRandJwtSecret(), jwt_secret_usage)
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

func LocalDB() bool {
	return !remote
}

func InstallWasmer() bool {
	return install_wasmer
}

func ApplySchema() bool {
	return create
}

func JwtSecret() string {
	return jwt_secret
}

func JwtSecretBytes() []byte {
	return []byte(jwt_secret)
}
