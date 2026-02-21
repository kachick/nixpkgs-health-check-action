# nixpkgs-health-check-action

[![CI - Nix Status](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/ci-nix.yml/badge.svg?branch=main)](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/ci-nix.yml?query=branch%3Amain+)

This GitHub Action helps ensure that specific Nixpkgs packages are buildable and up-to-date.

## Usage

### Check single package

You can use the composite actions to check a single package.

```yaml
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: cachix/install-nix-action@v31
      - uses: kachick/nixpkgs-health-check-action/.github/actions/hydra@main
        with:
          pname: 'hello'
```

### Check all packages of a maintainer (with skip-list support)

You can use the reusable workflow to check all packages maintained by a specific person.
It automatically parallelizes checks for each package and service.

```yaml
jobs:
  check:
    uses: kachick/nixpkgs-health-check-action/.github/workflows/health-check-by-maintainer.yml@main
    with:
      maintainer: 'kachick'
```

## Dependencies

- [hydra-check](https://github.com/nix-community/hydra-check)
- [nixpkgs-update-log-checker](https://github.com/kachick/nixpkgs-update-log-checker)
- [nixpkgs-maintained-by](https://github.com/kachick/nixpkgs-maintained-by)

## Scope

- Does not check the entire Nixpkgs repository. It only checks for specified packages.
- Does not check the latest version of upstream. It only checks [nixpkgs-update](https://github.com/nix-community/nixpkgs-update)'s [results](https://nixpkgs-update-logs.nix-community.org/).

## Advanced Configuration

### Skipping Checks

You can skip specific checks for packages with known upstream issues by creating a `nixpkgs-health-check-by-maintainer.toml` file in the root of your repository.
This is used by the `health-check-by-maintainer.yml` workflow.

```toml
[skip]
# package-name = { service-name = "Reason for skipping" }
typescript-go = { nixpkgs-update = "Failing in nixpkgs-update, see issue #..." }
```
