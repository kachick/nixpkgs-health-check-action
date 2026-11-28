package main

import (
	"context"
	"encoding/json/v2"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kachick/nixpkgs-health-check-action/internal/config"
	gh "github.com/kachick/nixpkgs-health-check-action/internal/github"
	"github.com/kachick/nixpkgs-health-check-action/internal/hydra"
	"github.com/kachick/nixpkgs-health-check-action/internal/report"
	"github.com/kachick/nixpkgs-health-check-action/internal/runner"
	"github.com/kachick/nixpkgs-health-check-action/internal/update"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("nixpkgs-health-check", flag.ContinueOnError)
	pnamesJSON := fs.String("pnames-json", "", "JSON array of package names")
	configPath := fs.String("config", "", "Path to TOML configuration file")
	runHydra := fs.Bool("hydra", true, "Run Hydra check")
	runUpdate := fs.Bool("nixpkgs-update", true, "Run nixpkgs-update check")
	concurrency := fs.Int("concurrency", 4, "Maximum number of packages to check in parallel")
	dashboardID := fs.String("dashboard-id", "", "Dashboard identifier; when non-empty, upsert a GitHub issue")
	dashboardTitle := fs.String("dashboard-title", "", "Title for the dashboard GitHub issue")
	assignee := fs.String("assignee", "", "GitHub username to assign when errors are present")
	repo := fs.String("repo", os.Getenv("GITHUB_REPOSITORY"), "GitHub repository (owner/repo)")
	summaryFile := fs.String("summary-file", os.Getenv("GITHUB_STEP_SUMMARY"), "Path to append Markdown summary")
	updateCheckerBin := fs.String("update-checker-bin", "", "Path to nixpkgs-update-log-checker binary")

	if err := fs.Parse(args); err != nil {
		return err
	}

	pnames, err := collectPnames(*pnamesJSON, fs.Args())
	if err != nil {
		return err
	}
	if len(pnames) == 0 {
		return fmt.Errorf("at least one package name is required via --pnames-json or positional arguments")
	}

	var cfg *config.Config
	if *configPath != "" {
		if _, err := os.Stat(*configPath); err == nil {
			cfg, err = config.Load(*configPath)
			if err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check config file %q: %w", *configPath, err)
		}
	}

	rep, err := runner.Run(ctx, pnames, runner.Options{
		Config:       cfg,
		RunHydra:     *runHydra,
		RunUpdate:    *runUpdate,
		Concurrency:  *concurrency,
		HydraRunner:  hydra.DefaultRunner,
		UpdateRunner: update.NewCommandRunner(*updateCheckerBin),
	})
	if err != nil {
		return err
	}

	md := report.RenderMarkdown(rep, report.RenderOptions{
		DashboardID: *dashboardID,
		Assignee:    *assignee,
		RunURL:      buildRunURL(),
		GeneratedAt: time.Now().UTC(),
	})

	fmt.Print(md)

	if *summaryFile != "" {
		f, err := os.OpenFile(*summaryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("open summary file %q: %w", *summaryFile, err)
		}
		_, writeErr := f.WriteString(md + "\n")
		closeErr := f.Close()
		if writeErr != nil {
			return fmt.Errorf("write summary file %q: %w", *summaryFile, writeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close summary file %q: %w", *summaryFile, closeErr)
		}
	}

	if *dashboardID != "" {
		client := &gh.Client{
			Token: os.Getenv("GITHUB_TOKEN"),
			Repo:  *repo,
		}
		res, err := client.UpsertDashboard(ctx, gh.UpsertOptions{
			DashboardID: *dashboardID,
			Title:       *dashboardTitle,
			Body:        md,
			Assignee:    *assignee,
			HasErrors:   rep.HasErrors(),
		})
		if err != nil {
			return err
		}
		action := "Updated"
		if res.Created {
			action = "Created"
		}
		fmt.Fprintf(os.Stderr, "%s dashboard issue #%d: %s\n", action, res.Number, res.HTMLURL)
	}

	return nil
}

func collectPnames(pnamesJSON string, positional []string) ([]string, error) {
	var pnames []string
	if strings.TrimSpace(pnamesJSON) != "" {
		if err := json.Unmarshal([]byte(pnamesJSON), &pnames); err != nil {
			return nil, fmt.Errorf("parse --pnames-json: %w", err)
		}
	}
	pnames = append(pnames, positional...)
	return pnames, nil
}

func buildRunURL() string {
	server := os.Getenv("GITHUB_SERVER_URL")
	repo := os.Getenv("GITHUB_REPOSITORY")
	runID := os.Getenv("GITHUB_RUN_ID")
	if server == "" || repo == "" || runID == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/actions/runs/%s", strings.TrimRight(server, "/"), repo, runID)
}
