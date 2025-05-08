package judge

import (
	_ "embed"
	"lameCode/platform/data"
	"testing"
)

//go:embed testdata/hello_world.go
var hello_world_GO string

//go:embed testdata/hello_world.c
var hello_world_C string

//go:embed testdata/hello_world.cpp
var hello_world_CPP string

//go:embed testdata/hello_world.rs
var hello_world_RS string

//go:embed testdata/readfile.go
var readfile_code string

//go:embed testdata/echo.go
var echo_code string

func TestRunMultipleTests_HelloWorld_multilang(t *testing.T) {
	var hello_world_tests []data.ChallengeTest = []data.ChallengeTest{
		{ExpectedOutput: "Hello world!"},
		{ExpectedOutput: "Hello world!", InputData: "asdascabs"},
		{ExpectedOutput: "Hello world!", InputData: "123jbasd\n\n\n"},
	}

	results, err := RunMultipleTests(hello_world_GO, "go", hello_world_tests)
	if err != nil {
		t.Error(err)
	}
	
	for _, r := range results {
		if !r.Pass {
			t.Error("Failed in Go hello world!")
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
			t.Error("Failed in C++ hello world!")
		}
	}

	results, err = RunMultipleTests(hello_world_RS, "rs", hello_world_tests)
	if err != nil {
		t.Error(err)
	}
	
	for _, r := range results {
		if !r.Pass {
			t.Error("Failed in Rust hello world!")
		}
	}
}
