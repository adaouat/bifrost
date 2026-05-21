# Workflow rules

## Roadmap process

Every code task follows a strict three-step commit flow. Never deviate from it.

1. **Start** — mark the task `[~]` in `docs/tasks/roadmap.md`, commit the roadmap change alone.
2. **Implement** — do the work (TDD: failing test first, then implementation).
3. **Done** — mark the task `[x]` in `docs/tasks/roadmap.md`, commit implementation + roadmap update together.

One task at a time. Never start a second task before the first is committed as `[x]`.
Never implement anything that is not a current roadmap task without explicit user approval.

## Commits

Conventional commits enforced by `hk` + `cocogitto`. Valid types:
`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `style`, `perf`, `ci`, `build`

Scope is the affected package or command, e.g. `feat(artifact): add --release-name flag`.

Never skip pre-commit hooks (`--no-verify` is forbidden).

## Linting

Run `hk check` before every commit. All issues must be resolved first.

For targeted auto-fixes:
- `hk fix -S <linter>` — fix one linter (e.g. `hk fix -S golangci-lint`, `hk fix -S yamlfmt`)
- `hk fix` — fix everything at once

Never commit code that does not pass `hk check`.

## Versions

Pin exact versions everywhere — no `latest` string in mise config, go.mod, or CI workflows.

**Exceptions** — these tools use `latest` intentionally (they are format/lint utilities
with no API surface that could break the build):
`pkl`, `tombi`, `typos`, `yamlfmt`
