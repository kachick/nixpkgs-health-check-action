{ pname, check }:
let
  config = builtins.fromTOML (builtins.readFile ./health-check-skips.toml);
  packageConfig = if builtins.hasAttr pname config then config.${pname} else { };
  key = "skip_${check}";
in
if builtins.hasAttr key packageConfig then packageConfig.${key} else false
