package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	fs.StringVar(&cfg.Target, "target", "", "install target: codex, claude, antigravity, opencode, pi")
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
	return fmt.Errorf("%w\n\nUsage: ./install.sh [--target codex|claude|antigravity|opencode] [--mode symlink|copy] [--force]", err)
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
					huh.NewOption("OpenCode", "opencode"),
					huh.NewOption("Pi", "pi"),
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
	if err := validateRepoRoot(cfg.RepoRoot); err != nil {
		return err
	}

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

	if cfg.Target == "opencode" {
		cmdCount, err := installOpenCodeCommands(cfg.RepoRoot, targetRoot, cfg.Force)
		if err != nil {
			return err
		}

		if err := verifyOpenCodeStructure(targetRoot); err != nil {
			return fmt.Errorf("post-install verification: %w", err)
		}

		fmt.Printf("  opencode commands: %d installed in %s\n", cmdCount, filepath.Join(targetRoot, "commands"))
	}

	if cfg.Target == "pi" {
		extSrc := filepath.Join(cfg.RepoRoot, "skills", "mothership", "extensions", "subagent.ts")
		extDst := filepath.Join(targetRoot, "extensions", "subagent.ts")
		if err := copyFile(extSrc, extDst, cfg.Force); err != nil {
			return err
		}
		fmt.Printf("  mothership_spawn extension: %s\n", extDst)
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

func validateRepoRoot(repoRoot string) error {
	requiredPaths := []struct {
		relPath string
		wantDir bool
	}{
		{relPath: "skills", wantDir: true},
		{relPath: "agents", wantDir: true},
		{relPath: "mothership-config", wantDir: true},
		{relPath: filepath.Join(".mothership", "hub", "README.md"), wantDir: false},
	}

	var missing []string
	var invalidType []string
	for _, requiredPath := range requiredPaths {
		fullPath := filepath.Join(repoRoot, requiredPath.relPath)
		info, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				missing = append(missing, fmt.Sprintf("%s (%s)", requiredPath.relPath, fullPath))
				continue
			}
			return fmt.Errorf("preflight validate repo root %q: inspect %s: %w", repoRoot, fullPath, err)
		}

		if requiredPath.wantDir && !info.IsDir() {
			invalidType = append(invalidType, fmt.Sprintf("%s (%s): must be a directory", requiredPath.relPath, fullPath))
			continue
		}

		if !requiredPath.wantDir && info.IsDir() {
			invalidType = append(invalidType, fmt.Sprintf("%s (%s): must be a file", requiredPath.relPath, fullPath))
		}
	}

	var problems []string
	if len(missing) > 0 {
		problems = append(problems, "missing required paths: "+strings.Join(missing, ", "))
	}
	if len(invalidType) > 0 {
		problems = append(problems, "wrong path types: "+strings.Join(invalidType, ", "))
	}

	if len(problems) > 0 {
		return fmt.Errorf("preflight validate repo root %q failed; %s", repoRoot, strings.Join(problems, "; "))
	}

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
	case "opencode":
		return resolveOpenCodeDir()
	case "pi":
		return filepath.Join(os.Getenv("HOME"), ".pi", "agent")
	default:
		return ""
	}
}

func resolveOpenCodeDir() string {
	if root := os.Getenv("OPENCODE_HOME"); root != "" {
		return root
	}

	home := os.Getenv("HOME")

	searchPaths := []string{}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		searchPaths = append(searchPaths, filepath.Join(xdg, "opencode"))
	}
	searchPaths = append(searchPaths,
		filepath.Join(home, ".config", "opencode"),
		filepath.Join(home, ".opencode"),
	)

	for _, candidate := range searchPaths {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "opencode")
	}
	return filepath.Join(home, ".config", "opencode")
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

var supportedTargets = map[string]bool{
	"codex": true, "claude": true, "antigravity": true, "opencode": true, "pi": true,
}

var supportedModes = map[string]bool{
	"symlink": true, "copy": true,
}

func isSupportedTarget(target string) bool {
	return supportedTargets[target]
}

func isSupportedMode(mode string) bool {
	return supportedModes[mode]
}

func installOpenCodeCommands(repoRoot, targetRoot string, force bool) (int, error) {
	commandsDir := filepath.Join(targetRoot, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		return 0, fmt.Errorf("create commands dir: %w", err)
	}

	skillsDir := filepath.Join(repoRoot, "skills")
	var count int

	err := filepath.Walk(skillsDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		rel, relErr := filepath.Rel(skillsDir, path)
		if relErr != nil {
			return fmt.Errorf("compute relative path for %s: %w", path, relErr)
		}

		if strings.HasPrefix(rel, ".system") {
			return nil
		}

		var cmdRel string
		dir := filepath.Dir(rel)
		if info.Name() == "SKILL.md" {
			if dir == "." {
				return nil
			}
			cmdRel = strings.ReplaceAll(dir, string(filepath.Separator), "/") + ".md"
		} else if strings.HasSuffix(info.Name(), ".md") && dir == "." {
			cmdRel = info.Name()
		} else {
			return nil
		}

		cmdFile := filepath.Join(commandsDir, cmdRel)
		if err := copyFile(path, cmdFile, force); err != nil {
			return fmt.Errorf("install command %s: %w", cmdRel, err)
		}

		count++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("walk skills for opencode commands: %w", err)
	}

	return count, nil
}

func verifyOpenCodeStructure(targetRoot string) error {
	commandsDir := filepath.Join(targetRoot, "commands")
	info, err := os.Stat(commandsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("commands directory missing: %s", commandsDir)
		}
		return fmt.Errorf("inspect commands dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("commands path is not a directory: %s", commandsDir)
	}

	var cmdFiles []string
	var emptyFiles []string
	if err := filepath.Walk(commandsDir, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil || fi.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		rel, relErr := filepath.Rel(commandsDir, path)
		if relErr != nil {
			rel = path
		}
		cmdFiles = append(cmdFiles, rel)

		if fi.Size() == 0 {
			emptyFiles = append(emptyFiles, rel)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walk commands dir: %w", err)
	}

	var problems []string
	if len(cmdFiles) == 0 {
		problems = append(problems, "no command files found in commands/")
	}
	if len(emptyFiles) > 0 {
		problems = append(problems, fmt.Sprintf("empty command files: %s", strings.Join(emptyFiles, ", ")))
	}

	if len(problems) > 0 {
		return fmt.Errorf("structure issues: %s", strings.Join(problems, "; "))
	}

	return nil
}
