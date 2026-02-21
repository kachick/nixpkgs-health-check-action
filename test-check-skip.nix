let
  checkSkip = import ./check-skip.nix;

  testCases = [
    # Case 1: Package with skip configuration (Should SKIP)
    {
      pname = "hello";
      service = "hydra";
      expectedSkip = true;
    }
    {
      pname = "hello";
      service = "nixpkgs-update";
      expectedSkip = true;
    }
    # Case 2: Package without skip configuration (Should RUN)
    {
      pname = "biz-ud-gothic";
      service = "hydra";
      expectedSkip = false;
    }
    {
      pname = "biz-ud-gothic";
      service = "nixpkgs-update";
      expectedSkip = false;
    }
  ];

  runTest =
    tc:
    let
      result = checkSkip {
        pname = tc.pname;
        service = tc.service;
      };
      status = if result.skip == tc.expectedSkip then "PASS" else "FAIL";
    in
    "${status}: ${tc.pname} [${tc.service}] -> skip=${toString result.skip} (reason: '${result.reason}')";

  results = map runTest testCases;
in
results
