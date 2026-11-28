package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/kachick/nixpkgs-health-check-action/internal/runner"
	"github.com/kachick/nixpkgs-health-check-action/internal/update"
)

type RenderOptions struct {
	DashboardID string
	Assignee    string
	RunURL      string
	GeneratedAt time.Time
}

func Marker(dashboardID string) string {
	return fmt.Sprintf("<!-- nixpkgs-health-check-dashboard: %s -->", dashboardID)
}

func RenderMarkdown(rep *runner.Report, opts RenderOptions) string {
	var b strings.Builder

	if opts.DashboardID != "" {
		fmt.Fprintf(&b, "%s\n\n", Marker(opts.DashboardID))
	}

	errCount := rep.ErrorCount()
	warnCount := rep.WarningCount()

	switch {
	case errCount > 0 && warnCount > 0:
		fmt.Fprintf(&b, "**Status:** ❌ %d error(s), ⚠️ %d warning(s)\n", errCount, warnCount)
	case errCount > 0:
		fmt.Fprintf(&b, "**Status:** ❌ %d error(s)\n", errCount)
	case warnCount > 0:
		fmt.Fprintf(&b, "**Status:** ⚠️ %d warning(s)\n", warnCount)
	default:
		b.WriteString("**Status:** ✅ All healthy\n")
	}

	if errCount > 0 && opts.Assignee != "" {
		fmt.Fprintf(&b, "**Assignee:** @%s\n", opts.Assignee)
	}

	ts := opts.GeneratedAt.UTC()
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	fmt.Fprintf(&b, "**Last updated:** %s\n", ts.Format(time.RFC3339))
	if opts.RunURL != "" {
		fmt.Fprintf(&b, "**Workflow run:** %s\n", opts.RunURL)
	}
	b.WriteString("\n")

	if errCount > 0 {
		writeErrorsSection(&b, rep)
	}
	if warnCount > 0 {
		writeWarningsSection(&b, rep)
	}
	writeSummaryTable(&b, rep)

	return b.String()
}

func writeErrorsSection(b *strings.Builder, rep *runner.Report) {
	b.WriteString("## Errors\n\n")

	hasHydraErrors := false
	for _, p := range rep.Packages {
		if p.Hydra != nil && len(p.Hydra.Errors) > 0 {
			hasHydraErrors = true
			break
		}
	}
	if hasHydraErrors {
		b.WriteString("### Hydra\n\n")
		b.WriteString("| Package | Attr | Arch | Status | Build |\n")
		b.WriteString("| --- | --- | --- | --- | --- |\n")
		for _, p := range rep.Packages {
			if p.Hydra == nil {
				continue
			}
			for _, e := range p.Hydra.Errors {
				fmt.Fprintf(b, "| `%s` | `%s` | `%s` | %s | %s |\n",
					escapeCell(p.Pname),
					escapeCell(e.Attr),
					escapeCell(e.Arch),
					escapeCell(e.Status),
					formatURL(e.URL),
				)
			}
		}
		b.WriteString("\n")
	}

	hasUpdateErrors := false
	for _, p := range rep.Packages {
		if p.Update != nil && p.Update.Status == update.StatusError {
			hasUpdateErrors = true
			break
		}
	}
	if hasUpdateErrors {
		b.WriteString("### nixpkgs-update\n\n")
		for _, p := range rep.Packages {
			if p.Update == nil || p.Update.Status != update.StatusError {
				continue
			}
			fmt.Fprintf(b, "<details>\n<summary><code>%s</code></summary>\n\n", escapeHTML(p.Pname))
			writeCodeBlock(b, p.Update.Output)
			b.WriteString("\n</details>\n\n")
		}
	}
}

