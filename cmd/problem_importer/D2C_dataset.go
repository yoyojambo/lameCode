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

	"slices"
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

func walkReceiver(in chan challenge, done chan<- error, q *data.Queries) {
	defer close(done)

	i := 0
	for ch := range in {
		err := insertEntry(ch, q)
		if err != nil {
			done <- err
			return
		}
		i += 1
		if i == 1 || i%100 == 0 {
			log.Println("Imported challenge", ch.Name)
			log.Println(i, "problems imported...")
		}
	}

	
	log.Println(i, "problems imported in total")
	done <- nil
}

var subDirectories []string = []string{"description", "samples", "solutions_c++", "solutions_python"}

// walkFunc walks a folder with the structure of description2code
// and inserts it to the database linked to q.
func walkParser(out chan<- challenge, done <-chan error) filepath.WalkFunc {
	var ch *challenge = nil
	return func(path string, d fs.FileInfo, walkError error) error {
		if d.IsDir() {
			return nil
		}
		if walkError != nil {
			return walkError
		}
		
		parts := strings.Split(filepath.ToSlash(path), "/")
		subDir := parts[len(parts)-2]

		// Don't read content of file if would not fit the structure
		// Skip to next file
		if _, found := slices.BinarySearch(subDirectories, subDir); !found {
			return nil
		}

		ch_name := parts[len(parts)-3] // eg. [two_sum]/description/description.txt
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Resetting ch is handled with each "description.txt" parsed
		// If it is nil, its probably the first challenge object parsed.
		// If not nil, assume that is a past object, and sends it off.
		switch subDir {
		case "description":
			// If there is no description.txt but a
			// description_annotated.txt is found, use that.
			if d.Name() == "description.txt" || d.Name() == "description_annotated.txt" {
				if ch != nil && ch_name != ch.Name { // Ch has not been overriden
					select {
					case out <- *ch: // Send ch
					case err := <-done:
						return fmt.Errorf("Receiver failed: %v", err)
					}
					ch = nil   // Un-set ch
				}
				if ch == nil { // Reset ch (or create for the first time)
					ch = &challenge{Name: ch_name, Tests: make(map[int]test)}
				}

				// If description is not already filled out ()
				if ch.Description == "" {
					ch.Description = string(content)
				}
			}
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
	done := make(chan error, 1)

	// Initialize receiver
	go walkReceiver(c, done, q)

	// Start generator
	err := filepath.Walk(path, walkParser(c, done))
	if err != nil {
		return err
	}
	close(c)

	// Wait until receiver is done
	err = <-done
	if err != nil {
		return err
	}

	return nil
}
