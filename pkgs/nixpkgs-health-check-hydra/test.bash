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

echo "Testing ignore_urls filter"
ignored_url="https://hydra.nixos.org/build/201"
other_url="https://hydra.nixos.org/build/202"
sample_json=$(
	cat <<EOF
{
  "1/foo": {
    "now_fail": [
      { "name": "foo-1.0", "job_name": "foo.x86_64-linux", "arch": "x86_64-linux", "status": "Failed", "build_url": "$ignored_url" },
      { "name": "foo-1.0", "job_name": "foo.aarch64-linux", "arch": "aarch64-linux", "status": "Failed", "build_url": "$other_url" }
    ],
    "still_fail": [],
    "aborted": [],
    "now_succeed": [],
    "still_succeed": [],
    "new": [],
    "unfinished": []
  }
}
EOF
)
filtered=$(printf '%s' "$sample_json" | jq --argjson ignore_urls "[\"$ignored_url\"]" --from-file "$JQ_FILE")
if ! printf '%s' "$filtered" | jq -e --arg u "$ignored_url" '.[] | select(.url == $u and .severity == "warning" and .icon == "⚠")' >/dev/null; then
	echo "FAIL: Expected $ignored_url to be downgraded to warning"
	FAILED=1
elif ! printf '%s' "$filtered" | jq -e --arg u "$other_url" '.[] | select(.url == $u and .severity == "error" and .icon == "✖")' >/dev/null; then
	echo "FAIL: Expected $other_url to remain error"
	FAILED=1
else
	echo "PASS: ignore_urls filter"
fi

exit "$FAILED"
