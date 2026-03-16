# nixpkgs-health-check-action

[![CI - Nix Status](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/ci-nix.yml/badge.svg?branch=main)](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/ci-nix.yml?query=branch%3Amain+)

This GitHub Action helps ensure that specific Nixpkgs packages are buildable and up-to-date.

## Usage

### Live example (Dogfooding)

This repository itself uses this action to check my maintained packages.

[Workflow](.github/workflows/dogfood-maintainer.yml): [![🐶 Status](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/dogfood-maintainer.yml/badge.svg?branch=main)](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/dogfood-maintainer.yml?query=branch%3Amain+)

### Check single package

You can use the composite action to check a single package.

```yaml
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: NixOS/nix-installer-action@main
      - uses: kachick/nixpkgs-health-check-action@main
        with:
          pname: 'hello'
```

### Check packages by list or maintainer

You can use the reusable workflows to check multiple packages.
They automatically parallelize checks for each package.

```yaml
jobs:
  check:
    # Use health-check.yml for a specific list of packages
    uses: kachick/nixpkgs-health-check-action/.github/workflows/health-check.yml@main
    with:
      pnames: '["hello", "biz-ud-gothic"]'
```

```yaml
jobs:
  check:
    # Use health-check-by-maintainer.yml for all packages of a person
    uses: kachick/nixpkgs-health-check-action/.github/workflows/health-check-by-maintainer.yml@main
    with:
      maintainer: 'kachick'
```

## Advanced Configuration: Skipping Checks

You can skip specific checks for packages with known upstream issues by creating a configuration file (default: `nixpkgs-health-check-by-maintainer.toml`) in your repository.

### Example configuration

```toml
[skip]
# package-name = { service-name = "Reason for skipping" }
typescript-go = { nixpkgs-update = "Failing in nixpkgs-update, see issue #..." }
hello = { hydra = "Just for testing", nixpkgs-update = "Just for testing" }
```

### Usage with custom config path

If you use a custom filename or store the config in a subdirectory, you must check out your repository first and provide the path.

```yaml
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6 # Required to read the config file
      - uses: kachick/nixpkgs-health-check-action/.github/workflows/health-check-by-maintainer.yml@main
        with:
          maintainer: 'your-id'
          config: '.github/my-skip-list.toml' # Specify your config path
```

## Dependencies

- [hydra-check](https://github.com/nix-community/hydra-check)
- [nixpkgs-update-log-checker](https://github.com/kachick/nixpkgs-update-log-checker)
- [nixpkgs-maintained-by](https://github.com/kachick/nixpkgs-maintained-by)

## Scope

- Does not check the entire Nixpkgs repository. It only checks for specified packages.
- Does not check the latest version of upstream. It only checks [nixpkgs-update](https://github.com/nix-community/nixpkgs-update)'s [results](https://nixpkgs-update-logs.nix-community.org/).
