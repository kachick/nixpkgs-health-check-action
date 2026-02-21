let
  checkSkip = import ./check-skip.nix;

  testCases = [
    # Case 1: Package with skip configuration (Should SKIP)
    {
      pname = "hello";
      check = "hydra";
      expectedSkip = true;
    }
    {
      pname = "hello";
      check = "nixpkgs-update";
      expectedSkip = true;
    }
    # Case 2: Package without skip configuration (Should RUN)
    {
      pname = "biz-ud-gothic";
      check = "hydra";
      expectedSkip = false;
    }
    {
      pname = "biz-ud-gothic";
      check = "nixpkgs-update";
      expectedSkip = false;
    }
  ];

  runTest =
    tc:
    let
      result = checkSkip {
        pname = tc.pname;
        check = tc.check;
      };
      status = if result.skip == tc.expectedSkip then "PASS" else "FAIL";
    in
    "${status}: ${tc.pname} [${tc.check}] -> skip=${toString result.skip} (reason: '${result.reason}')";

  results = map runTest testCases;
in
results
