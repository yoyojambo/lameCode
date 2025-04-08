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

func run() error {
	ctx := context.Background()
	if flag.NArg() < 2 {
		return fmt.Errorf(
			"Not enough arguments, expected csv file and database file",
		)
	}

	if flag.NArg() > 2 {
		return fmt.Errorf("Too many arguments")
	}

	// File names
	csv_fname, db_fname := flag.Arg(0), flag.Arg(1)

	db, err := sql.Open("sqlite", db_fname)
	if err != nil {
		return err
	}

	defer db.Close()

	r, err := os.Open(csv_fname)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	q := data.New(db).WithTx(tx)

	problems := ParseProblemsFromReader(r)

	for i, p := range problems {
		if i < 10 {
			log.Printf("Challenge #%d:\n  TITLE: %s\n  DESC: %s\n  DIFF: %d",
				i+1, p.Title, p.Description[:20], p.Difficulty)
		}
		_, err := q.NewChallenge(ctx, p.Title, p.Description, int64(p.Difficulty))
		if err != nil {
			return err
		}
	}

	tx.Commit()

	fmt.Println("Inserted ", len(problems), " problems")

	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}

	data.DB().Ping()
}
