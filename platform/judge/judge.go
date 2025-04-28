package judge

import (
	"fmt"
	"lameCode/platform/config"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var l = log.New(os.Stdout, "[judge] ", log.LstdFlags | log.Lmsgprefix)

// checkWasmtime checks if the wasmtime runtime can be found in the current environment
func lookPath(executable string) (string, error) {
	path, err := exec.LookPath(executable)
	return path, err
}

// Using sync.OnceValues to basically cache the result, but might be
// overkill or even slower, idk.
var getGoCompiler func() (string, error) = sync.OnceValues(
	func() (string, error) {
		return exec.LookPath("go")
	})

var getGccCompiler func() (string, error) = sync.OnceValues(
	func() (string, error) {
		return exec.LookPath("gcc")
	})

func callGoCompiler(program string) (executable string, err error) {
	if b, found := strings.CutSuffix(program, ".go"); found {
		executable = filepath.Base(b)
		if runtime.GOOS == "windows" {
			executable = executable + ".exe"
		}
	} else {
		err = fmt.Errorf("Expected program to end in .go, but was given %s", program)
	}
	
	compiler, err := getGoCompiler()
	if err != nil {
		return "", err
	}
	
	command_str := fmt.Sprintf("%s build -o %s %s", compiler, executable, program)
	command := strings.Split(command_str, " ")
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = "."
	cmd.Env = append(cmd.Env, "GOOS=wasip1")
	cmd.Env = append(cmd.Env, "GOARCH=wasm")

	// Apparently GOCACHE is necessary for go build, so here goes it.
	// It at least does seem capable of creating the directory itself.
	cmd.Env = append(cmd.Env, "GOCACHE=/tmp/gocache")

	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("Error running compiler:\nErr=%v\nOutput=%v", err, string(out))
	}

	if config.Debug() {
		l.Println("Ran compiler with", command_str)
		if len(out) != 0 {
			l.Printf("Compiler output: %v\n", string(out))
		}
		l.Println("Exited compiler with", cmd.ProcessState.ExitCode())
	}

	return
}

// RunProgramWithInput exec's executable, and passes input as its
// standard input. It returns it's output as a string and error in trying to run.
// It does not filter out exit code errors, so that can be returned.
func RunWasmProgramWithInput(executable, input string) (string, error) {
	cmd := exec.Command("wasmtime", executable)
	cmd.Stdin = strings.NewReader(input)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Error running submission: %v\n", err)
	}

	return string(out), nil
}

type Result struct {
	Name *string // Optional
	Pass bool
	Runtime int32 // In milisecconds
	MaxMemUsed int32 // In Bytes
}

// createProgram creates the temporary file with the code of the
// submission.
//
// It returns the name of the file, and any error value
// in creating or writing the file.  lang is the file extension and
// code is the string of the code submitted.
func createProgram(lang, code string) (string, error) {
	// Create and write temporary file for the code submission
	ext := lang
	f, err := os.CreateTemp("", "go_submission_*." + ext)
	if err != nil {
		return "", fmt.Errorf("Error creating submission code file: %v\n", err)
	}
	prog_name := f.Name()
	defer f.Close()

	_, err = f.WriteString(code)
	if err != nil {
		return prog_name, fmt.Errorf("Error writing to submission code file: %v\n", err)
	}

	return prog_name, nil
}

// compileProgram runs the appropiate scripts to create the
// executable.  Returns the name of the executable (if created) and
// errors encountered running the compiler.
func compileProgram(prog, lang string) (string, error) {
	switch lang {
	case "go":
		return callGoCompiler(prog)
	default:
		return "", fmt.Errorf("Language compiler not implemented/recognized")
	}
}

// TODO: Multiple tests per compiled program
func runGoProgram(code string, input string) (string, error) {
	prog_name, err := createProgram("go", code)
	if err != nil {
		return "", fmt.Errorf("Error creating program:\n%v\n", err)
	}
	// Delete after all tests are done (or failed to run)
	defer os.Remove(prog_name)
	
	executable, err := callGoCompiler(prog_name)
	if err != nil {
		return "", fmt.Errorf("Error compiling: %v\n", err)
	}

	defer os.Remove(executable)

	return RunWasmProgramWithInput(executable, input)
}

