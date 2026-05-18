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

func makeValidRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	mkdirAll(t, filepath.Join(repoRoot, "skills", "commit"))
	mkdirAll(t, filepath.Join(repoRoot, "skills", "obra", "making-plans"))
	mkdirAll(t, filepath.Join(repoRoot, "skills", ".system", "imagegen"))
	mkdirAll(t, filepath.Join(repoRoot, "agents"))
	mkdirAll(t, filepath.Join(repoRoot, "mothership-config"))
	mkdirAll(t, filepath.Join(repoRoot, ".mothership", "hub"))
	writeFile(t, filepath.Join(repoRoot, ".mothership", "hub", "README.md"), "hub contract")
	writeFile(t, filepath.Join(repoRoot, "skills", "commit", "SKILL.md"), "# Commit Skill\nCreate a git commit.")
	writeFile(t, filepath.Join(repoRoot, "skills", "obra", "making-plans", "SKILL.md"), "# Making Plans\nDesign an implementation plan.")
	writeFile(t, filepath.Join(repoRoot, "skills", ".system", "imagegen", "SKILL.md"), "# Image Gen\nInternal skill.")
	writeFile(t, filepath.Join(repoRoot, "skills", "risk-assessment.md"), "# Risk Assessment\nAssess project risks.")
	return repoRoot
}

func TestInstallOpenCodeTarget(t *testing.T) {
	repoRoot := makeValidRepo(t)
	targetRoot := filepath.Join(t.TempDir(), "opencode")
	t.Setenv("OPENCODE_HOME", targetRoot)

	cfg := config{
		RepoRoot: repoRoot,
		Target:   "opencode",
		Mode:     "copy",
	}

	if err := install(cfg); err != nil {
		t.Fatalf("install() error = %v", err)
	}

	// Standard dirs installed
	for _, rel := range []string{"skills", "agents", "mothership-config"} {
		p := filepath.Join(targetRoot, rel)
		if info, err := os.Stat(p); err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s, got err=%v isDir=%v", p, err, info != nil && info.IsDir())
		}
	}

	// Commands from SKILL.md
	cmdCommit := filepath.Join(targetRoot, "commands", "commit.md")
	if data, err := os.ReadFile(cmdCommit); err != nil {
		t.Fatalf("command %s missing: %v", cmdCommit, err)
	} else if !strings.Contains(string(data), "Commit Skill") {
		t.Fatalf("command %s has wrong content: %s", cmdCommit, string(data))
	}

	// Nested SKILL.md → nested command
	cmdPlan := filepath.Join(targetRoot, "commands", "obra", "making-plans.md")
	if _, err := os.Stat(cmdPlan); err != nil {
		t.Fatalf("nested command %s missing: %v", cmdPlan, err)
	}

	// Standalone .md in skills root → command
	cmdRisk := filepath.Join(targetRoot, "commands", "risk-assessment.md")
	if _, err := os.Stat(cmdRisk); err != nil {
		t.Fatalf("standalone command %s missing: %v", cmdRisk, err)
	}

	// .system skills excluded
	cmdSystem := filepath.Join(targetRoot, "commands", ".system", "imagegen.md")
	if _, err := os.Stat(cmdSystem); !os.IsNotExist(err) {
		t.Fatalf("system command %s should not exist", cmdSystem)
	}
}

func TestVerifyOpenCodeStructure(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		root := t.TempDir()
		mkdirAll(t, filepath.Join(root, "commands"))
		writeFile(t, filepath.Join(root, "commands", "test.md"), "# Test")

		if err := verifyOpenCodeStructure(root); err != nil {
			t.Fatalf("verifyOpenCodeStructure() error = %v", err)
		}
	})

	t.Run("missing commands dir", func(t *testing.T) {
		root := t.TempDir()
		err := verifyOpenCodeStructure(root)
		if err == nil {
			t.Fatal("expected error for missing commands dir")
		}
		if !strings.Contains(err.Error(), "commands directory missing") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty commands dir", func(t *testing.T) {
		root := t.TempDir()
		mkdirAll(t, filepath.Join(root, "commands"))

		err := verifyOpenCodeStructure(root)
		if err == nil {
			t.Fatal("expected error for empty commands dir")
		}
		if !strings.Contains(err.Error(), "no command files") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestResolveOpenCodeDir(t *testing.T) {
	t.Run("OPENCODE_HOME takes priority", func(t *testing.T) {
		custom := filepath.Join(t.TempDir(), "custom-opencode")
		mkdirAll(t, custom)
		t.Setenv("OPENCODE_HOME", custom)
		// Clear XDG to avoid interference
		t.Setenv("XDG_CONFIG_HOME", "")

		got := resolveOpenCodeDir()
		if got != custom {
			t.Fatalf("resolveOpenCodeDir() = %q, want %q", got, custom)
		}
	})

	t.Run("XDG_CONFIG_HOME fallback", func(t *testing.T) {
		xdg := filepath.Join(t.TempDir(), "xdg-config")
		mkdirAll(t, filepath.Join(xdg, "opencode"))
		t.Setenv("OPENCODE_HOME", "")
		t.Setenv("XDG_CONFIG_HOME", xdg)

		got := resolveOpenCodeDir()
		want := filepath.Join(xdg, "opencode")
		if got != want {
			t.Fatalf("resolveOpenCodeDir() = %q, want %q", got, want)
		}
	})

	t.Run("default when nothing exists", func(t *testing.T) {
		t.Setenv("OPENCODE_HOME", "")
		t.Setenv("XDG_CONFIG_HOME", "")

		got := resolveOpenCodeDir()
		want := filepath.Join(os.Getenv("HOME"), ".config", "opencode")
		if got != want {
			t.Fatalf("resolveOpenCodeDir() = %q, want %q", got, want)
		}
	})
}
