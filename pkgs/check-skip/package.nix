{
  writeShellApplication,
  nix,
  coreutils,
}:
writeShellApplication {
  name = "check-skip";
  runtimeInputs = [
    nix
    coreutils
  ];
  text = ''
    PNAME=""
    SERVICE=""
    CONFIG="nixpkgs-health-check-by-maintainer.toml"

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

    if [[ -z "$PNAME" || -z "$SERVICE" ]]; then
      echo "Usage: check-skip --pname <name> --service <hydra|nixpkgs-update> [--config <path>]"
      exit 1
    fi

    # Resolve config path to absolute path for Nix builtins.readFile
    if [[ "$CONFIG" != /* ]]; then
      CONFIG="$PWD/$CONFIG"
    fi

    if [[ ! -f "$CONFIG" ]]; then
      echo "Error: Config file '$CONFIG' not found or is not a regular file." >&2
      exit 1
    fi

    export PNAME="$PNAME"
    nix eval --impure --json --file ${./check-skip.nix} \
      --apply "f: f { pname = \"$PNAME\"; service = \"$SERVICE\"; configPath = \"$CONFIG\"; }"
  '';
}
