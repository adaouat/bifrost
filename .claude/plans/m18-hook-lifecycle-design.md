# Hook lifecycle: symmetric pre/post hooks per deploy stage

## Status

Proposed — pending implementation plan.

## Context

The current `Hooks` struct (`internal/config/schema.go`) defines four lifecycle hooks:

```go
type Hooks struct {
	PostExtract       []HookEntry `yaml:"post_extract"`
	PreLink           []HookEntry `yaml:"pre_link"`
	PreEnableRelease  []HookEntry `yaml:"pre_enable_release"`
	PostEnableRelease []HookEntry `yaml:"post_enable_release"`
}
```

These map onto the `deploy` pipeline in `internal/strategy/atomic/deployer.go`:

```
extract → post_extract → pre_link → [link shared] → pre_enable_release → [switch current] → post_enable_release → purge
```

Two problems:

1. **Naming inconsistency.** `post_extract` and `pre_link` are named after the
   adjacent pipeline action (`extract`, `link`). `pre_enable_release` /
   `post_enable_release` are named after an internal action ("enable
   release") that has no other name in the codebase — the user-facing verb
   for the same operation (switching the `current` symlink) is **activate**,
   as used by `bifrost release activate` and `bifrost release rollback`
   (`internal/cmd/release/activate.go`, `rollback.go`), both of which already
   fire `PreEnableRelease`/`PostEnableRelease` around `SetCurrent`.

2. **Incomplete coverage.** Only `extract` and `link` have a hook on one
   side; `enable_release` has both; `purge` (the final pipeline step) has
   none. There's no hook before the deploy pipeline starts.

## Decision

Replace the four hooks with eight, giving every pipeline stage a symmetric
`pre_<stage>` / `post_<stage>` pair:

| Hook | Fires | Status |
|---|---|---|
| `pre_extract` | Before artifact extraction | new |
| `post_extract` | After extraction | unchanged |
| `pre_link` | Before shared resource linking | unchanged |
| `post_link` | After shared resource linking | new |
| `pre_activate` | Before `current` symlink switch | renamed from `pre_enable_release` |
| `post_activate` | After `current` symlink switch | renamed from `post_enable_release` |
| `pre_purge` | Before old-release purge | new |
| `post_purge` | After old-release purge | new |

`pre_activate` / `post_activate` rename to align with the `activate` verb
already used by `bifrost release activate` / `bifrost release rollback`,
which fire the same hooks around `SetCurrent` and are renamed identically.

## Scope

### `deploy` command (`internal/strategy/atomic/deployer.go`)

Full 8-hook pipeline:

```
pre_extract → [extract]      → post_extract
pre_link    → [link shared]  → post_link
pre_activate → [switch current] → post_activate
pre_purge   → [purge]        → post_purge
```

### `release activate` / `release rollback`

Unchanged scope — only the renamed pair applies, around the existing
`SetCurrent` call:

```
[link shared] → pre_activate → [switch current] → post_activate
```

`pre_link` / `post_link` / `pre_extract` / `post_extract` / `pre_purge` /
`post_purge` do **not** apply to `activate` / `rollback` — these commands
don't extract or purge, and their existing `LinkShared` call has never fired
hooks. This change is a rename plus additive-to-`deploy` only; it does not
expand `activate`/`rollback`'s hook surface.

## Hook semantics

- All 8 hooks behave identically: same `HookEntry` fields (`cmd`, `priority`,
  `sudo`, `cmd_dir`, `allow_fail`, `interactive`), same execution via
  `internal/hooks/runner.go`.
- `pre_purge` / `post_purge` follow normal hook semantics — a failing command
  fails the deploy unless `allow_fail: true`, same as `post_extract` etc.
  today. This is independent of `PurgePlan`'s existing non-fatal-failure
  behavior (an error computing the purge plan still doesn't abort the
  deploy; it just skips purge).
- No changes to `HookData` / `Directories` — all 8 hooks see the same
  template variables (`.Directories.Working`, `.Directories.Current`,
  `.Directories.Releases`, `.Directories.Shared`, `.Variables`, `.Settings`,
  `.Env`). `pre_extract` sees `Directories.Working` as the just-created,
  still-empty release directory. `post_purge` sees `Directories.Working`
  unaffected by the purge of *other* release directories.

## Breaking change

`pre_enable_release` / `post_enable_release` are removed — no alias or
deprecation period. Strict YAML parsing (`forgeconfig.Load`) already rejects
unknown keys, so existing configs using the old names fail loudly with a
clear "unknown field" error rather than silently doing nothing.

## Implementation surface

- `internal/config/schema.go` — `Hooks` struct: rename 2 fields, add 4
- `internal/config/merge.go` — `sortedHooks(...)` calls for all 8 fields
- `internal/strategy/atomic/deployer.go` — 4 new `hooks.RunWithEvents(...)`
  call sites (`pre_extract`, `post_link`, `pre_purge`, `post_purge`); rename
  2 existing call sites' hook-name strings (`"pre_enable_release"` →
  `"pre_activate"`, etc.)
- `internal/cmd/release/activate.go`, `internal/cmd/release/rollback.go` —
  rename `PreEnableRelease`/`PostEnableRelease` field references and
  hook-name strings used in error messages
- `bifrost.schema.json` — update `hooks` definition to the 8-key shape
- `docs/specs/02-configuration.md` — update hook schema, lifecycle table,
  example config
- `docs/specs/05-hooks.md` — rewrite the lifecycle points table for all 8
  hooks, including which have shared symlinks present / `current` updated
- `docs/bifrost.sample.yml` — update annotated hooks section
- `internal/cmd/config/init.go` — update scaffold's `hooks:` block

## Testing

- `internal/config/schema_test.go` — new/renamed `Hooks` fields
- `internal/config/merge_test.go` — merge/sort behavior for new fields
- Deployer-level test — assert all 8 hook points fire in pipeline order for
  `deploy`
- `internal/cmd/release/activate.go` / `rollback.go` tests — renamed hook
  field/error-string assertions
- Integration tests — extend existing `deploy` integration tests to cover
  `pre_extract`, `post_link`, `pre_purge`, `post_purge`

## Documentation

- New ADR: `docs/adr/0012-hook-lifecycle-granularity.md` — records this
  decision and rationale (naming alignment with `activate`, full pipeline
  coverage, breaking-change handling)
- New roadmap milestone — v0 (`roadmap.md`) and v1 (`v1-roadmap.md`) are
  fully complete (M1–M17); this needs a new milestone, placement (new
  `v2-roadmap.md` vs. appended to an existing doc) to be decided during
  planning
