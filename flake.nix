{
  nixConfig = {
    extra-substituters = [
      "https://cache.garnix.io"
    ];
    extra-trusted-public-keys = [
      "cache.garnix.io:CTFPyKSLcx5RMJKfLo5EEPUObbA78b0YQ2DTCJXqr9g="
    ];
  };

  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.xz";
  };

  outputs =
    {
      self,
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
        pkgs.lib.filesystem.packagesFromDirectoryRecursive {
          inherit (pkgs) callPackage;
          directory = ./pkgs;
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
                gitleaks
              ]
            );
          };
        }
      );
    };
}
