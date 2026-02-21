let
  checkSkip = import ./check-skip.nix;

  # In real usage (via bash wrapper), configPath is always an absolute path string.
  # We simulate this by getting the absolute path of the current directory.
  pwd = toString ./.;

  testCases = [
    # Case 1: Package with skip configuration (Should SKIP)
    {
      pname = "hello";
      service = "hydra";
      configPath = "${pwd}/nixpkgs-health-check-by-maintainer.toml";
      expectedSkip = true;
    }
    {
      pname = "hello";
      service = "nixpkgs-update";
      configPath = "${pwd}/nixpkgs-health-check-by-maintainer.toml";
      expectedSkip = false;
    }
    # Case 2: Package without skip configuration (Should RUN)
    {
      pname = "biz-ud-gothic";
      service = "hydra";
      configPath = "${pwd}/nixpkgs-health-check-by-maintainer.toml";
      expectedSkip = false;
    }
    # Case 3: Simulation of the fix (absolute path string)
    {
      pname = "hello";
      service = "hydra";
      configPath = toString ./nixpkgs-health-check-by-maintainer.toml;
      expectedSkip = true;
    }
  ];

  runTest =
    tc:
    let
      result = checkSkip {
        inherit (tc) pname service configPath;
      };
      status = if result.skip == tc.expectedSkip then "PASS" else "FAIL";
    in
    "${status}: ${tc.pname} [${tc.service}] with ${tc.configPath} -> skip=${toString result.skip}";

  results = map runTest testCases;
in
results
