package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"backend/config"
)

// RunResult contains the result of running a single testcase or the entire unified test suite
type RunResult struct {
	Passed         bool   `json:"passed"`
	ActualOutput   string `json:"actual_output,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
	ErrorMessage   string `json:"error_message,omitempty"`
}

// IsNsjailAvailable checks if nsjail binary exists on system PATH or configured location
func IsNsjailAvailable() bool {
	if _, err := exec.LookPath("nsjail"); err == nil {
		return true
	}
	if _, err := os.Stat("/usr/local/bin/nsjail"); err == nil {
		return true
	}
	return false
}

// RunSubmission manages the secure execution and comparison of student code vs right solution.
// Uses Nsjail engine if available on system, otherwise falls back to hardened Docker container.
func RunSubmission(studentCode string, rightSolution string, test string) RunResult {
	if IsNsjailAvailable() {
		log.Println("Sandbox: Running via Nsjail Engine (High Speed)")
		return runSubmissionNsjail(studentCode, rightSolution, test)
	}

	log.Println("Sandbox: Running via Hardened Docker Engine")
	return runSubmissionDocker(studentCode, rightSolution, test)
}

// runSubmissionNsjail executes code using Nsjail (ultra-fast process isolation)
func runSubmissionNsjail(studentCode string, rightSolution string, test string) RunResult {
	tempDir, err := os.MkdirTemp("", "atom-nsjail-*")
	if err != nil {
		return RunResult{Passed: false, ErrorMessage: fmt.Sprintf("Failed to create temp directory: %v", err)}
	}
	_ = os.Chmod(tempDir, 0777)
	defer os.RemoveAll(tempDir)

	if err := setupWorkspaceFiles(tempDir, studentCode, rightSolution, test); err != nil {
		return RunResult{Passed: false, ErrorMessage: err.Error()}
	}

	// 1. Compile student code
	if errOutput := compileGoCodeNsjail(tempDir, "student/main.go"); errOutput != "" {
		return RunResult{Passed: false, ErrorMessage: "Compilation Error:\n" + errOutput}
	}

	// 2. Run Test Runner using Nsjail
	timeout := time.Duration(config.AppConfig.SandboxTimeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	nsjailPath := "nsjail"
	if _, err := exec.LookPath("nsjail"); err != nil {
		nsjailPath = "/usr/local/bin/nsjail"
	}

	cmd := exec.CommandContext(ctx, nsjailPath,
		"--config", "sandbox/nsjail.cfg",
		"-B", fmt.Sprintf("%s:/app", tempDir),
		"--cwd", "/app",
		"--", "go", "run", "test.go", "compare.go")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return RunResult{Passed: false, ErrorMessage: fmt.Sprintf("Execution timed out (Max %d seconds limit exceeded).", config.AppConfig.SandboxTimeout)}
	}

	return parseRunnerOutput(stdout.String(), stderr.String(), err)
}

// runSubmissionDocker executes code inside a hardened Docker container
func runSubmissionDocker(studentCode string, rightSolution string, test string) RunResult {
	tempDir, err := os.MkdirTemp("", "atom-docker-*")
	if err != nil {
		return RunResult{Passed: false, ErrorMessage: fmt.Sprintf("Failed to create temp directory: %v", err)}
	}
	_ = os.Chmod(tempDir, 0777)
	defer os.RemoveAll(tempDir)

	if err := setupWorkspaceFiles(tempDir, studentCode, rightSolution, test); err != nil {
		return RunResult{Passed: false, ErrorMessage: err.Error()}
	}

	// 1. Compile student code to catch syntax errors
	if errOutput := compileGoCodeDocker(tempDir, "student/main.go"); errOutput != "" {
		return RunResult{Passed: false, ErrorMessage: "Compilation Error:\n" + errOutput}
	}

	// 2. Run the Unified Test Runner inside Container
	timeout := time.Duration(config.AppConfig.SandboxTimeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	memLimit := fmt.Sprintf("%dm", config.AppConfig.SandboxMemoryMB)
	if config.AppConfig.SandboxMemoryMB <= 0 {
		memLimit = "256m"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"--memory="+memLimit,
		"--cpus=0.5",
		"--pids-limit=100",
		"--network", "none",
		"--security-opt=no-new-privileges:true",
		"--cap-drop=ALL",
		"--tmpfs", "/tmp:rw,exec,nosuid,size=64m",
		"-e", "GO111MODULE=on",
		"-v", fmt.Sprintf("%s:/app", tempDir),
		"-w", "/app",
		"golang:1.20-alpine",
		"go", "run", "test.go", "compare.go")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return RunResult{Passed: false, ErrorMessage: fmt.Sprintf("Execution timed out (Max %d seconds limit exceeded).", config.AppConfig.SandboxTimeout)}
	}

	return parseRunnerOutput(stdout.String(), stderr.String(), err)
}

// Helper to write workspace files
func setupWorkspaceFiles(tempDir, studentCode, rightSolution, test string) error {
	studentDir := filepath.Join(tempDir, "student")
	solutionDir := filepath.Join(tempDir, "solution")

	if err := os.MkdirAll(studentDir, 0755); err != nil {
		return fmt.Errorf("failed to create student dir")
	}
	if err := os.MkdirAll(solutionDir, 0755); err != nil {
		return fmt.Errorf("failed to create solution dir")
	}

	testfile := filepath.Join(tempDir, "test.go")
	comparefile := filepath.Join(tempDir, "compare.go")
	studentFile := filepath.Join(studentDir, "main.go")
	solutionFile := filepath.Join(solutionDir, "main.go")

	if err := os.WriteFile(testfile, []byte(test), 0644); err != nil {
		return fmt.Errorf("failed to write test code: %v", err)
	}
	if err := os.WriteFile(studentFile, []byte(studentCode), 0644); err != nil {
		return fmt.Errorf("failed to write student code: %v", err)
	}
	if err := os.WriteFile(solutionFile, []byte(rightSolution), 0644); err != nil {
		return fmt.Errorf("failed to write solution code: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module sandbox\n\ngo 1.20\n"), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod: %v", err)
	}

	compareCode := `package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"os"
	"math/rand"
	"strings"
)

