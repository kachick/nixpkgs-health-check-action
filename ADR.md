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

## 2. Combined Logic in Composite Action (`action.yml`)

### Context

We debated whether to split skip-logic (Planning) and execution (Runner).

### Decision

Keep the core execution and skip-check logic within a single Composite Action at the root.

### Why not separate 'Plan' Action?

- **Asset Access**: Reusable workflows don't automatically mount the action's scripts. Accessing `check-skip.bash` from a workflow requires fragile `actions/checkout` hacks with specific SHAs.
- **Standardization**: By using `github.action_path` inside a Composite Action, we guarantee robust path resolution across any caller repository without hacks.

## 3. Two-Layer Workflow Structure

### Context

Should we have a single workflow for everything?

### Decision

Keep `health-check.yml` (Generic List Runner) and `health-check-by-maintainer.yml` (Maintainer Adapter) separate.

### Why?

- **Validation**: Each workflow can have its own `required: true` inputs. A combined workflow would make input validation ambiguous.
- **Separation of Concerns**: The runner focuses on "how to check a list," while the adapter focuses on "how to get the list for a person."
