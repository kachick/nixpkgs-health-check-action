{
  lib,
  buildGoModule,
  makeWrapper,
  hydra-check,
}:
buildGoModule {
  pname = "nixpkgs-health-check";
  version = "0.1.0";

  src = lib.fileset.toSource {
    root = ../..;
    fileset = lib.fileset.unions [
      ../../go.mod
      ../../go.sum
      ../../cmd
      ../../internal
      ../../nixpkgs-health-check-by-maintainer.toml
    ];
  };

  vendorHash = "sha256-pbA/AlBz3cQYRTMnQ/qBPcinYOKokrBLNhkbRTq54gE=";

  subPackages = [ "cmd/nixpkgs-health-check" ];

  env.CGO_ENABLED = 0;

  nativeBuildInputs = [ makeWrapper ];

  # Tests live under internal/, so run go test across all packages rather than only subPackages.
  checkPhase = ''
    runHook preCheck
    go test ./...
    runHook postCheck
  '';

  postInstall = ''
    wrapProgram $out/bin/nixpkgs-health-check \
      --prefix PATH : ${lib.makeBinPath [ hydra-check ]}
  '';

  meta = {
    description = "Health checker for Nixpkgs packages with GitHub Dashboard Issue reporting";
    mainProgram = "nixpkgs-health-check";
  };
}
