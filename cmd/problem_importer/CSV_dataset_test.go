package main

import (
	"fmt"
	"os"
	"path"
	"testing"
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

	fmt.Printf("[CSV_dataset] Changed CWD from %s to %s", cur, path.Clean(path.Join(cur, traversal)))
}

func TestCsvParsing(t *testing.T) {
	f, err := os.Open("leetcode_dataset.csv")
	if err != nil {
		panic(err)
	}
	// The actual test
	problems := import_CsvDataset(f)

	for i := 0; i < 10; i++ {
		fmt.Println(problems[i].String())
	}
}
