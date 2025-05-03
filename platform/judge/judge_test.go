package judge

import (
	_ "embed"
	"lameCode/platform/data"
	"log"
	"testing"
)

//go:embed testdata/hello_world.go
var hello_world_GO string

//go:embed testdata/hello_world.c
var hello_world_C string

//go:embed testdata/hello_world.cpp
var hello_world_CPP string

//go:embed testdata/readfile.go
var readfile_code string

//go:embed testdata/echo.go
var echo_code string

func TestRunGoHelloWorldProgram(t *testing.T) {
	out, err := runGoProgram(hello_world_GO, "")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("hello_world.go output:", out)
}

func TestRunGoEchoProgram(t *testing.T) {
	in := "Text directly to its standard input.\n"
	out, err := runGoProgram(echo_code, in)
	if err != nil {
		t.Fatal(err)
	}

	if out != in {
		log.Fatalf("Expected: \"%s\"\nGot: \"%s\"\n", in, out)
	}
}

func TestRunMultipleTests_HelloWorld(t *testing.T) {
	var hello_world_tests []data.ChallengeTest = []data.ChallengeTest{
		{ExpectedOutput: "Hello world!"},
		{ExpectedOutput: "Hello world!", InputData: "asdascabs"},
	}

	results, err := RunMultipleTests(hello_world_GO, "go", hello_world_tests)
	if err != nil {
		t.Error(err)
	}
	
	for _, r := range results {
		if !r.Pass {
			t.Error("Failed in GO hello world!")
		}
	}

	results, err = RunMultipleTests(hello_world_C, "c", hello_world_tests)
	if err != nil {
		t.Error(err)
	}
	
	for _, r := range results {
		if !r.Pass {
			t.Error("Failed in C hello world!")
		}
	}

	results, err = RunMultipleTests(hello_world_CPP, "cpp", hello_world_tests)
	if err != nil {
		t.Error(err)
	}
	
	for _, r := range results {
		if !r.Pass {
			t.Error("Failed in CPP hello world!")
		}
	}
}
