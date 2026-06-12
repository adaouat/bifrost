# Workflow rules

## Roadmap process

Every code task follows a strict two-step commit flow. Never deviate from it.

1. **Implement** — do the work (TDD: failing test first, then implementation).
2. **Done** — mark the task `[x]` in `docs/tasks/roadmap.md`, commit implementation + roadmap update together.

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
- `hk fix -S <linter>` — fix one linter (e.g. `hk fix -S golangci_lint`, `hk fix -S yamlfmt`)
- `hk fix` — fix everything at once

Never commit code that does not pass `hk check`.

## Versions

Pin exact versions everywhere — no `latest` string in mise config, go.mod, or CI workflows.

**Exceptions** — these tools use `latest` intentionally (they are format/lint utilities
or dev/editor tooling with no API surface that could break the build):
`pkl`, `tombi`, `typos`, `yamlfmt`, `gopls` (and similar LSP/editor tools)

## GitHub Actions

Pin every action to a full commit SHA, never to a mutable tag. Add the semantic
version as a comment so the intent is still readable:

```yaml
uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6
```

To update an action, find the new SHA for the desired tag on GitHub
(`github.com/<owner>/<action>/tags`) and replace both the SHA and the comment.
Never use `@v4`, `@main`, or `@latest` — these are mutable and can be hijacked.

## Charmbracelet dependencies

All charmbracelet packages use the `charm.land` module registry, not `github.com/charmbracelet`.

```
go get charm.land/<module>/v2   # e.g. charm.land/huh/v2, charm.land/bubbles/v2
```

Never add `github.com/charmbracelet/<module>` as a direct dependency.
