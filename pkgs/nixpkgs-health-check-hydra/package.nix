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
  text = ''
    PNAME="$1"

    mkdir -p tmp
    hydra-check --json --channel=nixpkgs-unstable --eval "$PNAME" | tee tmp/raw_hydra.json
    jq --from-file ${./normalize-hydra-eval.jq} tmp/raw_hydra.json | tee tmp/normalized_hydra.json

    if jq -e '.[] | select(.severity == "error")' tmp/normalized_hydra.json > tmp/hydra_errors.jsonl; then
      echo "::error::Hydra check failed for $PNAME"
      cat tmp/hydra_errors.jsonl
      exit 1
    fi
  '';
}
