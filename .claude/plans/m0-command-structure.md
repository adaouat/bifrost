# Plan: Command Structure + Stub Commands

## Context

Before implementing stub commands we reviewed the planned command hierarchy against v0 requirements and future-version constraints (v1 SSH orchestration, v4/v5 Docker/k8s strategies). Two renames were approved and short flag aliases confirmed. Now is the cheapest time — no implementation code exists yet for these commands.

---

## Confirmed decisions

| # | Decision |
|---|---|
| 1 | Rename `artifact` → `deploy` |
| 2 | Rename `release enable` → `release activate` |
| 3 | No top-level `rollback` alias — only `release rollback` |
| 4 | Add `--env` and `--app` as aliases for `--environment` / `--application` |

---

## Final command structure

```
bifrost config [--environment/--env <env>] [--application/--app <app>]

bifrost deploy
  --environment/--env <env>   (required)
  --application/--app <app>   (required)
  --artifact <path>           (required)
  [--release-name <name>]
  [--init]

bifrost release list
  --environment/--env <env>   (required)
  --application/--app <app>   (required)

bifrost release activate
  --environment/--env <env>   (required)
  --application/--app <app>   (required)
  [--release <name>]

bifrost release rollback
  --environment/--env <env>   (required)
  --application/--app <app>   (required)
```

Global flags (all commands, already implemented in root.go):
`--config`, `--output`, `--dry-run`, `--verbose`

---

## Files to create / update

### New files (stubs only — no logic)

| File | Content |
|---|---|
| `internal/cmd/deploy.go` | `deploy` cobra.Command + flags |
| `internal/cmd/config.go` | `config` cobra.Command + flags |
| `internal/cmd/release/release.go` | `release` group cobra.Command (no RunE — shows help) |
| `internal/cmd/release/list.go` | `release list` cobra.Command + flags |
| `internal/cmd/release/activate.go` | `release activate` cobra.Command + flags |
| `internal/cmd/release/rollback.go` | `release rollback` cobra.Command + flags |

### Updated files

| File | Change |
|---|---|
| `internal/cmd/root.go` | Register `deploy`, `config`, `release` on root |
| `docs/specs/03-commands.md` | Rename `artifact`→`deploy`, `enable`→`activate` |
| `CLAUDE.md` | Update project layout section |
| `docs/tasks/roadmap.md` | Update task descriptions |

---

## Stub pattern

Each leaf command (`deploy`, `config`, `list`, `activate`, `rollback`) gets:
- `Use`, `Short` set correctly
- All flags defined (both long and alias forms binding to the same variable)
- `RunE` returning `fmt.Errorf("not yet implemented")` — fang displays this cleanly

Required-flag enforcement (`MarkFlagRequired`) is NOT added to stubs. It belongs in M1/M2 when the command is actually implemented, to avoid the alias (`--env`) bypassing the required check on `--environment`. Instead, M1+ RunE validates manually.

The `release` parent command has no `RunE` — cobra shows its subcommand help automatically.

---

## Flag alias implementation

cobra/pflag doesn't natively support long-form aliases. The pattern: define both flags binding to the same Go variable:

```go
var env string
f.StringVar(&env, "environment", "", "target environment")
f.StringVar(&env, "env", "", "alias for --environment")
```

Both appear in `--help`. Required validation (whichever alias was used) is handled manually in RunE in M1+. This avoids `MarkFlagRequired` breaking when only the alias is passed.

---

## Spec and docs updates

`docs/specs/03-commands.md`:
- Replace all `artifact` with `deploy` in the command synopsis and description
- Replace all `release enable` / `enable` with `release activate` / `activate`
- Add `--env` / `--app` aliases to the flag tables

`CLAUDE.md` project layout:
- `artifact.go` → `deploy.go`
- `release/enable.go` → `release/activate.go`

`docs/tasks/roadmap.md`:
- Update stub command list to reflect new names

---

## Verification

After implementation:
- `bifrost --help` lists `config`, `deploy`, `release` in the commands section
- `bifrost deploy --help` shows `--environment`/`--env`, `--application`/`--app`, `--artifact`, `--release-name`, `--init`
- `bifrost release --help` lists `list`, `activate`, `rollback`
- `bifrost release activate --help` shows correct flags
- `bifrost deploy --env prod --app web --artifact ./x.tar.gz` runs (returns "not yet implemented")
- `go test ./internal/...` passes
- `hk check` passes
