# nixpkgs-health-check-action

[![CI - Nix Status](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/ci-nix.yml/badge.svg?branch=main)](https://github.com/kachick/nixpkgs-health-check-action/actions/workflows/ci-nix.yml?query=branch%3Amain+)

This GitHub Action helps ensure that specific Nixpkgs packages are buildable and up-to-date.

## Usage

This repository has [simple](.github/workflows/health-check.yml) reusable workflow for my personal use.\
I recommend you write own workflow in your repository.\
However below my use-case might be an your example.

```yaml
name: by-maintainer # Keep short for the GitHub Web UI

on:
  schedule:
    # e.g: “At 10:00 on Monday. (UTC)”
    - cron: '0 10 * * 1'
  workflow_dispatch:
    inputs:
      maintainer:
        description: 'Maintainer ID of the package'
        required: true
        type: string
        default: 'SET_YOUR_ID'

permissions:
  contents: read

jobs:
  get-pnames:
    runs-on: ubuntu-24.04
    env:
      MAINTAINER: '${{ inputs.maintainer }}'
    outputs:
      json: '${{ steps.print-pnames-as-json.outputs.json }}'
    steps:
      - uses: cachix/install-nix-action@4e002c8ec80594ecd40e759629461e26c8abed15 # v31.9.0
        with:
          nix_path: 'nixpkgs=channel:nixpkgs-unstable'
          # Enable accept-flake-config for using the binary cache
          extra_nix_config: |
            sandbox = true
            accept-flake-config = true
      - id: print-pnames-as-json
        run: |
          {
            echo 'report<<JSON'
            nix run github:kachick/nixpkgs-maintained-by -- -id "$MAINTAINER" -json
            echo 'JSON'
          } 2>/dev/null | tee --append "$GITHUB_OUTPUT"

  check:
    needs: [get-pnames]
    strategy:
      matrix:
        pname: '${{ fromJson(needs.get-pnames.outputs.json) }}'
    uses: kachick/nixpkgs-health-check-action/.github/workflows/health-check.yml@main
    with:
      pname: '${{ matrix.pname }}'
```

## Dependencies

- [hydra-check](https://github.com/nix-community/hydra-check)
- [nixpkgs-update-log-checker](https://github.com/kachick/nixpkgs-update-log-checker)
- [nixpkgs-maintained-by](https://github.com/kachick/nixpkgs-maintained-by)

## Scope

- Does not check the entire Nixpkgs repository. It only checks for specified packages.
- Does not check the latest version of upstream. It only checks [nixpkgs-update](https://github.com/nix-community/nixpkgs-update)'s [results](https://nixpkgs-update-logs.nix-community.org/).
