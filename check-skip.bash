#!/usr/bin/env bash

set -euo pipefail

PNAME=""
SERVICE=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --pname)
      PNAME="$2"
      shift 2
      ;;
    --service)
      SERVICE="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1"
      exit 1
      ;;
  esac
done

if [[ -z "$PNAME" ]]; then
  PNAME="${PNAME:-}"
fi

if [[ -z "$PNAME" || -z "$SERVICE" ]]; then
  echo "Usage: $0 --pname <name> --service <hydra|nixpkgs-update>"
  exit 1
fi

# nix eval expects PNAME env var for builtins.getEnv
export PNAME="$PNAME"

nix eval --impure --json --expr "(import ./check-skip.nix { pname = builtins.getEnv \"PNAME\"; service = \"$SERVICE\"; })"
