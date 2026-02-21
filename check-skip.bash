#!/usr/bin/env bash

set -euo pipefail

PNAME=""
SERVICE=""
CONFIG=""
# Resolve the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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
    --config)
      CONFIG="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1"
      exit 1
      ;;
  esac
done

if [[ -z "$PNAME" || -z "$SERVICE" || -z "$CONFIG" ]]; then
  echo "Usage: $0 --pname <name> --service <hydra|nixpkgs-update> --config <absolute-path>"
  exit 1
fi

# nix eval expects PNAME env var for builtins.getEnv
export PNAME="$PNAME"

nix eval --impure --json --expr "(import \"$SCRIPT_DIR/check-skip.nix\" { pname = builtins.getEnv \"PNAME\"; service = \"$SERVICE\"; configPath = \"$CONFIG\"; })"
