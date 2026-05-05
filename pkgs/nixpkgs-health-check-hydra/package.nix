{
  writeShellApplication,
  hydra-check,
  jq,
  coreutils,
}:
writeShellApplication {
  name = "nixpkgs-health-check-hydra";
  runtimeInputs = [
    hydra-check
    jq
    coreutils
  ];
  # hydra-check 2.1.0 returns 1 if there are any build failures in the evaluation.
  # We ignore this exit code because we categorize failures (e.g. Dependency failed)
  # as warnings in the next jq step.
  # See https://github.com/nix-community/hydra-check/pull/94
  text = ''
    PNAME="$1"

    mkdir -p tmp
    hydra-check --json --channel=nixpkgs-unstable --eval "$PNAME" > tmp/raw_hydra.json || true
    cat tmp/raw_hydra.json
    jq --from-file ${./normalize-hydra-eval.jq} tmp/raw_hydra.json | tee tmp/normalized_hydra.json

    if jq -e '.[] | select(.severity == "warning")' tmp/normalized_hydra.json > tmp/hydra_warnings.jsonl; then
      echo "::warning::Hydra check found warnings for $PNAME"
      cat tmp/hydra_warnings.jsonl
    fi

    if jq -e '.[] | select(.severity == "error")' tmp/normalized_hydra.json > tmp/hydra_errors.jsonl; then
      echo "::error::Hydra check failed for $PNAME"
      cat tmp/hydra_errors.jsonl
      exit 1
    fi
  '';
}
