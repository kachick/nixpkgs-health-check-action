# nixpkgs-package-health-check-action

[![CI - Nix Status](https://github.com/kachick/nixpkgs-package-health-check-action/actions/workflows/ci-nix.yml/badge.svg?branch=main)](https://github.com/kachick/nixpkgs-package-health-check-action/actions/workflows/ci-nix.yml?query=branch%3Amain+)

This GitHub Action helps ensure that specific Nixpkgs packages are buildable and up-to-date.

## Dependencies

- [hydra-check](https://github.com/nix-community/hydra-check)
- [nixpkgs-update-log-checker](https://github.com/kachick/nixpkgs-update-log-checker)

## Future Plans

- The current nixpkgs-update-log-checker doesn't check if a tool is the latest version available globally.\
  I may add support for this by integrating with [Repology](https://repology.org/).

- Simplify the process of checking packages for which users are maintainers.

## Scope

- This action does not support linting-related issues.
- This action does not check the entire Nixpkgs repository. It only checks for specified packages.
