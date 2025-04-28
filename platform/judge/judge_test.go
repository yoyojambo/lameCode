package judge

import (
	_ "embed"
	"log"
	"testing"
)

//go:embed testdata/hello_world.go
var hello_world_code string

//go:embed testdata/readfile.go
var readfile_code string

//go:embed testdata/echo.go
var echo_code string

func TestRunGoHelloWorldProgram(t *testing.T)  {
	out, err := runGoProgram(hello_world_code, "")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("hello_world.go output:", out)
}

func TestRunGoEchoProgram(t *testing.T)  {
	in := "Text directly to its standard input.\n"
	out, err := runGoProgram(echo_code, in)
	if err != nil {
		t.Fatal(err)
	}

	if out != in {
		log.Fatalf("Expected: \"%s\"\nGot: \"%s\"\n", in, out)
	}
}
