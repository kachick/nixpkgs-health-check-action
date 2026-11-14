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

  packagesConfig = getAttrOrEmpty config "packages";
  packageConfig = getAttrOrEmpty packagesConfig pname;
  serviceConfig = getAttrOrEmpty packageConfig service;
  skipValue =
    if builtins.isAttrs serviceConfig && builtins.hasAttr "skip" serviceConfig then
      serviceConfig.skip
    else
      null;
  shouldSkip =
    if builtins.isString skipValue then
      skipValue != ""
    else if builtins.isBool skipValue then
      skipValue
    else
      false;
  reason =
    if !shouldSkip then
      ""
    else if builtins.isString skipValue then
      skipValue
    else
      "Skipped";
in
{
  skip = shouldSkip;
  inherit reason;
}
