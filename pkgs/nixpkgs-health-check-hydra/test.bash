#!/usr/bin/env bash

set -euo pipefail

SCRIPTPATH="$(cd "$(dirname "$0")" && pwd)"
JQ_FILE="$SCRIPTPATH/normalize-hydra-eval.jq"

FAILED=0

# Iterate through both test-snapshots (real data) and fixtures (synthetic data)
for dir in "$SCRIPTPATH/test-snapshots" "$SCRIPTPATH/fixtures"; do
    if [ ! -d "$dir" ]; then continue; fi

    for snap in "$dir"/*.json; do
        if [[ "$snap" == *.expected.json ]]; then
            continue
        fi

        name=$(basename "$snap" .json)
        expected="$dir/$name.expected.json"

        echo "Testing $dir: $name"

        if ! diff <(jq --from-file "$JQ_FILE" "$snap") "$expected"; then
            echo "FAIL: Snapshot mismatch for $name in $dir"
            FAILED=1
        else
            echo "PASS: $name"
        fi
    done
done

exit "$FAILED"
