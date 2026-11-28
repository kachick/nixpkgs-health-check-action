package update

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Status string

const (
	StatusOK      Status = "ok"
	StatusWarning Status = "warning"
	StatusError   Status = "error"
)

type CheckResult struct {
	Status      Status
	Output      string
	StaleIgnore bool
}

// Runner executes nixpkgs-update-log-checker and returns its combined output and exit code.
// A non-nil error indicates a tool execution failure (e.g., binary not found), not a health error.
type Runner func(ctx context.Context, pname string, configPath string) (output string, exitCode int, err error)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func NewCommandRunner(binPath string) Runner {
	return func(ctx context.Context, pname string, configPath string) (string, int, error) {
		var cmd *exec.Cmd
		args := []string{"--packages", pname}
		if configPath != "" {
			args = append(args, "--config", configPath)
		}

		if binPath != "" {
			cmd = exec.CommandContext(ctx, binPath, args...)
		} else if resolved, err := exec.LookPath("nixpkgs-update-log-checker"); err == nil {
			cmd = exec.CommandContext(ctx, resolved, args...)
		} else {
			nixArgs := append([]string{"run", "github:kachick/nixpkgs-update-log-checker", "--"}, args...)
			cmd = exec.CommandContext(ctx, "nix", nixArgs...)
		}

		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return out.String(), exitErr.ExitCode(), nil
			}
			return out.String(), -1, fmt.Errorf("run nixpkgs-update-log-checker for %q: %w", pname, err)
		}
		return out.String(), 0, nil
	}
}

func Check(ctx context.Context, run Runner, pname string, configTOML []byte, hasPackageIgnore bool) (*CheckResult, error) {
	var configPath string
	if len(configTOML) > 0 {
		tmpFile, err := os.CreateTemp("", "nixpkgs-update-config-*.toml")
		if err != nil {
			return nil, fmt.Errorf("create temp config for %q: %w", pname, err)
		}
		configPath = tmpFile.Name()
		defer os.Remove(configPath)

		if _, err := tmpFile.Write(configTOML); err != nil {
			tmpFile.Close()
			return nil, fmt.Errorf("write temp config for %q: %w", pname, err)
		}
		if err := tmpFile.Close(); err != nil {
			return nil, fmt.Errorf("close temp config for %q: %w", pname, err)
		}
	}

	rawOut, exitCode, err := run(ctx, pname, configPath)
	if err != nil {
		return nil, err
	}

	clean := strings.TrimSpace(ansiPattern.ReplaceAllString(rawOut, ""))
	res := &CheckResult{Output: clean}

	if exitCode != 0 {
		res.Status = StatusError
		return res, nil
	}

	if strings.Contains(clean, "[WARN]") {
		res.Status = StatusWarning
		return res, nil
	}

	res.Status = StatusOK
	if hasPackageIgnore {
		res.StaleIgnore = true
	}
	return res, nil
}
