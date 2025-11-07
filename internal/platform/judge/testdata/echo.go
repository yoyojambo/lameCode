package main

import (
	"io"
	"os"
)

func main() {
	stdin, _ := io.ReadAll(os.Stdin)

	os.Stderr.WriteString(string(stdin))
}
