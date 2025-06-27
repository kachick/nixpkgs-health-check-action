# nixpkgs-package-health-check-action

[![CI - Nix Status](https://github.com/kachick/nixpkgs-package-health-check-action/actions/workflows/ci-nix.yml/badge.svg?branch=main)](https://github.com/kachick/nixpkgs-package-health-check-action/actions/workflows/ci-nix.yml?query=branch%3Amain+)

This GitHub Action helps ensure that specific Nixpkgs packages are buildable and up-to-date.

## Usage

In your workflow, add this action after installing Nix, enable flake, and use cache helpers like this:

```yaml
name: Check packages health

env:
  # Delimited with " " is useful for workflow_dispatch rather than lines
  MY_MAINTAINED_PACKAGES: |
    typescript-go lima dprint brush

on:
  schedule:
    # e.g: “At 10:00 on Monday. (UTC)”
    - cron: '0 10 * * 1'
  workflow_dispatch:
    inputs:
      packages:
        description: 'Targets to check'
        required: true
        type: string
        default: '${{ env.MY_MAINTAINED_PACKAGES }}'

jobs:
  check:
    runs-on: ubuntu-24.04 # Only support official Linux runners
    permissions: {} # If required any permission, please sending a report
    steps:
      - uses: DeterminateSystems/nix-installer-action@main
      - uses: DeterminateSystems/magic-nix-cache-action@main
      - uses: kachick/nixpkgs-package-health-check-action@main
        with:
          pnames: '${{ inputs.packages || env.MY_MAINTAINED_PACKAGES }}'
```

Cache helper is optional. However nixpkgs-update-log-checker takes minutes to build, it does not provide binary cache.

## Dependencies

- [hydra-check](https://github.com/nix-community/hydra-check)
- [nixpkgs-update-log-checker](https://github.com/kachick/nixpkgs-update-log-checker)

## Future Plans

- The current nixpkgs-update-log-checker doesn't check if a tool is the latest version available globally.\
  I may add support for this by integrating with [Repology](https://repology.org/).
- Simplify the process of checking packages for which users are maintainers.
- Add outputs to the action to facilitate easier notification integration.

## Scope

- This action does not support linting-related issues.
- This action does not check the entire Nixpkgs repository. It only checks for specified packages.
