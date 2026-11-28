package hydra

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

type Severity string

const (
	SeveritySuccess Severity = "success"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type BuildResult struct {
	Tool     string   `json:"tool"`
	Attr     string   `json:"attr"`
	Arch     string   `json:"arch"`
	Status   string   `json:"status"`
	URL      *string  `json:"url"`
	Severity Severity `json:"severity"`
	Icon     string   `json:"icon"`
}

type CheckResult struct {
	Builds      []BuildResult
	Errors      []BuildResult
	Warnings    []BuildResult
	StaleIgnore []string
}

type rawEval struct {
	NowFail      []rawBuild `json:"now_fail"`
	StillFail    []rawBuild `json:"still_fail"`
	Aborted      []rawBuild `json:"aborted"`
	NowSucceed   []rawBuild `json:"now_succeed"`
	StillSucceed []rawBuild `json:"still_succeed"`
	New          []rawBuild `json:"new"`
	Unfinished   []rawBuild `json:"unfinished"`
}

type rawBuild struct {
	Name     string  `json:"name"`
	JobName  string  `json:"job_name"`
	Arch     *string `json:"arch"`
	Status   string  `json:"status"`
	BuildURL *string `json:"build_url"`
	URL      *string `json:"url"`
}

var evalKeyPattern = regexp.MustCompile(`^[0-9]+/(.+)$`)

// Runner executes hydra-check and returns its raw stdout.
type Runner func(ctx context.Context, pname string) ([]byte, error)

// DefaultRunner invokes hydra-check from PATH.
// hydra-check 2.1.0 returns exit code 1 when any build failed in the evaluation,
// so non-zero exit codes with stdout output are still parsed.
// See https://github.com/nix-community/hydra-check/pull/94
func DefaultRunner(ctx context.Context, pname string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "hydra-check", "--json", "--channel=nixpkgs-unstable", "--eval", pname)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && stdout.Len() > 0 {
			return stdout.Bytes(), nil
		}
		return nil, fmt.Errorf("run hydra-check for %q: %w (stderr: %s)", pname, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

func Check(ctx context.Context, run Runner, pname string, allIgnoreURLs, pkgIgnoreURLs []string) (*CheckResult, error) {
	raw, err := run(ctx, pname)
	if err != nil {
		return nil, err
	}
	builds, err := ParseEvalStream(bytes.NewReader(raw), allIgnoreURLs)
	if err != nil {
		return nil, fmt.Errorf("parse hydra-check output for %q: %w", pname, err)
	}

	res := &CheckResult{Builds: builds}
	seenURLs := make(map[string]struct{}, len(builds))
	for _, b := range builds {
		if b.URL != nil {
			seenURLs[*b.URL] = struct{}{}
		}
		switch b.Severity {
		case SeverityError:
			res.Errors = append(res.Errors, b)
		case SeverityWarning:
			res.Warnings = append(res.Warnings, b)
		}
	}

	if len(pkgIgnoreURLs) > 0 && len(builds) > 0 {
		for _, u := range pkgIgnoreURLs {
			if _, ok := seenURLs[u]; !ok {
				res.StaleIgnore = append(res.StaleIgnore, u)
			}
		}
	}
	return res, nil
}

// ParseEvalStream decodes one or more top-level JSON objects emitted by hydra-check --json --eval.
func ParseEvalStream(r io.Reader, ignoreURLs []string) ([]BuildResult, error) {
	dec := jsontext.NewDecoder(r)
	var results []BuildResult

	for {
		var top map[string]jsontext.Value
		err := json.UnmarshalDecode(dec, &top)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode hydra-check JSON stream: %w", err)
		}

		// Sort keys for deterministic output across Go map iterations.
		keys := make([]string, 0, len(top))
		for k := range top {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		for _, k := range keys {
			m := evalKeyPattern.FindStringSubmatch(k)
			if m == nil {
				continue
			}
			queryPname := m[1]

			var ev rawEval
			if err := json.Unmarshal(top[k], &ev); err != nil {
				return nil, fmt.Errorf("decode eval entry %q: %w", k, err)
			}

			groups := [][]rawBuild{
				ev.NowFail,
				ev.StillFail,
				ev.Aborted,
				ev.NowSucceed,
				ev.StillSucceed,
				ev.New,
				ev.Unfinished,
			}
			for _, group := range groups {
				for _, b := range group {
					if b.Arch == nil || *b.Arch == "" {
						continue
					}
					attr := stripArchSuffix(b.JobName)
					// Match canonical package on all platforms while skipping nested/test variants.
					if attr != queryPname {
						continue
					}
					urlPtr := b.BuildURL
					if urlPtr == nil {
						urlPtr = b.URL
					}
					sev := classifySeverity(*b.Arch, b.Status, urlPtr, ignoreURLs)
					results = append(results, BuildResult{
						Tool:     b.Name,
						Attr:     attr,
						Arch:     *b.Arch,
						Status:   b.Status,
						URL:      urlPtr,
						Severity: sev,
						Icon:     severityIcon(sev),
					})
				}
			}
		}
	}

	if results == nil {
		results = []BuildResult{}
	}
	return results, nil
}

func stripArchSuffix(jobName string) string {
	idx := strings.LastIndex(jobName, ".")
	if idx < 0 {
		return jobName
	}
	return jobName[:idx]
}

func classifySeverity(arch, status string, urlPtr *string, ignoreURLs []string) Severity {
	if status == "Succeeded" || status == "Queued" {
		return SeveritySuccess
	}
	if arch == "x86_64-darwin" ||
		status == "Dependency failed" ||
		status == "Cancelled" ||
		status == "Aborted" ||
		(urlPtr != nil && slices.Contains(ignoreURLs, *urlPtr)) {
		return SeverityWarning
	}
	return SeverityError
}

func severityIcon(sev Severity) string {
	switch sev {
	case SeveritySuccess:
		return "✔"
	case SeverityWarning:
		return "⚠"
	default:
		return "✖"
	}
}
