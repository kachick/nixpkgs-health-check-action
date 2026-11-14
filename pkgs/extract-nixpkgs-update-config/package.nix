{
  writeShellApplication,
  jq,
  yj,
  coreutils,
}:
writeShellApplication {
  name = "extract-nixpkgs-update-config";
  runtimeInputs = [
    jq
    yj
    coreutils
  ];
  text = ''
    if [[ $# -ne 1 ]]; then
      echo "Usage: extract-nixpkgs-update-config <config-path>" >&2
      exit 1
    fi

    CONFIG="$1"
    if [[ ! -f "$CONFIG" ]]; then
      echo "Error: Config file '$CONFIG' not found or is not a regular file." >&2
      exit 1
    fi

    yj -tj < "$CONFIG" | jq -c '{ ignore: (.["nixpkgs-update"].ignore // []) }' | yj -jt
  '';
}
