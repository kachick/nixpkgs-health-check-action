#!/usr/bin/env bash

set -euo pipefail

# This script runs snapshots tests for normalize-hydra-eval.jq
# To update snapshots, run:
#   jq --from-file normalize-hydra-eval.jq test/snapshots/hello.json > test/snapshots/hello.expected.json
#   jq --from-file normalize-hydra-eval.jq test/snapshots/html2pdf.json > test/snapshots/html2pdf.expected.json

SCRIPTPATH="$(cd "$(dirname "$0")" && pwd)"
ROOTPATH="$SCRIPTPATH"
JQ_FILE="$ROOTPATH/normalize-hydra-eval.jq"

FAILED=0

for snap in "$ROOTPATH"/test/snapshots/*.json; do
    if [[ "$snap" == *.expected.json ]]; then
        continue
    fi
    
    name=$(basename "$snap" .json)
    expected="$ROOTPATH/test/snapshots/$name.expected.json"
    
    echo "Testing snapshot: $name"
    
    if ! diff <(jq --from-file "$JQ_FILE" "$snap") "$expected"; then
        echo "FAIL: Snapshot mismatch for $name"
        FAILED=1
    else
        echo "PASS: $name"
    fi
done

exit "$FAILED"
