package main

import (
	"context"
	"fmt"
	"io/fs"
	"lameCode/platform/data"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type challenge struct {
	Name        string
	Description string
	Tests       map[int]test
	solutions   []solution
}

type solution struct {
	Language string
	Content  string
}

type test struct {
	In  string
	Out string
}

func insertEntry(ch challenge, q *data.Queries) error {
	ctx := context.Background()

	// INSERT Challenge row
	challenge_id, err := q.NewChallenge(ctx, ch.Name, ch.Description, 0)
	if err != nil {
		return err
	}

	// INSERT its respective tests
	for _, test := range ch.Tests {
		_, err := q.NewChallengeTest(ctx, challenge_id, test.In, test.Out)
		if err != nil {
			return err
		}
	}

	return nil
}

// walkFunc walks a folder with the structure of description2code
// and inserts it to the database linked to q.
func walkParser(c chan challenge) filepath.WalkFunc {
	var ch *challenge = nil
	return func(path string, d fs.FileInfo, _ error) error {
		if d.IsDir() {
			return nil
		}
		parts := strings.Split(filepath.ToSlash(path), "/")
		ch_name := parts[len(parts)-3] // eg. [two_sum]/description/description.txt
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Resetting ch is handled with each "description.txt" parsed
		// If it is nil, its probably the first challenge object parsed.
		// If not nil, assume that is a past object, and sends it off.
		switch subDir := parts[len(parts)-2]; subDir {
		case "description":
			if d.Name() == "description.txt" {
				if ch != nil {
					c <- *ch // Send ch
					ch = nil // Un-set ch
				}
				if ch == nil { // Reset ch (or create for the first time)
					ch = &challenge{Name: ch_name, Tests: make(map[int]test)}
				}

				ch.Description = string(content)
			}
			// TODO: if there is no description.txt but a
			// description_annotated.txt is found, use that.
		case "samples":
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
		case "solutions_c++", "solutions_python":
			// TODO: Add this solutions as an example user?
			// (as an opt-in option)
			return filepath.SkipDir
		}

		return nil
	}
}

func import_Description2Code(path string, q *data.Queries) error {
	c := make(chan challenge, 3)

	// Initialize receiver
	go func() {
		i := 0
		for ch := range c {
			err := insertEntry(ch, q)
			if err != nil {
				log.Fatalln("Async insertEntry error:", err)
			}
			i += 1
			if i == 1 && i % 100 == 0 {
				log.Println("Imported challenge", ch.Name)
				log.Println(i, "problems imported...")
			}
		}
	}()

	// Start generator
	err := filepath.Walk(path, walkParser(c))
	if err != nil {
		return err
	}

	return nil
}