func writeWarningsSection(b *strings.Builder, rep *runner.Report) {
	b.WriteString("## Warnings\n\n")

	hasHydraWarnings := false
	for _, p := range rep.Packages {
		if p.Hydra != nil && len(p.Hydra.Warnings) > 0 {
			hasHydraWarnings = true
			break
		}
	}
	if hasHydraWarnings {
		b.WriteString("### Hydra\n\n")
		b.WriteString("| Package | Attr | Arch | Status | Build |\n")
		b.WriteString("| --- | --- | --- | --- | --- |\n")
		for _, p := range rep.Packages {
			if p.Hydra == nil {
				continue
			}
			for _, w := range p.Hydra.Warnings {
				fmt.Fprintf(b, "| `%s` | `%s` | `%s` | %s | %s |\n",
					escapeCell(p.Pname),
					escapeCell(w.Attr),
					escapeCell(w.Arch),
					escapeCell(w.Status),
					formatURL(w.URL),
				)
			}
		}
		b.WriteString("\n")
	}

	for _, p := range rep.Packages {
		if p.Update != nil && p.Update.Status == update.StatusWarning {
			fmt.Fprintf(b, "### nixpkgs-update (`%s`)\n\n", escapeCell(p.Pname))
			writeCodeBlock(b, p.Update.Output)
			b.WriteString("\n")
		}
	}

	hasStale := false
	for _, p := range rep.Packages {
		if (p.Hydra != nil && len(p.Hydra.StaleIgnore) > 0) || (p.Update != nil && p.Update.StaleIgnore) {
			hasStale = true
			break
		}
	}
	if hasStale {
		b.WriteString("### Stale Ignore Rules\n\n")
		for _, p := range rep.Packages {
			if p.Hydra != nil && len(p.Hydra.StaleIgnore) > 0 {
				for _, u := range p.Hydra.StaleIgnore {
					fmt.Fprintf(b, "- `%s` (hydra): %s\n", escapeCell(p.Pname), u)
				}
			}
			if p.Update != nil && p.Update.StaleIgnore {
				fmt.Fprintf(b, "- `%s` (nixpkgs-update): `packages.%s.nixpkgs-update.ignore` matched no warnings\n", escapeCell(p.Pname), escapeCell(p.Pname))
			}
		}
		b.WriteString("\n")
	}
}

func writeSummaryTable(b *strings.Builder, rep *runner.Report) {
	b.WriteString("## Packages\n\n")
	b.WriteString("| Package | Hydra | nixpkgs-update |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, p := range rep.Packages {
		fmt.Fprintf(b, "| `%s` | %s | %s |\n",
			escapeCell(p.Pname),
			formatHydraCell(p),
			formatUpdateCell(p),
		)
	}
}

func formatHydraCell(p runner.PackageReport) string {
	if !p.HydraEnabled {
		return "-"
	}
	if p.HydraSkipped {
		return fmt.Sprintf("⏭ Skipped (%s)", escapeCell(p.HydraSkipReason))
	}
	if p.Hydra == nil {
		return "-"
	}
	if len(p.Hydra.Errors) > 0 {
		return fmt.Sprintf("✖ %d error(s)", len(p.Hydra.Errors))
	}
	if len(p.Hydra.Warnings) > 0 || len(p.Hydra.StaleIgnore) > 0 {
		parts := []string{}
		if len(p.Hydra.Warnings) > 0 {
			parts = append(parts, fmt.Sprintf("%d warning(s)", len(p.Hydra.Warnings)))
		}
		if len(p.Hydra.StaleIgnore) > 0 {
			parts = append(parts, "stale ignore")
		}
		return "⚠ " + strings.Join(parts, ", ")
	}
	return "✔ OK"
}

func formatUpdateCell(p runner.PackageReport) string {
	if !p.UpdateEnabled {
		return "-"
	}
	if p.UpdateSkipped {
		return fmt.Sprintf("⏭ Skipped (%s)", escapeCell(p.UpdateSkipReason))
	}
	if p.Update == nil {
		return "-"
	}
	switch p.Update.Status {
	case update.StatusError:
		return "✖ Error"
	case update.StatusWarning:
		return "⚠ Warning"
	default:
		if p.Update.StaleIgnore {
			return "⚠ Stale ignore"
		}
		return "✔ OK"
	}
}

func formatURL(u *string) string {
	if u == nil || *u == "" {
		return "-"
	}
	return *u
}

func escapeCell(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "|", `\|`)
	return s
}

func escapeHTML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

func writeCodeBlock(b *strings.Builder, content string) {
	// Use a fence longer than any backtick run in content to prevent breaking out of the block.
	fence := "```"
	for strings.Contains(content, fence) {
		fence += "`"
	}
	fmt.Fprintf(b, "%stext\n%s\n%s\n", fence, content, fence)
}
