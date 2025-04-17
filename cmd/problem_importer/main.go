package main

import (
	"fmt"
	"log"

	"flag"
	"lameCode/platform/data"

	"database/sql"
	_ "modernc.org/sqlite"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var (
	dry_run   = flag.Bool("dry-run", false, "Parse and load in a in-memory ephemeral database the file.")
	desc2code = flag.Bool("description2code", true, "Parse a folder with the structure of the description2code dataset")
	create_db = flag.Bool("create", false, "Load schema on database upon connection. Will show and error if schema was already present")
	remote_db = flag.Bool("remote", false, "Assume the given database is a valid URL to a turso database with auth included")
)

func run() error {
	if !*dry_run && flag.NArg() < 2 || flag.NArg() < 1 {
		return fmt.Errorf(
			"Not enough arguments, expected csv file and database file (unless using --dry-run)",
		)
	}

	if *dry_run && flag.NArg() > 1 || flag.NArg() > 2 {
		return fmt.Errorf("Too many arguments")
	}

	// File names
	target_fname := flag.Arg(0)
	db_fname := ":memory:"
	if !*dry_run {
		db_fname = flag.Arg(1)
	}

	
	sql_driver := "sqlite"
	if *remote_db {
		sql_driver = "libsql"
	}

	db, err := sql.Open(sql_driver, db_fname)
	if err != nil {
		return err
	}

	log.Println("Succesful database connection!")

	// If the connected database is memory-only, load schema.
	// TODO: Check non-memory databases for the schema, and apply
	// if necessary.
	if db_fname == ":memory:" || *create_db {
		err := data.LoadSchema(db)
		if err != nil {
			log.Println("Ignoring error, assuming schema already existed...\n", err)
		}
	}

	q := data.New(db)

	if *desc2code {
		err := import_Description2Code(target_fname, q)
		if err != nil {
			return err
		}
	} else {
		err := import_CsvDataset(target_fname, q)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Parse()
	if *dry_run {
		log.Printf("Running a dry-run on %s\n", flag.Arg(0))
	} else {
		log.Printf("Importing dataset to %s\n", flag.Arg(1))
	}

	if err := run(); err != nil {
		log.Fatalln(err)
	}

	log.Println("Problems imported succesfully")
}
