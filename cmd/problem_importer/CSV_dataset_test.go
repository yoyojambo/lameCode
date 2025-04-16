package main

import (
	"encoding/csv"
	"os"
	"testing"
)

func TestCsvParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CSV parsing test in short mode")
	}

	f, err := os.Open("leetcode_dataset.csv")
	if err != nil {
		t.Fatal(err)
	}
	csv_r := csv.NewReader(f)

	values, err := csv_r.ReadAll()
	if err != nil {
		t.Error(err)
	}

	problems := parseProblems(values[1:])

	for i := 0; i < 10; i++ {
		t.Log(problems[i].String())
	}
}

