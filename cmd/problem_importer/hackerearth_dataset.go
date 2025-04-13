package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"lameCode/platform/data"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type challenge struct {
	Name string
	Description string
	Tests map[int]test
	solutions []solution
}

type solution struct {
	Language string
	Content string
}

type test struct {
	In string
	Out string
}

func insertEntry(ch challenge, q *data.Queries) error {
	ctx := context.Background()
	challenge_id, err := q.NewChallenge(ctx, ch.Name, ch.Description, 0)
	if err != nil {
		return err
	}

	for _, test := range ch.Tests {
		_, err := q.NewChallengeTest(ctx, challenge_id, test.In, test.Out)
		if err != nil {
			return err
		}
	}

	return nil
}

// walkFunc walks the folder with the structure of description2code
func walkParser(db *sql.DB) filepath.WalkFunc {
	// Something to process the walked path...
	// could be a goroutine and channels
	// could just be a map (somehow needs to respond to )
	var ch *challenge = nil
	return func (path string, d fs.FileInfo, err error) error {
		if d.IsDir() {
			if ch != nil {
			}
			return nil
		}
		parts := strings.Split(filepath.ToSlash(path), "/")
		ch_name := parts[len(parts)-3]
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if ch == nil {
			// Assert that description is the first file to be walked
			if d.Name() != "description.txt" {
				return fmt.Errorf(
					"Expected to walk directory in order... %s is not description.txt",
					d.Name(),
				)
			}
			ch = &challenge{Name: ch_name, Tests: make(map[int]test)}
		}

		// Verifies/validates subfolder structure, mostly.
		// For the most part, it could be done with only filename
		// eg XX_input, description.txt, YYYYYY.txt
		subDir := parts[len(parts)-2]
		if subDir == "description" {
			if d.Name() == "description.txt" {
				ch.Name = string(content)
			}
		} else if subDir == "samples" {
			if before, f := strings.CutSuffix(d.Name(), "_input.txt"); f {
				n, err := strconv.ParseInt(before, 10, 32)
				if err != nil {
					return fmt.Errorf("%s is not a valid sample file.\n%v",
						d.Name(), err)
				}
				
				t := ch.Tests[int(n)] // get zero value 
				ch.Tests[int(n)] = test{string(content), t.Out}
			} else if before, f := strings.CutSuffix(d.Name(), "_output.txt"); f {
				n, err := strconv.ParseInt(before, 10, 32)
				if err != nil {
					return fmt.Errorf("%s is not a valid sample file.\n%v",
						d.Name(), err)
				}
				
				t := ch.Tests[int(n)] // get zero value 
				ch.Tests[int(n)] = test{t.In, string(content)}
			}
		} else if subDir == "solutions_c++" || subDir == "solutions_python" {
			// Currently doesn't handle provided solutions.
			return filepath.SkipDir
		} else {
			return filepath.SkipDir
		}

		return nil
	}
}

func import_Description2Code(path string, db *sql.DB) error {
	err := filepath.Walk(path, walkParser(db))
	if err != nil {
		return err
	}

	return nil
}
