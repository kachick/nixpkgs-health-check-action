{ pname, check }:
let
  config = builtins.fromTOML (builtins.readFile ./health-check-skips.toml);
  packageConfig = if builtins.hasAttr pname config then config.${pname} else { };
  hasCheck = builtins.hasAttr check packageConfig;
in
{
  skip = hasCheck;
  reason = if hasCheck then packageConfig.${check} else "";
}
