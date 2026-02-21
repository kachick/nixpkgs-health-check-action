{ pname, check }:
let
  config = builtins.fromTOML (builtins.readFile ./nixpkgs-health-check.toml);
  skipConfig = if builtins.hasAttr "skip" config then config.skip else { };
  packageConfig = if builtins.hasAttr pname skipConfig then skipConfig.${pname} else { };
  hasCheck = builtins.hasAttr check packageConfig;
in
{
  skip = hasCheck;
  reason = if hasCheck then packageConfig.${check} else "";
}