type Result struct {
	Passed         bool   ` + "`" + `json:"passed"` + "`" + `
	ErrorMessage   string ` + "`" + `json:"error_message,omitempty"` + "`" + `
	ExpectedOutput string ` + "`" + `json:"expected_output,omitempty"` + "`" + `
	ActualOutput   string ` + "`" + `json:"actual_output,omitempty"` + "`" + `
}

func formatOutput(val any) string {
	switch v := val.(type) {
	case rune:
		return fmt.Sprintf("%q", v)
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%#v", v)
	}
}

func Compare(expected, actual interface{}, testCase string) {
	if !reflect.DeepEqual(expected, actual) {
		res := Result{
			Passed:         false,
			ErrorMessage:   fmt.Sprintf("Failed on input: %s", testCase),
			ExpectedOutput: formatOutput(expected),
			ActualOutput:   formatOutput(actual),
		}
		b, _ := json.Marshal(res)
		fmt.Println(string(b))
		os.Exit(1)
	}
}

func Success() {
	res := Result{Passed: true}
	b, _ := json.Marshal(res)
	fmt.Println(string(b))
	os.Exit(0)
}

func GenerateRandomString(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteByte(byte(rand.Intn(95) + 32))
	}
	return sb.String()
}
`
	return os.WriteFile(comparefile, []byte(compareCode), 0644)
}

func compileGoCodeNsjail(dir, filename string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nsjailPath := "nsjail"
	if _, err := exec.LookPath("nsjail"); err != nil {
		nsjailPath = "/usr/local/bin/nsjail"
	}

	cmd := exec.CommandContext(ctx, nsjailPath,
		"--config", "sandbox/nsjail.cfg",
		"-B", fmt.Sprintf("%s:/app", dir),
		"--cwd", "/app",
		"--", "go", "build", "-o", "bin", filename)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(stderr.String())
	}
	return ""
}

func compileGoCodeDocker(dir, filename string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	memLimit := fmt.Sprintf("%dm", config.AppConfig.SandboxMemoryMB)
	if config.AppConfig.SandboxMemoryMB <= 0 {
		memLimit = "256m"
	}

	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"--memory="+memLimit,
		"--cpus=0.5",
		"--pids-limit=100",
		"--network", "none",
		"--security-opt=no-new-privileges:true",
		"--cap-drop=ALL",
		"--tmpfs", "/tmp:rw,exec,nosuid,size=64m",
		"-e", "GO111MODULE=on",
		"-v", fmt.Sprintf("%s:/app", dir),
		"-w", "/app",
		"golang:1.20-alpine",
		"go", "build", "-o", "bin", filename)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(stderr.String())
	}
	return ""
}

func parseRunnerOutput(rawOut, rawErr string, cmdErr error) RunResult {
	rawOut = strings.TrimSpace(rawOut)
	rawErr = strings.TrimSpace(rawErr)

	var result RunResult
	parseErr := json.Unmarshal([]byte(rawOut), &result)
	if parseErr != nil {
		msg := rawErr
		if rawErr == "" {
			msg = rawOut
		}
		if cmdErr != nil && msg == "" {
			msg = cmdErr.Error()
		}
		return RunResult{
			Passed:       false,
			ErrorMessage: "Test Runner Error (or Invalid Output):\n" + msg,
		}
	}

	return result
}
