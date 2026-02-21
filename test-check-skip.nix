let
  checkSkip = import ./check-skip.nix;

  testCases = [
    {
      pname = "pinact";
      check = "nixpkgs-update";
      expectedSkip = true;
    }
    {
      pname = "pinact";
      check = "hydra";
      expectedSkip = false;
    }
    {
      pname = "unknown";
      check = "hydra";
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
