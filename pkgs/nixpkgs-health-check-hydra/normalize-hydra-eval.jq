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
      # 1. Attribute path matches (canonical or nested, e.g., python3Packages.magika-cli)
      (($attr == $query_pname) or ($attr | endswith("." + $query_pname))) and
      # 2. Exclude specialized hardware variants (e.g., pkgsRocm.magika-cli) unless explicitly requested.
      # This is safer than name matching because package naming can be inconsistent.
      (
        (($attr | test("rocm"; "i") | not) or ($query_pname | test("rocm"; "i"))) and
        (($attr | test("cuda"; "i") | not) or ($query_pname | test("cuda"; "i")))
      )
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
