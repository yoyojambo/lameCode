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

// LanguageOption is what is used in the template for a problem, to
// only advertise available languages.
type LanguageOption struct {
	Lang       string
	PrettyName string
	Compiler   string
}

// CompilerConfig abstracts the calling of a compiler for any language.
// It indicates
//   - The file extension of the code
//   - The extension of the output
//     (currently only wasm until I figure out more ways of isolating code execution)
//   - The environment variables to pass the command
//   - The format to call the compiler, inserting the executable output name and the code
type CompilerConfig struct {
	srcExt, exeExt, cmdName string
	// buildArgs is called per-compile to customize flags (e.g. different targets).
	buildArgs func(src, exe string) []string
	// environ can set GOOS, WASM flags, etc.
	environ []string
}

// Immutable essentially
func (cfg CompilerConfig) SourceExt() string { return cfg.srcExt }
func (cfg CompilerConfig) ExecExt() string   { return cfg.exeExt }
func (cfg CompilerConfig) CmdAndArgs(src, exe string) (string, []string, []string) {
	return cfg.cmdName, cfg.buildArgs(src, exe), cfg.environ
}

func (c CompilerConfig) Compile(sourcePath string) (string, error) {
	if !strings.HasSuffix(sourcePath, "."+c.SourceExt()) {
		return "", fmt.Errorf("expected .%s, got %s", c.SourceExt(), sourcePath)
	}

	base := filepath.Base(strings.TrimSuffix(sourcePath, "."+c.SourceExt()))
	exe := base + "." + c.ExecExt()

	cmdName, args, extraEnv := c.CmdAndArgs(sourcePath, exe)
	cmd := exec.Command(cmdName, args...)
	cmd.Env = append(os.Environ(), extraEnv...)

	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%s failed: %v\n%s", cmdName, err, out)
	}
	return exe, nil
}

var languageOptions []LanguageOption = make([]LanguageOption, 0, 5)
var compilers []CompilerConfig

var checkCompilersOnce sync.Once

func LanguageOptions() []LanguageOption {
	checkCompilersOnce.Do(checkCompilers)
	return languageOptions
}

// TODO: Add compiler version to Compiler field
// It only runs once anyway
func checkCompilers() {
	// Normal rust compiler
	if path, err := exec.LookPath("rustc"); err == nil {
		l.Println("Found rustc")
		compilers = append(compilers, CompilerConfig{
			srcExt:  "rust",
			exeExt:  "wasm",
			cmdName: path,
			buildArgs: func(src, exe string) []string {
				return []string{"--target=wasm32-wasip1", "-O", "-o", exe, src}
			},
			environ: []string{},
		})

		cmd := exec.Command(path, "--version")
		out, err := cmd.Output()
		if err != nil {
			l.Printf("Failed running rustc to get version: %v", err)
		}

		out_split := strings.Split(string(out), " ")
		if len(out_split) > 1 {
			languageOptions = append(languageOptions, LanguageOption{"rust", "Rust", strings.Join(out_split[:2], " ")})
		} else {
			languageOptions = append(languageOptions, LanguageOption{"rust", "Rust", "rustc"})
		}
	}

	// TinyGo (preferred) or Go fallback
	if tg, err := exec.LookPath("tinygo"); err == nil {
		l.Println("Found tinygo")
		compilers = append(compilers, CompilerConfig{
			srcExt:  "go",
			exeExt:  "wasm",
			cmdName: tg,
			buildArgs: func(src, exe string) []string {
				return []string{"build", "-o", exe, "-target", "wasi", src}
			},
			// environ can be left zero-valued if there is no env variables to set
			//environ: []string{},
		})
		languageOptions = append(languageOptions, LanguageOption{"go", "Go", "tinygo"})
	} else if gopath, err := exec.LookPath("go"); err == nil {
		l.Println("Found go")
		compilers = append(compilers, CompilerConfig{
			srcExt:  "go",
			exeExt:  "wasm",
			cmdName: gopath,
			buildArgs: func(src, exe string) []string {
				return []string{"build", "-o", exe, src}
			},
			environ: []string{"GOOS=wasip1", "GOARCH=wasm"},
		})
		
		cmd := exec.Command(gopath, "version")
		out, err := cmd.Output()
		if err != nil {
			l.Printf("Failed running rustc to get version: %v", err)
		}

		out_split := strings.Split(string(out), " ")
		if len(out_split) > 2 {
			languageOptions = append(languageOptions, LanguageOption{"go", "Go", out_split[2]})
		} else {
			languageOptions = append(languageOptions, LanguageOption{"go", "Go", "go"})
		}
	}

	// Emscripten for C and C++
	if emcc, err := exec.LookPath("emcc"); err == nil {
		l.Println("Found emcc")
		compilers = append(compilers, CompilerConfig{
			srcExt:  "c",
			exeExt:  "wasm",
			cmdName: emcc,
			buildArgs: func(src, exe string) []string {
				return []string{src, "-o", exe, "-Oz", "-sSTANDALONE_WASM"}
			},
		})
		languageOptions = append(languageOptions, LanguageOption{"c", "C", "emcc"})
	}
	if empp, err := exec.LookPath("em++"); err == nil {
		l.Println("Found em++")
		compilers = append(compilers, CompilerConfig{
			srcExt:  "cpp",
			exeExt:  "wasm",
			cmdName: empp,
			buildArgs: func(src, exe string) []string {
				return []string{src, "-o", exe, "-Oz", "-sSTANDALONE_WASM"}
			},
		})
	}

	l.Println("Loaded all available compilers")
}

const installWasmer_cmd = "curl https://get.wasmer.io -sSfL | sh"
const wasmerDir = "/tmp/wasmer"

var installWasmerOnce sync.Once

// goroutine to install wasmer
func installWasmer() {
	cmd := exec.Command("sh", "-c", installWasmer_cmd)
	// Don't install in $HOME, but /tmp, and don't crate full verbose logs
	cmd.Env = append(cmd.Env, "WASMER_DIR="+wasmerDir, "WASMER_INSTALL_LOG=quiet")
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
	}

	// Check if iwasm is there
	// this is the one that is included with the Docker image
	runtime, err = exec.LookPath("iwasm")
	if err == nil {
		return exec.Command(runtime, executable), nil
	} else if config.Debug() {
		l.Printf("iwasm runtime not found: %v\n", err)
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
			installWasmerOnce.Do(func() {
				go installWasmer()
				l.Println("Running wasmer install script")
			})
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
func compileProgram(file, lang string) (string, error) {
	for _, c := range compilers {
		if c.SourceExt() == lang {
			return c.Compile(file)
		}
	}
	return "", fmt.Errorf("unsupported language: %s", lang)
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
