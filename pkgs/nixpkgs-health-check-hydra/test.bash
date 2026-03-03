#!/usr/bin/env bash

set -euo pipefail

# This script runs snapshots tests for normalize-hydra-eval.jq
# To update snapshots, run:
#   jq --from-file pkgs/nixpkgs-health-check-hydra/normalize-hydra-eval.jq pkgs/nixpkgs-health-check-hydra/test-snapshots/hello.json > pkgs/nixpkgs-health-check-hydra/test-snapshots/hello.expected.json
#   jq --from-file pkgs/nixpkgs-health-check-hydra/normalize-hydra-eval.jq pkgs/nixpkgs-health-check-hydra/test-snapshots/html2pdf.json > pkgs/nixpkgs-health-check-hydra/test-snapshots/html2pdf.expected.json

SCRIPTPATH="$(cd "$(dirname "$0")" && pwd)"
JQ_FILE="$SCRIPTPATH/normalize-hydra-eval.jq"
SNAPSHOT_DIR="$SCRIPTPATH/test-snapshots"

FAILED=0

for snap in "$SNAPSHOT_DIR"/*.json; do
    if [[ "$snap" == *.expected.json ]]; then
        continue
    fi
    
    name=$(basename "$snap" .json)
    expected="$SNAPSHOT_DIR/$name.expected.json"
    
    echo "Testing snapshot: $name"
    
    if ! diff <(jq --from-file "$JQ_FILE" "$snap") "$expected"; then
        echo "FAIL: Snapshot mismatch for $name"
        FAILED=1
    else
        echo "PASS: $name"
    fi
done

exit "$FAILED"
