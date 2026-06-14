# ADR-0012: Hook lifecycle granularity — 8 symmetric pre/post stages

**Status:** Accepted

## Context

The `Hooks` struct (`internal/config/schema.go`) originally defined four lifecycle
hooks: `post_extract`, `pre_link`, `pre_enable_release`, `post_enable_release`. These
map onto the `deploy` pipeline in `internal/strategy/atomic/deployer.go`:

```
extract → post_extract → pre_link → [link shared] → pre_enable_release → [switch current] → post_enable_release → purge
```

Two problems:

1. **Naming inconsistency.** `post_extract` and `pre_link` are named after the
   adjacent pipeline action (`extract`, `link`). `pre_enable_release` /
   `post_enable_release` are named after an internal action ("enable release") that
   has no other name in the codebase — the user-facing verb for the same operation
   (switching the `current` symlink) is **activate**, as used by `bifrost release
   activate` and `bifrost release rollback`, both of which already fire
   `PreEnableRelease`/`PostEnableRelease` around `SetCurrent`.

2. **Incomplete coverage.** Only `extract` and `link` have a hook on one side;
   `enable_release` has both; `purge` (the final pipeline step) has none. There's no
   hook before the deploy pipeline starts.

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

`pre_activate` / `post_activate` align with the `activate` verb already used by
`bifrost release activate` / `bifrost release rollback`, which fire the same hooks
around `SetCurrent` and are renamed identically. `release activate` / `release
rollback` are scoped to only `pre_activate`/`post_activate` — they don't extract or
purge, so the other six hooks don't apply.

All 8 hooks behave identically: same `HookEntry` fields (`cmd`, `priority`, `sudo`,
`cmd_dir`, `allow_fail`, `interactive`), same execution via
`internal/hooks/runner.go`. `pre_purge` / `post_purge` follow normal hook semantics —
a failing command fails the deploy unless `allow_fail: true`, independent of
`PurgePlan`'s existing non-fatal-failure behavior (an error computing the purge plan
still doesn't abort the deploy; it just skips purge).

## Breaking change

`pre_enable_release` / `post_enable_release` are removed — no alias or deprecation
period. Strict YAML parsing (`forgeconfig.Load`) already rejects unknown keys, so
existing configs using the old names fail loudly with a clear "unknown field" error
rather than silently doing nothing.

## Alternatives considered

| Option | Why not chosen |
|---|---|
| Keep `pre_enable_release`/`post_enable_release` as aliases | Permanent dual-naming adds maintenance cost and config-reader confusion for no benefit — v0 has no external users yet to migrate |
| Add only `pre_purge`/`post_purge`, leave naming as-is | Doesn't fix the naming inconsistency; a future rename would be a second breaking change |
| Add a generic `hooks: { <any_stage>: [...] }` map instead of fixed fields | Loses compile-time field validation and JSON Schema completeness for marginal flexibility v0 doesn't need |

## Rationale

A fixed, symmetric 8-hook struct keeps the config schema, JSON Schema, and deployer
pipeline in lockstep and easy to validate. Doing the naming rename and the coverage
expansion together is a single breaking change instead of two.

## References

- Design notes: [`.claude/plans/m18-hook-lifecycle-design.md`](../../.claude/plans/m18-hook-lifecycle-design.md)
- [Spec 02 — Configuration](../specs/02-configuration.md)
- [Spec 05 — Hooks](../specs/05-hooks.md)
