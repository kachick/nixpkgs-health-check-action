[
  to_entries[]
  | select(.key | test("^[0-9]+/"))
  | .key as $tool
  | .value
  | [ .now_fail[], .still_fail[], .aborted[], .now_succeed[], .still_succeed[], .new[], .unfinished[] ]
  | .[]
  | select(.arch != null)
  | (
      if .status == "Succeeded" then "success"
      elif .status == "Dependency failed" or .status == "Queued" then "warning"
      else "error"
      end
    ) as $severity
  | {
      tool: ($tool | sub("^[0-9]+/"; "")),
      arch: .arch,
      status: .status,
      url: (.build_url // .url),
      severity: $severity,
      icon: { "success": "✔", "warning": "⚠", "error": "✖" }[$severity]
    }
]
