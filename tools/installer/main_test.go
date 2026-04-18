package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallFailsPreflightBeforeTargetMutation(t *testing.T) {
	codexHome := filepath.Join(t.TempDir(), ".codex")
	t.Setenv("CODEX_HOME", codexHome)

	cfg := config{
		RepoRoot: filepath.Join(t.TempDir(), "wrong-root"),
		Target:   "codex",
		Mode:     "copy",
	}

	err := install(cfg)
	if err == nil {
		t.Fatal("install() error = nil, want preflight failure")
	}

	errMsg := err.Error()
	for _, want := range []string{
		`preflight validate repo root`,
		"skills",
		"agents",
		"mothership-config",
		filepath.Join(".mothership", "hub", "README.md"),
	} {
		if !strings.Contains(errMsg, want) {
			t.Fatalf("install() error = %q, want substring %q", errMsg, want)
		}
	}

	if _, statErr := os.Stat(codexHome); !os.IsNotExist(statErr) {
		t.Fatalf("target root %s exists or returned unexpected error: %v", codexHome, statErr)
	}
}

func TestInstallFailsPreflightWhenRequiredPathHasWrongType(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "skills"), "not a directory")
	mkdirAll(t, filepath.Join(repoRoot, "agents"))
	mkdirAll(t, filepath.Join(repoRoot, "mothership-config"))
	mkdirAll(t, filepath.Join(repoRoot, ".mothership", "hub"))
	writeFile(t, filepath.Join(repoRoot, ".mothership", "hub", "README.md"), "hub contract")

	codexHome := filepath.Join(t.TempDir(), ".codex")
	t.Setenv("CODEX_HOME", codexHome)

	cfg := config{
		RepoRoot: repoRoot,
		Target:   "codex",
		Mode:     "copy",
	}

	err := install(cfg)
	if err == nil {
		t.Fatal("install() error = nil, want preflight failure")
	}

	errMsg := err.Error()
	for _, want := range []string{
		`preflight validate repo root`,
		"wrong path types",
		"skills",
		"must be a directory",
	} {
		if !strings.Contains(errMsg, want) {
			t.Fatalf("install() error = %q, want substring %q", errMsg, want)
		}
	}

	if _, statErr := os.Stat(codexHome); !os.IsNotExist(statErr) {
		t.Fatalf("target root %s exists or returned unexpected error: %v", codexHome, statErr)
	}

	for _, relPath := range []string{
		"skills",
		"agents",
		"mothership-config",
		filepath.Join(".mothership", "hub", "README.md"),
	} {
		dstPath := filepath.Join(codexHome, relPath)
		if _, statErr := os.Stat(dstPath); !os.IsNotExist(statErr) {
			t.Fatalf("destination %s exists or returned unexpected error: %v", dstPath, statErr)
		}
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
