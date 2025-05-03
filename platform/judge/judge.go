package judge

import (
	"fmt"
	"lameCode/platform/config"
	"lameCode/platform/data"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var l = log.New(os.Stdout, "[judge] ", log.LstdFlags|log.Lmsgprefix)

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

var getGppCompiler func() (string, error) = sync.OnceValues(
	func() (string, error) {
		return exec.LookPath("g++")
	})

var getWasmtime func() (string, error) = sync.OnceValues(
	func() (string, error) {
		return exec.LookPath("wasmtime")
	})

func callGoCompiler(program string) (executable string, err error) {
	if b, found := strings.CutSuffix(program, ".go"); found {
		executable = filepath.Base(b) + ".wasm"
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

// callGccCompiler compiles a C program to WASM using Emscripten (assuming it's in the PATH).
func callGccCompiler(program string) (executable string, err error) {
	if b, found := strings.CutSuffix(program, ".c"); found {
		executable = filepath.Base(b) + ".wasm"
	} else {
		return "", fmt.Errorf("Expected program to end in .c, but was given %s", program)
	}

	compiler, err := getGccCompiler()
	if err != nil {
		return "", fmt.Errorf("gcc compiler not found: %w", err)
	}

	// Assuming you have Emscripten's gcc wrapper in your PATH, which is common.
	// Emscripten will handle the WASM compilation.
	command := []string{
		compiler,
		program,
		"-o", executable,
		// Add Emscripten specific flags for WASI if needed, e.g., -s WASI=1
		// For a basic example, the default might work, but for more complex
		// C programs, you might need specific Emscripten flags.
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = "."
	// Emscripten might need specific environment variables set, depending on
	// your Emscripten installation and configuration.

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Error running GCC compiler:\nErr=%v\nOutput=%v", err, string(out))
	}

	if config.Debug() {
		l.Printf("Ran GCC compiler with %v\n", command)
		if len(out) != 0 {
			l.Printf("GCC compiler output: %v\n", string(out))
		}
	}

	return executable, nil
}

// callGppCompiler compiles a C++ program to WASM using Emscripten (assuming it's in the PATH).
func callGppCompiler(program string) (executable string, err error) {
	if b, found := strings.CutSuffix(program, ".cpp"); found {
		executable = filepath.Base(b) + ".wasm"
	} else {
		return "", fmt.Errorf("Expected program to end in .cpp, but was given %s", program)
	}

	compiler, err := getGppCompiler()
	if err != nil {
		return "", fmt.Errorf("g++ compiler not found: %w", err)
	}

	// Assuming you have Emscripten's g++ wrapper in your PATH.
	command := []string{
		compiler,
		program,
		"-o", executable,
		// Add Emscripten specific flags for WASI if needed, e.g., -s WASI=1
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = "."
	// Emscripten might need specific environment variables.

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Error running G++ compiler:\nErr=%v\nOutput=%v", err, string(out))
	}

	if config.Debug() {
		l.Printf("Ran G++ compiler with %v\n", command)
		if len(out) != 0 {
			l.Printf("G++ compiler output: %v\n", string(out))
		}
	}

	return executable, nil
}

// RunProgramWithInput exec's executable, and passes input as its
// standard input. It returns it's output as a string and error in trying to run.
// It does not filter out exit code errors, so that can be returned.
func RunWasmProgramWithInput(executable, input string) (string, error) {
	// Check wasmtime is there
	wasmtimePath, err := getWasmtime()
	if err != nil {
		return "", fmt.Errorf("wasmtime runtime not found: %w", err)
	}

	cmd := exec.Command(wasmtimePath, executable)
	cmd.Stdin = strings.NewReader(input)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Error running submission: %v\n", err)
	}

	return string(out), nil
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
	// Use a more specific pattern for temp files
	f, err := os.CreateTemp("", fmt.Sprintf("submission_*.%s", ext))
	if err != nil {
		return "", fmt.Errorf("Error creating submission code file: %v\n", err)
	}
	prog_name := f.Name()
	defer f.Close() // Close the file handle after creating and writing

	_, err = f.WriteString(code)
	if err != nil {
		// Return the program name even if writing fails for potential cleanup
		return prog_name, fmt.Errorf("Error writing to submission code file %s: %v\n", prog_name, err)
	}

	return prog_name, nil
}

// compileProgram runs the appropiate scripts to create the
// executable.  Returns the name of the executable (if created) and
// errors encountered running the compiler.
func compileProgram(code, lang string) (string, error) {
	switch lang {
	case "go":
		return callGoCompiler(code)
	case "c":
		return callGccCompiler(code)
	case "cpp":
		return callGppCompiler(code)
	default:
		return "", fmt.Errorf("Language compiler not implemented/recognized for language: %s", lang)
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

// runCProgram compiles and runs a C program.
func runCProgram(code string, input string) (string, error) {
	prog_name, err := createProgram("c", code)
	if err != nil {
		return "", fmt.Errorf("Error creating C program file:\n%v\n", err)
	}
	defer os.Remove(prog_name)

	executable, err := compileProgram(prog_name, "c")
	if err != nil {
		return "", fmt.Errorf("Error compiling C program: %v\n", err)
	}
	defer os.Remove(executable)

	return RunWasmProgramWithInput(executable, input)
}

// runCppProgram compiles and runs a C++ program.
func runCppProgram(code string, input string) (string, error) {
	prog_name, err := createProgram("cpp", code)
	if err != nil {
		return "", fmt.Errorf("Error creating C++ program file:\n%v\n", err)
	}
	defer os.Remove(prog_name)

	executable, err := compileProgram(prog_name, "cpp")
	if err != nil {
		return "", fmt.Errorf("Error compiling C++ program: %v\n", err)
	}
	defer os.Remove(executable)

	return RunWasmProgramWithInput(executable, input)
}

type Result struct {
	Name       string // Optional
	Pass       bool
	Runtime    int32 // In milisecconds
	MaxMemUsed int32 // In Bytes
}

func RunMultipleTests(code, lang string, challenges []data.ChallengeTest) ([]Result, error) {
	prog_name, err := createProgram(lang, code)
	if err != nil {
		return nil, fmt.Errorf("Error creating %s program file:\n%v\n", lang, err)
	}
	defer os.Remove(prog_name)

	executable, err := compileProgram(prog_name, lang)
	if err != nil {
		return nil, fmt.Errorf("Error compiling %s program: %v\n", lang, err)
	}
	defer os.Remove(executable)

	results := make([]Result, 0, len(challenges))
	for i, c := range challenges {
		in := c.InputData
		out, err := RunWasmProgramWithInput(executable, in)
		if err != nil {
			return results, fmt.Errorf("Error running test #%d: %v", i, err)
		}

		out_s, expected_s := strings.TrimSpace(out), strings.TrimSpace(c.ExpectedOutput)
		pass := out_s == expected_s

		// Express differences that fail the test
		if config.Debug() && !pass {
			got_b := []byte(out_s)
			expected_b := []byte(expected_s)

			mismatches := 0
			for j := 0; j < min(len(got_b), len(expected_b)); j++ {
				if mismatches > 4 {
					break
				}
				if got_b[j] != expected_b[j] {
					mismatches++
					l.Printf("Mismatch byte #%d in test #%d\nGot %q expected %q\n",
						j, i, got_b[j], expected_b[j])
				}
			}
		}

		r := Result{
			Name: "Test #" + strconv.Itoa(i),
			Pass: pass,
		}
		results = append(results, r)
	}

	return results, nil
}
