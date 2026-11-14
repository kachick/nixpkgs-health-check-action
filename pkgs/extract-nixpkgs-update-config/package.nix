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
    PNAME=""
    PACKAGE_ONLY=false
    CONFIG=""

    while [[ $# -gt 0 ]]; do
      case $1 in
      --pname)
        if [[ $# -lt 2 ]]; then
          echo "Error: --pname requires a package name" >&2
          exit 1
        fi
        PNAME="$2"
        shift 2
        ;;
      --package-only)
        PACKAGE_ONLY=true
        shift
        ;;
      -*)
        echo "Unknown option: $1" >&2
        exit 1
        ;;
      *)
        if [[ -n "$CONFIG" ]]; then
          echo "Unexpected argument: $1" >&2
          exit 1
        fi
        CONFIG="$1"
        shift
        ;;
      esac
    done

    if [[ -z "$CONFIG" ]]; then
      echo "Usage: extract-nixpkgs-update-config [--pname <pname>] [--package-only] <config-path>" >&2
      exit 1
    fi

    if [[ "$PACKAGE_ONLY" == "true" && -z "$PNAME" ]]; then
      echo "Error: --package-only requires --pname" >&2
      exit 1
    fi

    if [[ ! -f "$CONFIG" ]]; then
      echo "Error: Config file '$CONFIG' not found or is not a regular file." >&2
      exit 1
    fi

    if [[ "$PACKAGE_ONLY" == "true" ]]; then
      yj -tj < "$CONFIG" | jq -c --arg p "$PNAME" '{ ignore: (.packages[$p]["nixpkgs-update"].ignore // []) }' | yj -jt
    elif [[ -n "$PNAME" ]]; then
      yj -tj < "$CONFIG" | jq -c --arg p "$PNAME" '{ ignore: ((.global["nixpkgs-update"].ignore // []) + (.packages[$p]["nixpkgs-update"].ignore // [])) }' | yj -jt
    else
      yj -tj < "$CONFIG" | jq -c '{ ignore: ((.global["nixpkgs-update"].ignore // []) + ([(.packages // {})[] | .["nixpkgs-update"]?.ignore // []] | add // [])) }' | yj -jt
    fi
  '';
}
