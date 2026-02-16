package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testDockerConfig struct {
	Auths map[string]registryAuth `json:"auths"`
}

func runForTest(t *testing.T, args ...string) (int, string, string) {
	t.Helper()

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	exitCode := run("dologen", args, stdout, stderr)

	return exitCode, stdout.String(), stderr.String()
}

func TestRunVersion(t *testing.T) {
	exitCode, stdout, stderr := runForTest(t, "--version")

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if stdout != version+"\n" {
		t.Fatalf("unexpected version output: %q", stdout)
	}
}

func TestRunRejectsMissingUsername(t *testing.T) {
	exitCode, _, stderr := runForTest(t, "--password", "secret", "--server", "registry.local")
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "error: username cannot be empty") {
		t.Fatalf("expected username error, got %q", stderr)
	}
}

func TestRunCreatesJSON(t *testing.T) {
	exitCode, stdout, stderr := runForTest(t,
		"--username", "alice",
		"--password", "s3cr3t",
		"--server", "registry.example.com",
	)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
	}

	var config testDockerConfig
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &config); err != nil {
		t.Fatalf("output should be valid json: %v", err)
	}

	auth, ok := config.Auths["registry.example.com"]
	if !ok {
		t.Fatalf("expected registry key in output, got %v", config.Auths)
	}
	if auth.Username != "alice" || auth.Password != "s3cr3t" {
		t.Fatalf("unexpected credentials in output: %+v", auth)
	}
}

func TestRunCreatesBase64EncodedJSON(t *testing.T) {
	exitCode, stdout, stderr := runForTest(t,
		"--username", "alice",
		"--password", "s3cr3t",
		"--server", "registry.example.com",
		"--base64",
	)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(stdout))
	if err != nil {
		t.Fatalf("expected base64 output: %v", err)
	}

	var config testDockerConfig
	if err := json.Unmarshal(decoded, &config); err != nil {
		t.Fatalf("decoded output should be valid json: %v", err)
	}
}

func TestRunUsesPasswordFile(t *testing.T) {
	tempDir := t.TempDir()
	passwordPath := filepath.Join(tempDir, "password.txt")
	if err := os.WriteFile(passwordPath, []byte("from-file\n"), 0o600); err != nil {
		t.Fatalf("write password file: %v", err)
	}

	exitCode, stdout, stderr := runForTest(t,
		"--username", "alice",
		"--password-file", passwordPath,
		"--server", "registry.example.com",
	)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
	}
	if strings.Contains(stdout, "from-file\\n") {
		t.Fatalf("password should have trailing newline stripped")
	}
}

func TestRunWarnsOnLoosePasswordFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	passwordPath := filepath.Join(tempDir, "password.txt")
	if err := os.WriteFile(passwordPath, []byte("from-file"), 0o644); err != nil {
		t.Fatalf("write password file: %v", err)
	}

	exitCode, _, stderr := runForTest(t,
		"--username", "alice",
		"--password-file", passwordPath,
		"--server", "registry.example.com",
	)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stderr, "warning: password file") {
		t.Fatalf("expected permissions warning, got %q", stderr)
	}
}

func TestRunCompletionBash(t *testing.T) {
	exitCode, stdout, stderr := runForTest(t, "--completion", "bash")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "complete -F") {
		t.Fatalf("expected bash completion output, got %q", stdout)
	}
}

func TestRunRejectsUnknownCompletionShell(t *testing.T) {
	exitCode, _, stderr := runForTest(t, "--completion", "fish")
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "unsupported shell") {
		t.Fatalf("expected unsupported shell error, got %q", stderr)
	}
}
