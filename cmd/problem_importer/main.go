package main

import (
	"database/sql"
	"fmt"
	"os"

	"flag"
	"lameCode/platform/data"
	"log"

	"context"
	_ "modernc.org/sqlite"
)

var dry_run = flag.Bool("dry-run", false, "Parse and load in a in-memory ephemeral database the file.")
var desc2code = flag.Bool("description2code", true, "Parse a folder with the structure of the description2code dataset")

var create_db = flag.Bool("create", false, "Load schema on database upon connection. Will show and error if schema was already present")

func run() error {
	ctx := context.Background()
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
	
	db, err := sql.Open("sqlite", db_fname)
	if err != nil {
		return err
	}

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

	defer db.Close()
	if *desc2code {
		err := import_Description2Code(target_fname, q)
		if err != nil {
			return err
		}
	} else {
		r, err := os.Open(target_fname)
		if err != nil {
			return err
		}

		problems := import_CsvDataset(r)

		for i, p := range problems {
			if i < 10 {
				log.Printf("Challenge #%d:\n  TITLE: %s\n  DESC: %s\n  DIFF: %d\n",
					i+1, p.Title, p.Description[:20], p.Difficulty)
			}
			_, err := q.NewChallenge(ctx, p.Title, p.Description, int64(p.Difficulty))
			if err != nil {
				return err
			}
		}
		fmt.Println("Inserted ", len(problems), " problems")
	}

	return nil
}

func main() {
	flag.Parse()
	if *dry_run {
		log.Printf("Running a dry-run on %s\n", flag.Arg(0))
	}
	if err := run(); err != nil {
		log.Fatalln(err)
	}

	log.Println("Imported succesfully")
}
