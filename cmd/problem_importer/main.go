package main

import (
	"database/sql"
	"fmt"
	"os"

	"lameCode/platform/data"
	"log"
	"flag"

	_ "modernc.org/sqlite"
)

var dry_run = flag.Bool("dry-run", false, "Parse and load in a in-memory ephemeral database the file.")

func run() error {
	if len(os.Args) == 1 {
		return fmt.Errorf("No csv file given to read")
	}

	

	if len(os.Args) > 3 {
		return fmt.Errorf("Too many arguments")
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return err
	}

	db.Ping()

	fmt.Println(os.Args[1])

	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}

	data.DB().Ping()
}
