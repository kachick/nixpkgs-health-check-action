{
  pname,
  service,
  configPath,
}:
let
  # configPath is already an absolute path string from the bash wrapper.
  # Using it directly with readFile is safer than path concatenation in Nix.
  configContent = builtins.readFile configPath;
  config = builtins.fromTOML configContent;
  skipConfig = if builtins.hasAttr "skip" config then config.skip else { };
  packageConfig = if builtins.hasAttr pname skipConfig then skipConfig.${pname} else { };
  hasService = builtins.hasAttr service packageConfig;
in
{
  skip = hasService;
  reason = if hasService then packageConfig.${service} else "";
}
