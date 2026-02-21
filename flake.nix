{
  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.xz";
  };

  outputs =
    {
      nixpkgs,
      ...
    }:
    let
      inherit (nixpkgs) lib;
      forAllSystems = lib.genAttrs lib.systems.flakeExposed;
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          check-skip = pkgs.writeShellApplication {
            name = "check-skip";
            runtimeInputs = [ pkgs.nix ];
            text = ''
              # Simplified logic of check-skip.bash, now part of the flake
              PNAME=""
              SERVICE=""
              CONFIG=""
              while [[ $# -gt 0 ]]; do
                case $1 in
                  --pname) PNAME="$2"; shift 2 ;;
                  --service) SERVICE="$2"; shift 2 ;;
                  --config) CONFIG="$2"; shift 2 ;;
                  *) echo "Unknown argument: $1"; exit 1 ;;
                esac
              done
              if [[ -z "$PNAME" || -z "$SERVICE" || -z "$CONFIG" ]]; then
                echo "Usage: check-skip --pname <name> --service <hydra|nixpkgs-update> --config <path>" >&2
                exit 1
              fi
              # Resolve to absolute path
              if [[ "$CONFIG" != /* ]]; then CONFIG="$PWD/$CONFIG"; fi
              if [[ ! -f "$CONFIG" ]]; then echo "Error: '$CONFIG' not found" >&2; exit 1; fi

              nix eval --impure --json --expr "(import ${./check-skip.nix} { pname = \"$PNAME\"; service = \"$SERVICE\"; configPath = \"$CONFIG\"; })"
            '';
          };

          normalize-hydra = pkgs.writeShellApplication {
            name = "normalize-hydra";
            runtimeInputs = [ pkgs.jq ];
            text = ''
              jq --from-file ${./normalize-hydra-eval.jq} "$@"
            '';
          };
        }
      );

      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt-tree);
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShellNoCC {
            # Correct nixd inlay hints
            env.NIX_PATH = "nixpkgs=${nixpkgs.outPath}";

            buildInputs = (
              with pkgs;
              [
                # https://github.com/NixOS/nix/issues/730#issuecomment-162323824
                bashInteractive
                findutils # xargs
                nixfmt
                nixfmt-tree
                nixd
                go-task
                hydra-check

                dprint
                typos
                zizmor

                shfmt
                shellcheck
              ]
            );
          };
        }
      );
    };
}
