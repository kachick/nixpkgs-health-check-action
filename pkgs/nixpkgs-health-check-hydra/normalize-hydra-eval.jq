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
      # This matches canonical packages on ALL platforms (e.g. magika-cli.x86_64-linux, magika-cli.aarch64-darwin)
      # while ignoring nested, specialized, and test variants (unless they are explicitly queried).
      $attr == $query_pname
    )
  | (
      if .status == "Succeeded" or .status == "Queued" then "success"
      else
        if .arch == "x86_64-darwin" or .status == "Dependency failed" or .status == "Cancelled" or .status == "Aborted" then "warning"
        else "error"
        end
      end
    ) as $severity
  | {
      tool: .name,
      attr: (.job_name | sub("\\.[^.]+$"; "")),
      arch: .arch,
      status: .status,
      url: (.build_url // .url),
      severity: $severity,
      icon: { "success": "✔", "warning": "⚠", "error": "✖" }[$severity]
    }
]
