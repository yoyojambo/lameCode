package main

import (
	"database/sql"
	"lameCode/platform/data"
	"fmt"
	"os"
	"path"
	"testing"
	_ "modernc.org/sqlite"
)

func init() {
	cur, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Kinda fragile, to be honest.
	// Probably should do something like
	// https://stackoverflow.com/a/29541248
	traversal := "../../testdata"
	err = os.Chdir(path.Join(cur, traversal))
	if err != nil {
		panic(err)
	}

	fmt.Printf("[D2C_dataset] Changed CWD from %s to %s\n", cur, path.Clean(path.Join(cur, traversal)))
}

func TestD2C_import(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping description2code test in short mode")
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	err = data.LoadSchema(db)
	defer db.Close()

	queries := data.New(db)
	err = import_Description2Code("hackerearth", queries)
	if err != nil {
		t.Fatal(err)
	}
}
