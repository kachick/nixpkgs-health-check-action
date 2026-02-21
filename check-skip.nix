{
  pname,
  service,
  configPath,
}:
let
  # Ensure the path is treated correctly.
  # readFile accepts a string representing an absolute path.
  config = builtins.fromTOML (builtins.readFile configPath);

  # Helper to safely traverse the attribute set
  getAttrOrEmpty = s: k: if builtins.isAttrs s && builtins.hasAttr k s then s.${k} else { };

  skipConfig = getAttrOrEmpty config "skip";
  packageConfig = getAttrOrEmpty skipConfig pname;
  hasService = builtins.isAttrs packageConfig && builtins.hasAttr service packageConfig;
in
{
  skip = hasService;
  reason = if hasService then packageConfig.${service} else "";
}
