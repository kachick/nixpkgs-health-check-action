{
  writeShellApplication,
  hydra-check,
  jq,
  yj,
  coreutils,
}:
writeShellApplication {
  name = "nixpkgs-health-check-hydra";
  runtimeInputs = [
    hydra-check
    jq
    yj
    coreutils
  ];
  # hydra-check 2.1.0 returns 1 if there are any build failures in the evaluation.
  # We ignore this exit code because we categorize failures (e.g. Dependency failed)
  # as warnings in the next jq step.
  # See https://github.com/nix-community/hydra-check/pull/94
  text = ''
    PNAME=""
    CONFIG=""

    while [[ $# -gt 0 ]]; do
      case $1 in
      --config)
        if [[ $# -lt 2 ]]; then
          echo "Error: --config requires a file path" >&2
          exit 1
        fi
        CONFIG="$2"
        shift 2
        ;;
      -*)
        echo "Unknown option: $1" >&2
        exit 1
        ;;
      *)
        if [[ -n "$PNAME" ]]; then
          echo "Unexpected argument: $1" >&2
          exit 1
        fi
        PNAME="$1"
        shift
        ;;
      esac
    done

    if [[ -z "$PNAME" ]]; then
      echo "Usage: nixpkgs-health-check-hydra [--config <path>] <pname>" >&2
      exit 1
    fi

    PACKAGE_IGNORE_URLS="[]"
    GLOBAL_IGNORE_URLS="[]"
    if [[ -n "$CONFIG" ]]; then
      if [[ ! -f "$CONFIG" ]]; then
        echo "Error: Config file '$CONFIG' not found or is not a regular file." >&2
        exit 1
      fi
      config_json=$(yj -tj < "$CONFIG")
      PACKAGE_IGNORE_URLS=$(printf '%s' "$config_json" | jq -c --arg p "$PNAME" '(.packages[$p].hydra.ignore // []) | map(if (.url | type) == "string" and (.url | length) > 0 then .url else error("packages." + $p + ".hydra.ignore entry requires a non-empty string url") end)')
      GLOBAL_IGNORE_URLS=$(printf '%s' "$config_json" | jq -c '(.global.hydra.ignore // []) | map(if (.url | type) == "string" and (.url | length) > 0 then .url else error("global.hydra.ignore entry requires a non-empty string url") end)')
    fi
    IGNORE_URLS=$(jq -cn --argjson a "$PACKAGE_IGNORE_URLS" --argjson b "$GLOBAL_IGNORE_URLS" '$a + $b | unique')

    mkdir -p tmp
    hydra-check --json --channel=nixpkgs-unstable --eval "$PNAME" > tmp/raw_hydra.json || true
    cat tmp/raw_hydra.json
    jq --argjson ignore_urls "$IGNORE_URLS" --from-file ${./normalize-hydra-eval.jq} tmp/raw_hydra.json | tee tmp/normalized_hydra.json

    if jq -e '.[] | select(.severity == "warning")' tmp/normalized_hydra.json > tmp/hydra_warnings.jsonl; then
      echo "::warning::Hydra check found warnings for $PNAME"
      cat tmp/hydra_warnings.jsonl
      if [[ -n "''${GITHUB_STEP_SUMMARY:-}" ]]; then
        {
          echo "### Hydra Warnings (\`$PNAME\`)"
          echo ""
          echo "| Attr | Arch | Status | URL |"
          echo "| --- | --- | --- | --- |"
          jq -r '.[] | select(.severity == "warning") | "| `\(.attr)` | `\(.arch)` | \(.status) | \(.url) |"' tmp/normalized_hydra.json
          echo ""
        } >> "$GITHUB_STEP_SUMMARY"
      fi
    fi

    if [[ "$PACKAGE_IGNORE_URLS" != "[]" ]] && jq -e 'length > 0' tmp/normalized_hydra.json > /dev/null; then
      unused_urls=$(jq -r --argjson pkg_urls "$PACKAGE_IGNORE_URLS" '([.[].url]) as $seen | $pkg_urls[] | select(($seen | index(.)) == null)' tmp/normalized_hydra.json)
      if [[ -n "$unused_urls" ]]; then
        echo "::warning::Unused Hydra ignore rules for $PNAME (config may be stale)"
        printf '%s\n' "$unused_urls"
        if [[ -n "''${GITHUB_STEP_SUMMARY:-}" ]]; then
          {
            echo "### Stale Config Warning: Hydra (\`$PNAME\`)"
            echo ""
            echo "The following \`packages.$PNAME.hydra.ignore\` URLs did not match any current Hydra build:"
            echo ""
            while IFS= read -r url; do
              echo "- $url"
            done <<< "$unused_urls"
            echo ""
          } >> "$GITHUB_STEP_SUMMARY"
        fi
      fi
    fi

    if jq -e '.[] | select(.severity == "error")' tmp/normalized_hydra.json > tmp/hydra_errors.jsonl; then
      echo "::error::Hydra check failed for $PNAME"
      cat tmp/hydra_errors.jsonl
      exit 1
    fi
  '';
}
