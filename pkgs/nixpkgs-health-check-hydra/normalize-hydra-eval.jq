[
  to_entries[]
  | select(.key | test("^[0-9]+/"))
  | (.key | sub("^[0-9]+/"; "")) as $query_pname
  | .value
  | [ .now_fail[], .still_fail[], .aborted[], .now_succeed[], .still_succeed[], .new[], .unfinished[] ]
  | .[]
  | select(.arch != null)
  | select(
      (.job_name | sub("\\.[^.]+$"; "")) as $attr |
      # Strictly match the attribute path to the query.
      # This naturally excludes specialized hardware variants (e.g., pkgsRocm.magika-cli)
      # and other integration tests (e.g., tests.haskell.ghcWithPackages.hello).
      $attr == $query_pname
    )
  | (
      if .status == "Succeeded" then "success"
      elif .status == "Dependency failed" or .status == "Queued" then "warning"
      else "error"
      end
    ) as $severity
  | {
      tool: .name,
      arch: .arch,
      status: .status,
      url: (.build_url // .url),
      severity: $severity,
      icon: { "success": "✔", "warning": "⚠", "error": "✖" }[$severity]
    }
]
