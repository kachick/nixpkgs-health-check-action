# Architecture Decision Record (ADR)

This document records the architectural decisions made during the evolution of this repository.

## 1. Single-Dimensional Matrix (Package-Centric)

### Context

We wanted to run checks for each package across multiple services (`hydra`, `nixpkgs-update`).

### Decision

Use a single-dimensional matrix of packages (`pname`). Run services sequentially as steps within a single Job.

### Why not 2D Matrix (pname x service)?

- **UI Limitation**: GitHub Actions sidebar (tree view) prioritizes matrix variable values over the `name:` property. A 2D matrix results in ambiguous labels or forces ugly string hacks in matrix values.
- **Scalability**: Job slots are limited. 2D matrix doubles the Job count, hitting concurrency limits faster.
- **Readability**: Sequential steps within a job provide a cleaner summary in the execution detail view.

## 2. Decoupled Planning and Execution

### Context

We debated where the skip-logic (TOML checking) should reside: inside the Runner Action or in a separate Planning step.

### Decision

Move all decision-making logic (parsing skip-lists) into the `plan` job of the Reusable Workflow. The core Composite Action (`action.yml`) remains a "dumb" runner that simply executes what it is told to do via boolean inputs.

### Why?

- **Predictability**: When a workflow calls the action with `hydra: true`, the action will execute Hydra. There is no "hidden" skip logic inside the action that might override the caller's intent.
- **Efficiency**: Skip-list (TOML) is parsed once in the `plan` job to build the execution matrix, rather than being parsed repeatedly in every matrix job.
- **UI Clarity**: If a package is skipped for all services, its Job is never even created in the GitHub UI, keeping the results view clean.
- **Single Responsibility**: The Action focuses on "how to check," while the Workflow focuses on "who to check."

## 3. Two-Layer Workflow Structure

### Context

Should we have a single workflow for everything?

### Decision

Keep `health-check.yml` (Generic List Runner) and `health-check-by-maintainer.yml` (Maintainer Adapter) separate.

### Why?

- **Validation**: Each workflow can have its own `required: true` inputs. A combined workflow would make input validation ambiguous.
- **Separation of Concerns**: The runner focuses on "how to check a list," while the adapter focuses on "how to get the list for a person."
