package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
)

type config struct {
	RepoRoot string
	Target   string
	Mode     string
	Force    bool
}

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if cfg.RepoRoot == "" {
		cfg.RepoRoot, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve repo root: %v\n", err)
			os.Exit(1)
		}
	}

	if err := promptMissing(&cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := install(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseFlags(args []string) (config, error) {
	var cfg config
	fs := flag.NewFlagSet("mothership-installer", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVar(&cfg.RepoRoot, "repo-root", "", "repository root")
	fs.StringVar(&cfg.Target, "target", "", "install target: codex, claude, antigravity")
	fs.StringVar(&cfg.Mode, "mode", "", "install mode: symlink, copy")
	fs.BoolVar(&cfg.Force, "force", false, "replace existing paths")

	if err := fs.Parse(args); err != nil {
		return cfg, usageErr(err)
	}

	if cfg.Target != "" && !isSupportedTarget(cfg.Target) {
		return cfg, fmt.Errorf("unsupported target %q", cfg.Target)
	}

	if cfg.Mode != "" && !isSupportedMode(cfg.Mode) {
		return cfg, fmt.Errorf("unsupported mode %q", cfg.Mode)
	}

	return cfg, nil
}

func usageErr(err error) error {
	return fmt.Errorf("%w\n\nUsage: ./install.sh [--target codex|claude|antigravity] [--mode symlink|copy] [--force]", err)
}

func promptMissing(cfg *config) error {
	if cfg.Target != "" && cfg.Mode != "" {
		return nil
	}

	if !isInteractive() {
		return errors.New("missing --target or --mode and no interactive terminal is available")
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Install target").
				Description("Choose where to install the Mothership package").
				Options(
					huh.NewOption("Codex", "codex"),
					huh.NewOption("Claude", "claude"),
					huh.NewOption("Antigravity", "antigravity"),
				).
				Value(&cfg.Target),
			huh.NewSelect[string]().
				Title("Install mode").
				Description("Choose whether to symlink repo content or copy it").
				Options(
					huh.NewOption("Symlink", "symlink"),
					huh.NewOption("Copy", "copy"),
				).
				Value(&cfg.Mode),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("interactive prompt failed: %w", err)
	}

	return nil
}

func isInteractive() bool {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	stdoutInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (stdinInfo.Mode()&os.ModeCharDevice) != 0 && (stdoutInfo.Mode()&os.ModeCharDevice) != 0
}

func install(cfg config) error {
	targetRoot := targetRootFor(cfg.Target)
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		return fmt.Errorf("create target root: %w", err)
	}

	installs := []struct {
		src string
		dst string
	}{
		{src: filepath.Join(cfg.RepoRoot, "skills"), dst: filepath.Join(targetRoot, "skills")},
		{src: filepath.Join(cfg.RepoRoot, "agents"), dst: filepath.Join(targetRoot, "agents")},
		{src: filepath.Join(cfg.RepoRoot, "mothership-config"), dst: filepath.Join(targetRoot, "mothership-config")},
	}

	for _, item := range installs {
		if err := installPath(item.src, item.dst, cfg.Mode, cfg.Force); err != nil {
			return err
		}
	}

	hubDir := filepath.Join(targetRoot, ".mothership", "hub")
	if err := os.MkdirAll(hubDir, 0o755); err != nil {
		return fmt.Errorf("create hub dir: %w", err)
	}

	hubReadme := filepath.Join(hubDir, "README.md")
	if err := copyFile(filepath.Join(cfg.RepoRoot, ".mothership", "hub", "README.md"), hubReadme, cfg.Force); err != nil {
		return err
	}

	fmt.Printf("Installed Mothership to %s\n", targetRoot)
	fmt.Printf("  target: %s\n", cfg.Target)
	fmt.Printf("  mode: %s\n", cfg.Mode)
	fmt.Printf("  skills: %s\n", filepath.Join(targetRoot, "skills"))
	fmt.Printf("  agents: %s\n", filepath.Join(targetRoot, "agents"))
	fmt.Printf("  mothership-config: %s\n", filepath.Join(targetRoot, "mothership-config"))
	fmt.Printf("  hub contract: %s\n", hubReadme)

	return nil
}

func targetRootFor(target string) string {
	switch target {
	case "codex":
		if root := os.Getenv("CODEX_HOME"); root != "" {
			return root
		}
		return filepath.Join(os.Getenv("HOME"), ".codex")
	case "claude":
		return filepath.Join(os.Getenv("HOME"), ".claude")
	case "antigravity":
		return filepath.Join(os.Getenv("HOME"), ".antigravity")
	default:
		return ""
	}
}

func installPath(src, dst, mode string, force bool) error {
	if err := removeExisting(dst, force); err != nil {
		return err
	}

	switch mode {
	case "symlink":
		if err := os.Symlink(src, dst); err != nil {
			return fmt.Errorf("symlink %s -> %s: %w", dst, src, err)
		}
	case "copy":
		if err := copyDir(src, dst); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported install mode %q", mode)
	}

	return nil
}

func removeExisting(path string, force bool) error {
	_, err := os.Lstat(path)
	if err == nil {
		if !force {
			return fmt.Errorf("destination already exists: %s\nre-run with --force to replace it", path)
		}
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove %s: %w", path, err)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("inspect %s: %w", path, err)
	}

	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("compute relative path: %w", err)
		}

		targetPath := filepath.Join(dst, rel)

		if info.IsDir() {
			if err := os.MkdirAll(targetPath, info.Mode().Perm()); err != nil {
				return fmt.Errorf("create dir %s: %w", targetPath, err)
			}
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("read symlink %s: %w", path, err)
			}
			if err := os.Symlink(linkTarget, targetPath); err != nil {
				return fmt.Errorf("copy symlink %s: %w", targetPath, err)
			}
			return nil
		}

		return copyFileWithMode(path, targetPath, info.Mode().Perm())
	})
}

func copyFile(src, dst string, force bool) error {
	if err := removeExisting(dst, force); err != nil {
		return err
	}
	return copyFileWithMode(src, dst, 0o644)
}

func copyFileWithMode(src, dst string, mode os.FileMode) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer input.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent dir for %s: %w", dst, err)
	}

	output, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}

	return nil
}

func isSupportedTarget(target string) bool {
	return target == "codex" || target == "claude" || target == "antigravity"
}

func isSupportedMode(mode string) bool {
	return mode == "symlink" || mode == "copy"
}
