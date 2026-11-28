# nixpkgs-health-check-action

[![CI Status](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/ci.yml?query=branch%3Amain+)

Checks whether specific Nixpkgs packages build on Hydra and pass `nixpkgs-update`, and reports results to `GITHUB_STEP_SUMMARY` and an optional GitHub Dashboard Issue.

## Usage

### Live example (Dogfooding)

- Maintained packages: [Workflow](.github/workflows/dogfood-maintainer.yml) [![🐶 Status](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/dogfood-maintainer.yml/badge.svg?branch=main)](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/dogfood-maintainer.yml?query=branch%3Amain+)
- Interested packages: [Workflow](.github/workflows/dogfood-interests.yml) [![🐶 (Interests) Status](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/dogfood-interests.yml/badge.svg?branch=main)](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/dogfood-interests.yml?query=branch%3Amain+)

### Check single or multiple packages

```yaml
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: NixOS/nix-installer-action@v1
      - uses: kachick/nixpkgs-health-check-action@main
        with:
          pname: 'hello'
```

### Check packages with a Dashboard Issue

Set `dashboard-id` (and grant `issues: write`) to keep one open issue updated with the latest status. When `assignee` is set, the workflow assigns that user while errors exist and unassigns them once all errors are resolved. Package failures do not fail the workflow run; only tool or configuration errors fail CI.

```yaml
jobs:
  check:
    permissions:
      contents: read
      issues: write
    uses: kachick/nixpkgs-health-check-action/.github/workflows/health-check.yml@main
    with:
      pnames: '["hello", "biz-ud-gothic"]'
      dashboard-id: 'packages'
      assignee: 'your-id'
```

```yaml
jobs:
  check:
    permissions:
      contents: read
      issues: write
    uses: kachick/nixpkgs-health-check-action/.github/workflows/health-check-by-maintainer.yml@main
    with:
      maintainer: 'kachick'
      dashboard-id: 'maintainer'
```

## Configuration: Skipping and Ignoring Checks

Skip or downgrade known issues by placing a TOML file (default: `nixpkgs-health-check-by-maintainer.toml`) in the repository.

```toml
[packages.typescript-go.nixpkgs-update]
skip = "Failing in nixpkgs-update, see issue #..."

[packages.hello]
hydra = { skip = "Just for testing" }
nixpkgs-update = { skip = "Just for testing" }

# Downgrade specific failing Hydra build URLs to warnings
[[packages.rumdl.hydra.ignore]]
url = "https://hydra.nixos.org/build/123456"
reason = "Fix merged, waiting for next evaluation"

# Skip nixpkgs-update logs matching an exact URL
[[packages.fresh-editor.nixpkgs-update.ignore]]
url = "https://nixpkgs-update-logs.nixos.org/fresh-editor/2026-11-10.log"
reason = "Fix merged, waiting for next run"

# Global nixpkgs-update ignore rules across packages
[[global.nixpkgs-update.ignore]]
pattern = 'go: go\.mod requires go >= [0-9.]+'
reason = "Allow Go version delay across packages"
```

## Dependencies

- [hydra-check](https://github.com/nix-community/hydra-check)
- [nixpkgs-update-log-checker](https://github.com/kachick/nixpkgs-update-log-checker)
- [nixpkgs-maintained-by](https://github.com/kachick/nixpkgs-maintained-by)

## Scope

- Checks only specified packages, not all of Nixpkgs.
- Checks [nixpkgs-update](https://github.com/nix-community/nixpkgs-update) [logs](https://nixpkgs-update-logs.nixos.org/), not upstream releases directly.
