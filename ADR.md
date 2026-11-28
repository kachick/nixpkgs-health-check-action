# Architecture Decision Record (ADR)

This document records the architectural decisions made during the evolution of this repository.

## 1. Dashboard Issue Reporting Instead of Failing CI (2026-11)

### Context

Scheduled health checks used to exit with code 1 whenever a package failed on Hydra or `nixpkgs-update`. When an upstream issue could not be fixed right away, daily workflow failure alerts became noisy, even though the check itself ran as expected.

### Decision

Separate tool/configuration errors from package health results:

- Exit with code 1 only when the check itself fails (bad config, missing binary, GitHub API failure).
- Write package errors, warnings, and stale ignore rules into a single GitHub Dashboard Issue (similar to Renovate's Dependency Dashboard).
- Assign the configured user when at least one error is present, and unassign when all errors are gone.

### Security Guardrails

- Match existing dashboard issues by an HTML comment marker (`<!-- nixpkgs-health-check-dashboard: <id> -->`), and only update open issues authored by `github-actions[bot]` or a repository collaborator (`OWNER`, `MEMBER`, `COLLABORATOR`).
- Keep `issues: write` out of pull request jobs; PR runs only print the report to `GITHUB_STEP_SUMMARY`.

## 2. Single-Job Parallel Go Runner (2026-11)

### Context

Earlier versions used a `plan` job plus a per-package GitHub Actions matrix (`pname`) with separate bash/jq/Nix scripts (`check-skip`, `extract-nixpkgs-update-config`, `nixpkgs-health-check-hydra`). That structure was chosen so each package had its own job row in the Actions UI when CI failures were the alert mechanism.

### Decision

Replace the matrix and shell scripts with a single Go CLI (`nixpkgs-health-check`) running all packages concurrently in one job.

### Why?

- **Efficiency**: Installs Nix and builds `nixpkgs-update-log-checker` once per run instead of repeating setup for every package.
- **Simpler Aggregation**: Building one Dashboard Issue from all packages needs no artifact upload/download or separate report job.
- **Maintainability**: Typed TOML/JSON parsing (`BurntSushi/toml` + Go 1.27 `encoding/json/v2`), unit tests, and GitHub API calls live in one static binary instead of shell pipelines.

## 3. Two-Layer Workflow Structure

### Context

Should a single reusable workflow handle both explicit package lists and maintainer lookups?

### Decision

Keep `health-check.yml` (Generic List Runner) and `health-check-by-maintainer.yml` (Maintainer Adapter) separate.

### Why?

- **Validation**: Each workflow keeps clear `required: true` inputs.
- **Separation of Concerns**: The runner checks a list of packages, while the adapter resolves the list for a maintainer.
