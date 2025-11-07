package data

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)


func TestSchemaStatements(t *testing.T) {
	statements := GetSchemaStatements()

	for i, s := range statements {
		t.Logf("[%d] \n%s\n", i, s)
	}
}

func TestLoadSchema(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	err = LoadSchema(db)
	if err != nil {
		t.Fatal(err)
	}
}
