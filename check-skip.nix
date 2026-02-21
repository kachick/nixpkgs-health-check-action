{ pname, service, configPath }:
let
  config = builtins.fromTOML (builtins.readFile (/. + configPath));
  skipConfig = if builtins.hasAttr "skip" config then config.skip else { };
  packageConfig = if builtins.hasAttr pname skipConfig then skipConfig.${pname} else { };
  hasService = builtins.hasAttr service packageConfig;
in
{
  skip = hasService;
  reason = if hasService then packageConfig.${service} else "";
}
