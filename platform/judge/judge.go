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

type LanguageOption struct {
	Lang string
	PrettyName string
	Compiler string
}

var checkCompilersOnce sync.Once
var languageOptions []LanguageOption

// TODO: Add compiler version to Compiler field
// It only runs once anyway
func checkCompilers() {
	// Check for Rust compiler
	if _, err := exec.LookPath("rustc"); err == nil {
		// "rust" instead of "rs" for the monaco editor
		languageOptions = append(languageOptions, LanguageOption{"rust", "Rust", "rustc"})
	}

	// Check for either go compiler
	if _, err := exec.LookPath("go"); err == nil {
		languageOptions = append(languageOptions, LanguageOption{"go", "Go", "go"})
	} else if _, err := exec.LookPath("tinygo"); err == nil {
		languageOptions = append(languageOptions, LanguageOption{"go", "Go", "tinygo"})
	}

	// Check for emscripten (includes clang)
	if _, err := exec.LookPath("emcc"); err == nil {
		languageOptions = append(languageOptions, LanguageOption{"c", "C", "emcc"})
	}
	// check em++ separately cause you never know what's up with the host system
	if _, err := exec.LookPath("em++"); err == nil {
		languageOptions = append(languageOptions, LanguageOption{"cpp", "C++", "em++"})
	}

	for _, option := range languageOptions {
		l.Printf("Found %s for %s\n", option.Compiler, option.PrettyName)
	}
}

func LanguageOptions() []LanguageOption {
	checkCompilersOnce.Do(checkCompilers)
	return languageOptions
}

// callGoCompiler builds Go programs using TinyGo (WASI target), falls back to standard Go
func callGoCompiler(program string) (string, error) {
	if !strings.HasSuffix(program, ".go") {
		return "", fmt.Errorf("expected .go, got %s", program)
	}
	exe := filepath.Base(strings.TrimSuffix(program, ".go")) + ".wasm"

	compiler, err := exec.LookPath("tinygo")
	if err == nil {
		// TinyGo for WASI
		cmd := exec.Command(compiler, "build", "-o", exe, "-target", "wasi", program)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("tinygo build failed: %v\n%s", err, out)
		}
		return exe, nil
	}
	// Fallback to standard Go (js/wasm)
	compiler, err = exec.LookPath("go")
	if err != nil {
		return "", err
	}

	cmd := exec.Command(compiler, "build", "-o", exe, program)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("go build failed: %v\n%s", err, out)
	}
	return exe, nil
}

// callRustCompiler builds Rust programs targeting wasm32-wasi
func callRustCompiler(program string) (string, error) {
	if !strings.HasSuffix(program, ".rust") {
		return "", fmt.Errorf("expected .rust, got %s", program)
	}
	exe := filepath.Base(strings.TrimSuffix(program, ".rust")) + ".wasm"
	compiler, err := exec.LookPath("rustc")
	if err != nil {
		return "", err
	}
	cmd := exec.Command(compiler, "--target=wasm32-wasip1", "-O", "-o", exe, program)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("rustc failed: %v\n%s", err, out)
	}
	return exe, nil
}

// callGccCompiler uses emcc to compile C to standalone WASM
func callGccCompiler(program string) (string, error) {
	if !strings.HasSuffix(program, ".c") {
		return "", fmt.Errorf("expected .c, got %s", program)
	}
	exe := filepath.Base(strings.TrimSuffix(program, ".c")) + ".wasm"
	emcc, err := exec.LookPath("emcc")
	if err != nil {
		return "", err
	}
	cmd := exec.Command(emcc, program, "-o", exe, "-Oz", "-sSTANDALONE_WASM")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("emcc failed: %v\n%s", err, out)
	}
	return exe, nil
}

// callGppCompiler uses em++ to compile C++ to standalone WASM
func callGppCompiler(program string) (string, error) {
	if !strings.HasSuffix(program, ".cpp") {
		return "", fmt.Errorf("expected .cpp, got %s", program)
	}
	exe := filepath.Base(strings.TrimSuffix(program, ".cpp")) + ".wasm"
	empp, err := exec.LookPath("em++")
	if err != nil {
		return "", err
	}
	cmd := exec.Command(empp, program, "-o", exe, "-Oz", "-sSTANDALONE_WASM")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("em++ failed: %v\n%s", err, out)
	}
	return exe, nil
}

const installWasmer_cmd = "curl https://get.wasmer.io -sSfL | sh"
const wasmerDir = "/tmp/wasmer"

var installWasmerOnce sync.Once
// goroutine to install wasmer
func installWasmer() {
	cmd := exec.Command("sh", "-c", installWasmer_cmd)
	// Don't install in $HOME, but /tmp
	cmd.Env = append(cmd.Env, "WASMER_DIR=" + wasmerDir, "WASMER_INSTALL_LOG=quiet")
	out, err := cmd.CombinedOutput()
	if err != nil {
		l.Printf("Error running wasmer installer: %v\n", err)
	}
	l.Println("Wasmer installer output:\n", string(out))
}

// Checks what (if any) wasm runtime is present in the system,
// and creates the appropiate command.
// TODO: Maybe do this proactively? Why wait for first submission?
func resolveWasmRuntime(executable string) (*exec.Cmd, error) {
	// Check wasmtime is there
	runtime, err := exec.LookPath("wasmtime")
	if err == nil {
		return exec.Command(runtime, executable), nil
	} else if config.Debug() {
		l.Printf("wasmtime runtime not found: %v\n", err)		
	}

	// Check if wasmer is there
	runtime, err = exec.LookPath("wasmer")
	if err == nil {
		return exec.Command(runtime, "run", executable), nil
	} else if config.Debug() {
		l.Printf("wasmer runtime not found: %v\n", err)		
	}

	// Check for self-installed wasmer
	// Or begin installation
	if config.InstallWasmer() {
		fullPath := wasmerDir + "/bin/wasmer"
		_, err := exec.LookPath(fullPath)
		if err == nil { // Already installed
			l.Println("Using self-installed wasmer")
			return exec.Command(fullPath, "run", executable), nil
		} else {
			// Only ever call once the installing of wasmer
			installWasmerOnce.Do(func() { go installWasmer() })
			return nil, fmt.Errorf("Installing wasmer")
		}
	}

	// No WASM runtime found
	return nil, fmt.Errorf("No WASM runtime found!")
}

// RunProgramWithInput exec's executable, and passes input as its
// standard input. It returns it's output as a string and error in trying to run.
// It does not filter out exit code errors, so that can be returned.
func RunWasmProgramWithInput(executable, input string) (string, error) {
	// Get command depending of available wasm runtime
	cmd, err := resolveWasmRuntime(executable)
	if err != nil {
		l.Println("WASM runtime not resolved")
		return "", err
	}

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
// compileProgram selects the right compiler based on lang
func compileProgram(code, lang string) (string, error) {
	switch lang {
	case "go":
		return callGoCompiler(code)
	case "c":
		return callGccCompiler(code)
	case "cpp":
		return callGppCompiler(code)
	case "rust": // The frontend uses "rust" instead if "rs" for the monaco editor
		return callRustCompiler(code)
	default:
		return "", fmt.Errorf("unsupported language: %s", lang)
	}
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
		// Actual error message from compiler, don't add newline
		return nil, fmt.Errorf("Error compiling %s program:\n%v", lang, err)
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
