package report

import (
	"strings"
	"testing"
	"time"

	"github.com/kachick/nixpkgs-health-check-action/internal/hydra"
	"github.com/kachick/nixpkgs-health-check-action/internal/runner"
	"github.com/kachick/nixpkgs-health-check-action/internal/update"
)

func TestRenderMarkdownErrorsAndHealthy(t *testing.T) {
	url := "https://hydra.nixos.org/build/123"
	rep := &runner.Report{
		Packages: []runner.PackageReport{
			{
				Pname:        "broken-pkg",
				HydraEnabled: true,
				Hydra: &hydra.CheckResult{
					Errors: []hydra.BuildResult{
						{Attr: "broken-pkg", Arch: "x86_64-linux", Status: "Failed", URL: &url, Severity: hydra.SeverityError, Icon: "✖"},
					},
				},
				UpdateEnabled: true,
				Update: &update.CheckResult{
					Status: update.StatusError,
					Output: "```boom```",
				},
			},
			{
				Pname:           "skipped-pkg",
				HydraEnabled:    true,
				HydraSkipped:    true,
				HydraSkipReason: "Known issue",
				UpdateEnabled:   true,
				Update:          &update.CheckResult{Status: update.StatusOK},
			},
		},
	}

	md := RenderMarkdown(rep, RenderOptions{
		DashboardID: "maintainer",
		Assignee:    "kachick",
		RunURL:      "https://github.com/kachick/repo/actions/runs/42",
		GeneratedAt: time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC),
	})

	for _, want := range []string{
		"<!-- nixpkgs-health-check-dashboard: maintainer -->",
		"**Status:** ❌ 2 error(s)",
		"**Assignee:** @kachick",
		"## Errors",
		"https://hydra.nixos.org/build/123",
		"````text",
		"⏭ Skipped (Known issue)",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q:\n%s", want, md)
		}
	}

	healthy := RenderMarkdown(&runner.Report{
		Packages: []runner.PackageReport{
			{
				Pname:         "ok-pkg",
				HydraEnabled:  true,
				Hydra:         &hydra.CheckResult{},
				UpdateEnabled: true,
				Update:        &update.CheckResult{Status: update.StatusOK},
			},
		},
	}, RenderOptions{
		DashboardID: "maintainer",
		Assignee:    "kachick",
		GeneratedAt: time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC),
	})

	if !strings.Contains(healthy, "**Status:** ✅ All healthy") {
		t.Errorf("expected healthy status:\n%s", healthy)
	}
	if strings.Contains(healthy, "@kachick") {
		t.Errorf("healthy report should not mention assignee:\n%s", healthy)
	}
}
