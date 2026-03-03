{
  pname,
  service,
  configPath,
}:
let
  # The root cause was path resolution.
  # configPath must be an absolute path string starting with '/'.
  # Using it directly with readFile avoids Nix's path concatenation pitfalls.
  config = builtins.fromTOML (builtins.readFile configPath);

  # Basic structure validation to prevent crashes on malformed TOML
  getAttrOrEmpty = s: k: if builtins.isAttrs s && builtins.hasAttr k s then s.${k} else { };

  skipConfig = getAttrOrEmpty config "skip";
  packageConfig = getAttrOrEmpty skipConfig pname;
  hasService = builtins.isAttrs packageConfig && builtins.hasAttr service packageConfig;
in
{
  skip = hasService;
  reason = if hasService then packageConfig.${service} else "";
}
