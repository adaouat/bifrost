# Hook Lifecycle Granularity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the four `Hooks` lifecycle fields (`post_extract`, `pre_link`, `pre_enable_release`, `post_enable_release`) with eight symmetric `pre_<stage>`/`post_<stage>` hooks covering the full `deploy` pipeline, and rename the activation-related pair to `pre_activate`/`post_activate` to match the `release activate`/`release rollback` verb.

**Architecture:** This is a config-schema rename plus additive pipeline hooks. `internal/config/schema.go` gains 4 fields and renames 2; `internal/config/loader.go` and `merge.go` thread all 8 through default-priority application and `Merge()`; `internal/strategy/atomic/deployer.go` fires all 8 around the `deploy` pipeline stages (extract, link, activate, purge); `internal/cmd/deploy.go`'s dry-run preview reflects the same 8; `internal/cmd/release/activate.go` and `rollback.go` rename their existing `pre_activate`/`post_activate` call sites (no new hooks for those commands). This is a breaking change — no alias for the old names, since `forgeconfig.Load` already rejects unknown YAML keys with a clear error.

**Tech Stack:** Go, `internal/config` (schema/loader/merge), `internal/strategy/atomic` (deployer), `internal/cmd` (deploy dry-run, release activate/rollback), `internal/hooks` (runner — no behavioral change, one doc comment), testcontainers-go (integration tests).

---

## Reference: final `Hooks` struct shape

```go
type Hooks struct {
	PreExtract   []HookEntry `yaml:"pre_extract"   json:"pre_extract,omitempty"`
	PostExtract  []HookEntry `yaml:"post_extract"  json:"post_extract,omitempty"`
	PreLink      []HookEntry `yaml:"pre_link"      json:"pre_link,omitempty"`
	PostLink     []HookEntry `yaml:"post_link"     json:"post_link,omitempty"`
	PreActivate  []HookEntry `yaml:"pre_activate"  json:"pre_activate,omitempty"`
	PostActivate []HookEntry `yaml:"post_activate" json:"post_activate,omitempty"`
	PrePurge     []HookEntry `yaml:"pre_purge"     json:"pre_purge,omitempty"`
	PostPurge    []HookEntry `yaml:"post_purge"    json:"post_purge,omitempty"`
}
```

Pipeline order for `deploy`:

```
pre_extract → [extract]      → post_extract
pre_link    → [link shared]  → post_link
pre_activate → [switch current] → post_activate
pre_purge   → [purge]        → post_purge
```

`release activate` / `release rollback` only use `pre_activate` / `post_activate`, around the existing `SetCurrent` call (unchanged scope from today's `pre_enable_release`/`post_enable_release`).

## All files touched (reference list)

| File | Change |
|---|---|
| `internal/config/schema.go` | Rename 2 `Hooks` fields, add 4 |
| `internal/config/schema_test.go` | Update `TestHooks_AllLifecyclePoints` |
| `internal/config/loader.go` | `applyHookDefaults` for all 8 fields |
| `internal/config/loader_test.go` | Rename hook field/yaml refs in `TestLoad_FullFile`, `TestParse_HookPriorityInNestedApp` |
| `testdata/bifrost-full.yml` | `post_enable_release` → `post_activate` |
| `internal/config/merge.go` | `Merge()` `Hooks` block for all 8 fields |
| `internal/config/merge_test.go` | Rename/extend `TestMerge_HooksConcatAndSort`, `TestMerge_AllHookTypesPopulated` |
| `internal/config/flatgen_test.go` | Rename `PostEnableRelease` → `PostActivate` throughout |
| `testdata/bifrost-flat.yml` | `post_enable_release` → `post_activate` |
| `internal/strategy/atomic/deployer.go` | 4 new hook call sites, rename 2, update `deployStepTotal` |
| `internal/hooks/runner.go` | Update doc comment example lifecycle name |
| `internal/cmd/deploy.go` | Dry-run preview: rename 2 hook lines, add 4 new |
| `internal/cmd/release/activate.go` | Rename `PreEnableRelease`/`PostEnableRelease` → `PreActivate`/`PostActivate` |
| `internal/cmd/release/rollback.go` | Same rename |
| `testdata/bifrost-deploy-sudo-test.yml` | `post_enable_release` → `post_activate` |
| `testdata/bifrost-deploy-hooks-test.yml` | New 8-hook fixture |
| `internal/cmd/deploy_integration_test.go` | Extend `TestDeployCmd_Hooks` for 8-hook ordering |
| `bifrost.schema.json` | `hooks` definition: 8 keys |
| `internal/cmd/config/init.go` | Scaffold `hooks:` block: 8 keys |
| `docs/specs/02-configuration.md` | Schema, lifecycle table, example |
| `docs/specs/05-hooks.md` | Lifecycle points table |
| `docs/bifrost.sample.yml` | Annotated hooks section |
| `docs/adr/0012-hook-lifecycle-granularity.md` | New ADR |
| `docs/adr/README.md` | ADR index row |
| `docs/tasks/v1-roadmap.md` | New milestone M18 |

---

### Task 1: Rename and extend the `Hooks` schema

**Files:**
- Modify: `internal/config/schema.go:13-19`
- Test: `internal/config/schema_test.go:39-45`

- [x] **Step 1: Write the failing test**

Replace `TestHooks_AllLifecyclePoints` in `internal/config/schema_test.go`:

```go
func TestHooks_AllLifecyclePoints(t *testing.T) {
	h := config.Hooks{}
	_ = h.PreExtract
	_ = h.PostExtract
	_ = h.PreLink
	_ = h.PostLink
	_ = h.PreActivate
	_ = h.PostActivate
	_ = h.PrePurge
	_ = h.PostPurge
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/... -run TestHooks_AllLifecyclePoints`
Expected: FAIL — `config.Hooks` has no field `PreExtract` (compile error)

- [x] **Step 3: Update the `Hooks` struct**

In `internal/config/schema.go`, replace:

```go
// Hooks groups shell commands by deployment lifecycle point.
type Hooks struct {
	PostExtract       []HookEntry `yaml:"post_extract"        json:"post_extract,omitempty"`
	PreLink           []HookEntry `yaml:"pre_link"            json:"pre_link,omitempty"`
	PreEnableRelease  []HookEntry `yaml:"pre_enable_release"  json:"pre_enable_release,omitempty"`
	PostEnableRelease []HookEntry `yaml:"post_enable_release" json:"post_enable_release,omitempty"`
}
```

with:

```go
// Hooks groups shell commands by deployment lifecycle point.
type Hooks struct {
	PreExtract   []HookEntry `yaml:"pre_extract"   json:"pre_extract,omitempty"`
	PostExtract  []HookEntry `yaml:"post_extract"  json:"post_extract,omitempty"`
	PreLink      []HookEntry `yaml:"pre_link"      json:"pre_link,omitempty"`
	PostLink     []HookEntry `yaml:"post_link"     json:"post_link,omitempty"`
	PreActivate  []HookEntry `yaml:"pre_activate"  json:"pre_activate,omitempty"`
	PostActivate []HookEntry `yaml:"post_activate" json:"post_activate,omitempty"`
	PrePurge     []HookEntry `yaml:"pre_purge"     json:"pre_purge,omitempty"`
	PostPurge    []HookEntry `yaml:"post_purge"    json:"post_purge,omitempty"`
}
```

- [x] **Step 4: Confirm the schema test passes (package won't fully build yet)**

Run: `go vet ./internal/config/... 2>&1 | head -30`
Expected: errors in `loader.go`, `merge.go`, `loader_test.go`, `merge_test.go`,
`flatgen_test.go` referencing the old field names — this is expected. `schema.go` and
`schema_test.go` themselves are correct. Fixed across Tasks 2–4. Do not commit yet.

---

### Task 2: Update `applyHookDefaults` and its tests for all 8 hook fields

**Files:**
- Modify: `internal/config/loader.go:113-118` (`applyHookDefaults`)
- Test: `internal/config/loader_test.go` (`TestLoad_FullFile`, `TestParse_HookPriorityInNestedApp`)
- Modify: `testdata/bifrost-full.yml:31`

- [x] **Step 1: Update `TestLoad_FullFile`'s hook assertions**

In `internal/config/loader_test.go`, within `TestLoad_FullFile`, replace:

```go
	if len(web.Hooks.PostEnableRelease) != 2 {
		t.Errorf("post_enable_release hooks: got %d, want 2", len(web.Hooks.PostEnableRelease))
	}
	if p := web.Hooks.PostEnableRelease[0].Priority; p == nil || *p != 10 {
		t.Errorf("first hook priority: want 10, got %v", p)
	}
```

with:

```go
	if len(web.Hooks.PostActivate) != 2 {
		t.Errorf("post_activate hooks: got %d, want 2", len(web.Hooks.PostActivate))
	}
	if p := web.Hooks.PostActivate[0].Priority; p == nil || *p != 10 {
		t.Errorf("first hook priority: want 10, got %v", p)
	}
```

- [x] **Step 2: Update `TestParse_HookPriorityInNestedApp`**

Replace:

```go
func TestParse_HookPriorityInNestedApp(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(`
paths:
  releases_root: /x
  shared_root: /y
environments:
  prod:
    applications:
      web:
        hooks:
          pre_enable_release:
            - cmd: "echo migrate"
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := cfg.Environments["prod"].Applications["web"].Hooks.PreEnableRelease[0]
	if h.Priority == nil || *h.Priority != 99999 {
		t.Errorf("nested hook priority default: want 99999, got %v", h.Priority)
	}
}
```

with:

```go
func TestParse_HookPriorityInNestedApp(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(`
paths:
  releases_root: /x
  shared_root: /y
environments:
  prod:
    applications:
      web:
        hooks:
          pre_activate:
            - cmd: "echo migrate"
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := cfg.Environments["prod"].Applications["web"].Hooks.PreActivate[0]
	if h.Priority == nil || *h.Priority != 99999 {
		t.Errorf("nested hook priority default: want 99999, got %v", h.Priority)
	}
}
```

- [x] **Step 3: Update `testdata/bifrost-full.yml`**

In `testdata/bifrost-full.yml`, replace:

```yaml
        hooks:
          post_extract:
            - cmd: "composer install --no-dev --optimize-autoloader"
              cmd_dir: "{{ .Directories.Working }}"
          post_enable_release:
            - cmd: "systemctl reload php-fpm"
              sudo: true
              priority: 10
            - cmd: "systemctl reload nginx"
              sudo: true
              priority: 20
```

with:

```yaml
        hooks:
          post_extract:
            - cmd: "composer install --no-dev --optimize-autoloader"
              cmd_dir: "{{ .Directories.Working }}"
          post_activate:
            - cmd: "systemctl reload php-fpm"
              sudo: true
              priority: 10
            - cmd: "systemctl reload nginx"
              sudo: true
              priority: 20
```

- [x] **Step 4: Run tests to verify they fail**

Run: `go test ./internal/config/... -run 'TestLoad_FullFile|TestParse_HookPriorityInNestedApp'`
Expected: FAIL — `config.Hooks` has no field `PostActivate`/`PreActivate` yet (compile
error in `loader.go`'s `applyHookDefaults`, which still references
`PreEnableRelease`/`PostEnableRelease`)

- [x] **Step 5: Update `applyHookDefaults`**

In `internal/config/loader.go`, replace:

```go
func applyHookDefaults(h *Hooks) {
	setDefaultPriority(h.PostExtract)
	setDefaultPriority(h.PreLink)
	setDefaultPriority(h.PreEnableRelease)
	setDefaultPriority(h.PostEnableRelease)
}
```

with:

```go
func applyHookDefaults(h *Hooks) {
	setDefaultPriority(h.PreExtract)
	setDefaultPriority(h.PostExtract)
	setDefaultPriority(h.PreLink)
	setDefaultPriority(h.PostLink)
	setDefaultPriority(h.PreActivate)
	setDefaultPriority(h.PostActivate)
	setDefaultPriority(h.PrePurge)
	setDefaultPriority(h.PostPurge)
}
```

- [x] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/config/... -run 'TestLoad_FullFile|TestParse_HookPriorityInNestedApp|TestHooks_AllLifecyclePoints'`
Expected: PASS (note: `merge.go`, `merge_test.go`, `flatgen_test.go` still reference old
field names — `go test ./internal/config/...` as a whole still fails to build. Fixed in
Tasks 3–4. Do not commit yet.)

---

### Task 3: Update `Merge()` and its tests for all 8 hook fields

**Files:**
- Modify: `internal/config/merge.go:61-66`
- Test: `internal/config/merge_test.go:113-147` (`TestMerge_HooksConcatAndSort`), `internal/config/merge_test.go:262-286` (`TestMerge_AllHookTypesPopulated`)

- [x] **Step 1: Update `TestMerge_HooksConcatAndSort` to use `PostActivate`**

In `internal/config/merge_test.go`, replace the three `PostEnableRelease` references in
`TestMerge_HooksConcatAndSort` with `PostActivate`:

```go
func TestMerge_HooksConcatAndSort(t *testing.T) {
	p := func(n int) *int { return &n }
	cfg := base()
	cfg.Hooks.PostActivate = []config.HookEntry{
		{Cmd: "global-reload", Priority: p(20)},
	}
	env := cfg.Environments["prod"]
	env.Hooks.PostActivate = []config.HookEntry{
		{Cmd: "env-notify", Priority: p(30)},
	}
	env.Applications = map[string]config.Application{
		"web": {Hooks: config.Hooks{
			PostActivate: []config.HookEntry{
				{Cmd: "app-cache", Priority: p(10)},
			},
		}},
	}
	cfg.Environments["prod"] = env

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hooks := m.Hooks.PostActivate
	if len(hooks) != 3 {
		t.Fatalf("hook count: got %d, want 3", len(hooks))
	}
	// sorted by priority: app(10) < global(20) < env(30)
	want := []string{"app-cache", "global-reload", "env-notify"}
	for i, h := range hooks {
		if h.Cmd != want[i] {
			t.Errorf("hooks[%d].cmd: got %q, want %q", i, h.Cmd, want[i])
		}
	}
}
```

- [x] **Step 2: Replace `TestMerge_AllHookTypesPopulated` with all 8 fields**

```go
func TestMerge_AllHookTypesPopulated(t *testing.T) {
	p := func(n int) *int { return &n }
	cfg := base()
	cfg.Hooks.PreExtract = []config.HookEntry{{Cmd: "pre-extract", Priority: p(10)}}
	cfg.Hooks.PostExtract = []config.HookEntry{{Cmd: "post-extract", Priority: p(10)}}
	cfg.Hooks.PreLink = []config.HookEntry{{Cmd: "pre-link", Priority: p(10)}}
	cfg.Hooks.PostLink = []config.HookEntry{{Cmd: "post-link", Priority: p(10)}}
	cfg.Hooks.PreActivate = []config.HookEntry{{Cmd: "pre-activate", Priority: p(10)}}
	cfg.Hooks.PostActivate = []config.HookEntry{{Cmd: "post-activate", Priority: p(10)}}
	cfg.Hooks.PrePurge = []config.HookEntry{{Cmd: "pre-purge", Priority: p(10)}}
	cfg.Hooks.PostPurge = []config.HookEntry{{Cmd: "post-purge", Priority: p(10)}}

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Hooks.PreExtract) != 1 || m.Hooks.PreExtract[0].Cmd != "pre-extract" {
		t.Error("PreExtract not populated correctly")
	}
	if len(m.Hooks.PostExtract) != 1 || m.Hooks.PostExtract[0].Cmd != "post-extract" {
		t.Error("PostExtract not populated correctly")
	}
	if len(m.Hooks.PreLink) != 1 || m.Hooks.PreLink[0].Cmd != "pre-link" {
		t.Error("PreLink not populated correctly")
	}
	if len(m.Hooks.PostLink) != 1 || m.Hooks.PostLink[0].Cmd != "post-link" {
		t.Error("PostLink not populated correctly")
	}
	if len(m.Hooks.PreActivate) != 1 || m.Hooks.PreActivate[0].Cmd != "pre-activate" {
		t.Error("PreActivate not populated correctly")
	}
	if len(m.Hooks.PostActivate) != 1 || m.Hooks.PostActivate[0].Cmd != "post-activate" {
		t.Error("PostActivate not populated correctly")
	}
	if len(m.Hooks.PrePurge) != 1 || m.Hooks.PrePurge[0].Cmd != "pre-purge" {
		t.Error("PrePurge not populated correctly")
	}
	if len(m.Hooks.PostPurge) != 1 || m.Hooks.PostPurge[0].Cmd != "post-purge" {
		t.Error("PostPurge not populated correctly")
	}
}
```

- [x] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/config/... -run 'TestMerge_HooksConcatAndSort|TestMerge_AllHookTypesPopulated'`
Expected: FAIL (compile error — `Merge()` doesn't populate `PreExtract`, `PostLink`, etc.,
and `m.Hooks.PostActivate` doesn't exist as a populated field yet)

- [x] **Step 4: Update `Merge()`'s `Hooks` block**

In `internal/config/merge.go`, replace:

```go
		Hooks: Hooks{
			PostExtract:       sortedHooks(cfg.Hooks.PostExtract, env.Hooks.PostExtract, app.Hooks.PostExtract),
			PreLink:           sortedHooks(cfg.Hooks.PreLink, env.Hooks.PreLink, app.Hooks.PreLink),
			PreEnableRelease:  sortedHooks(cfg.Hooks.PreEnableRelease, env.Hooks.PreEnableRelease, app.Hooks.PreEnableRelease),
			PostEnableRelease: sortedHooks(cfg.Hooks.PostEnableRelease, env.Hooks.PostEnableRelease, app.Hooks.PostEnableRelease),
		},
```

with:

```go
		Hooks: Hooks{
			PreExtract:   sortedHooks(cfg.Hooks.PreExtract, env.Hooks.PreExtract, app.Hooks.PreExtract),
			PostExtract:  sortedHooks(cfg.Hooks.PostExtract, env.Hooks.PostExtract, app.Hooks.PostExtract),
			PreLink:      sortedHooks(cfg.Hooks.PreLink, env.Hooks.PreLink, app.Hooks.PreLink),
			PostLink:     sortedHooks(cfg.Hooks.PostLink, env.Hooks.PostLink, app.Hooks.PostLink),
			PreActivate:  sortedHooks(cfg.Hooks.PreActivate, env.Hooks.PreActivate, app.Hooks.PreActivate),
			PostActivate: sortedHooks(cfg.Hooks.PostActivate, env.Hooks.PostActivate, app.Hooks.PostActivate),
			PrePurge:     sortedHooks(cfg.Hooks.PrePurge, env.Hooks.PrePurge, app.Hooks.PrePurge),
			PostPurge:    sortedHooks(cfg.Hooks.PostPurge, env.Hooks.PostPurge, app.Hooks.PostPurge),
		},
```

- [x] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/config/... -run 'TestMerge_HooksConcatAndSort|TestMerge_AllHookTypesPopulated'`
Expected: PASS (note: `flatgen_test.go` still references `PostEnableRelease` — full
`go test ./internal/config/...` still fails to build. Fixed in Task 4. Do not commit yet.)

---

### Task 4: Update `flatgen_test.go` and its fixture for `PostActivate`

**Files:**
- Test: `internal/config/flatgen_test.go` (lines 38, 123-129, 177, 190, 202, 283-290)
- Modify: `testdata/bifrost-flat.yml:8-10`

- [x] **Step 1: Rename `PostEnableRelease` → `PostActivate` in `flatgenBase`**

In `internal/config/flatgen_test.go`, in `flatgenBase()`, replace:

```go
						Hooks: config.Hooks{
							PostEnableRelease: []config.HookEntry{
								{Cmd: "systemctl reload nginx", Sudo: true, Priority: &p},
							},
						},
```

with:

```go
						Hooks: config.Hooks{
							PostActivate: []config.HookEntry{
								{Cmd: "systemctl reload nginx", Sudo: true, Priority: &p},
							},
						},
```

- [x] **Step 2: Rename in `TestGenerateFlatConfig_HooksMerged`**

Replace:

```go
	if len(parsed.Hooks.PostEnableRelease) != 1 {
		t.Fatalf("post_enable_release: got %d hooks, want 1", len(parsed.Hooks.PostEnableRelease))
	}
	if parsed.Hooks.PostEnableRelease[0].Cmd != "systemctl reload nginx" {
		t.Errorf("hook cmd: got %q, want systemctl reload nginx", parsed.Hooks.PostEnableRelease[0].Cmd)
	}
	if !parsed.Hooks.PostEnableRelease[0].Sudo {
		t.Error("hook sudo: got false, want true")
	}
```

with:

```go
	if len(parsed.Hooks.PostActivate) != 1 {
		t.Fatalf("post_activate: got %d hooks, want 1", len(parsed.Hooks.PostActivate))
	}
	if parsed.Hooks.PostActivate[0].Cmd != "systemctl reload nginx" {
		t.Errorf("hook cmd: got %q, want systemctl reload nginx", parsed.Hooks.PostActivate[0].Cmd)
	}
	if !parsed.Hooks.PostActivate[0].Sudo {
		t.Error("hook sudo: got false, want true")
	}
```

- [x] **Step 3: Rename the three `PostEnableRelease` fields in `flatgenLayered`**

In `flatgenLayered()`, replace all three occurrences of:

```go
				Hooks: config.Hooks{
					PostEnableRelease: []config.HookEntry{
```

with:

```go
				Hooks: config.Hooks{
					PostActivate: []config.HookEntry{
```

(There are three blocks — global-level `Hooks` on the `Config`, env-level `Hooks` on
`prod`, and app-level `Hooks` on `web`. Apply the same field rename to each; the
`Cmd`/`Priority` values inside each block are unchanged: `"global-reload"`/`p(20)`,
`"env-notify"`/`p(30)`, `"app-cache"`/`p(10)` respectively.)

- [x] **Step 4: Rename in `TestGenerateFlatConfig_HooksSortedAcrossLevels`**

Replace:

```go
func TestGenerateFlatConfig_HooksSortedAcrossLevels(t *testing.T) {
	flat := flatten(t, flatgenLayered())

	hooks := flat.Hooks.PostEnableRelease
	if len(hooks) != 3 {
		t.Fatalf("post_enable_release: got %d hooks, want 3", len(hooks))
	}
```

with:

```go
func TestGenerateFlatConfig_HooksSortedAcrossLevels(t *testing.T) {
	flat := flatten(t, flatgenLayered())

	hooks := flat.Hooks.PostActivate
	if len(hooks) != 3 {
		t.Fatalf("post_activate: got %d hooks, want 3", len(hooks))
	}
```

- [x] **Step 5: Update `testdata/bifrost-flat.yml`**

Replace:

```yaml
hooks:
  post_enable_release:
    - cmd: "systemctl reload nginx"
      sudo: true
```

with:

```yaml
hooks:
  post_activate:
    - cmd: "systemctl reload nginx"
      sudo: true
```

- [x] **Step 6: Run all `internal/config` tests**

Run: `go test ./internal/config/...`
Expected: PASS — full `internal/config` package builds and all tests pass.

- [x] **Step 7: Run `go build ./...` to check downstream packages**

Run: `go build ./... 2>&1 | head -30`
Expected: errors in `internal/strategy/atomic/deployer.go`, `internal/cmd/deploy.go`,
`internal/cmd/release/activate.go`, `internal/cmd/release/rollback.go` referencing
removed fields — expected, fixed in Tasks 5–7. Do not commit yet.

---

### Task 5: Update the `deploy` pipeline in `internal/strategy/atomic/deployer.go`

**Files:**
- Modify: `internal/strategy/atomic/deployer.go:21-34` (`deployStepTotal`)
- Modify: `internal/strategy/atomic/deployer.go:119-238` (hook call sites + hookData placement)
- Modify: `internal/strategy/atomic/deployer.go:240-262` (purge hooks)
- Modify: `internal/hooks/runner.go:49` (doc comment)

This task has no new Go unit test (hook firing order is verified by the integration test
in Task 8, and unit tests for this package only check type wiring — see
`internal/strategy/atomic/deployer_test.go`). Apply the changes below, then verify with
`go build ./...` and `go test ./...`.

- [x] **Step 1: Update `deployStepTotal`**

Replace:

```go
// deployStepTotal counts the numbered step lines a human-mode deploy renders:
// seven always-present steps plus one per configured hook group.
func deployStepTotal(cfg *config.MergedConfig) int {
	total := 7 // config, release dir, extract, shared dirs, shared files, symlink, purge
	for _, h := range [][]config.HookEntry{
		cfg.Hooks.PostExtract, cfg.Hooks.PreLink,
		cfg.Hooks.PreEnableRelease, cfg.Hooks.PostEnableRelease,
	} {
		if len(h) > 0 {
			total++
		}
	}
	return total
}
```

with:

```go
// deployStepTotal counts the numbered step lines a human-mode deploy renders:
// seven always-present steps plus one per configured hook group.
func deployStepTotal(cfg *config.MergedConfig) int {
	total := 7 // config, release dir, extract, shared dirs, shared files, symlink, purge
	for _, h := range [][]config.HookEntry{
		cfg.Hooks.PreExtract, cfg.Hooks.PostExtract,
		cfg.Hooks.PreLink, cfg.Hooks.PostLink,
		cfg.Hooks.PreActivate, cfg.Hooks.PostActivate,
		cfg.Hooks.PrePurge, cfg.Hooks.PostPurge,
	} {
		if len(h) > 0 {
			total++
		}
	}
	return total
}
```

- [x] **Step 2: Move `hookData`/`hookEventFn` construction earlier and add `pre_extract`**

Replace the block from the human-mode header print through the `pre_link` hook (originally
lines 119–194):

```go
	if !d.jsonMode() {
		_, _ = fmt.Fprint(d.out, tui.DeployHeader(d.mode, opts.Env, opts.App, filepath.Base(releaseDir)))
		sp.Step("Config loaded and validated", "")
		sp.Step("Release directory created", "")
		tui.PrintDetail(d.mode, d.out, "Root path: "+merged.ReleasesRoot)
		tui.PrintDetail(d.mode, d.out, "Shared path: "+merged.SharedRoot)
	}

	deployStart := time.Now()

	// Extract artifact.
	info, err := os.Stat(opts.Artifact)
	if err != nil {
		return fmt.Errorf("stat artifact: %w", err)
	}

	if d.jsonMode() {
		emit.Emit(map[string]any{"event": "start", "step": "extract", "artifact": opts.Artifact})
	}
	extractStart := time.Now()
	updateProgress, doneProgress := tui.NewProgressBar(d.mode, info.Size(), "Extracting artifact", d.out)
	var jsonProgressFn func(n int64)
	if d.jsonMode() {
		var written int64
		total := info.Size()
		jsonProgressFn = func(n int64) {
			written += n
			emit.Emit(map[string]any{"event": "progress", "step": "extract", "bytes": written, "total": total})
		}
	}
	progressFn := func(n int64) {
		updateProgress(n)
		if jsonProgressFn != nil {
			jsonProgressFn(n)
		}
	}
	if err := Extract(ctx, opts.Artifact, releaseDir, progressFn); err != nil {
		return emitError("extract", fmt.Errorf("extracting artifact: %w", err))
	}
	doneProgress()
	d.debug("artifact extracted", "artifact", opts.Artifact, "bytes", info.Size())
	if d.jsonMode() {
		emit.Emit(map[string]any{"event": "done", "step": "extract", "duration_ms": time.Since(extractStart).Milliseconds()})
	}
	if !d.jsonMode() {
		sp.Step("Artifact extracted", fmt.Sprintf("(%.1fs)", time.Since(extractStart).Seconds()))
	}

	hookData := hooks.HookData{
		Settings:  merged.Settings,
		Variables: merged.Variables,
		Directories: hooks.Directories{
			Working:  releaseDir,
			Current:  filepath.Join(merged.ReleasesRoot, "current"),
			Releases: merged.ReleasesRoot,
			Shared:   merged.SharedRoot,
		},
		Env: envMap(),
	}
	hookEventFn := hookEmitter(emit)

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PostExtract, hookData, releaseDir, d.out, d.confirmFn, "post_extract", hookEventFn); err != nil {
		return emitError("post_extract", fmt.Errorf("post_extract hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PostExtract) > 0 {
		n := len(merged.Hooks.PostExtract)
		sp.Step("post_extract hooks", fmt.Sprintf("(%d/%d)", n, n))
	}

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PreLink, hookData, releaseDir, d.out, d.confirmFn, "pre_link", hookEventFn); err != nil {
		return emitError("pre_link", fmt.Errorf("pre_link hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PreLink) > 0 {
		n := len(merged.Hooks.PreLink)
		sp.Step("pre_link hooks", fmt.Sprintf("(%d/%d)", n, n))
	}
```

with:

```go
	if !d.jsonMode() {
		_, _ = fmt.Fprint(d.out, tui.DeployHeader(d.mode, opts.Env, opts.App, filepath.Base(releaseDir)))
		sp.Step("Config loaded and validated", "")
		sp.Step("Release directory created", "")
		tui.PrintDetail(d.mode, d.out, "Root path: "+merged.ReleasesRoot)
		tui.PrintDetail(d.mode, d.out, "Shared path: "+merged.SharedRoot)
	}

	deployStart := time.Now()

	hookData := hooks.HookData{
		Settings:  merged.Settings,
		Variables: merged.Variables,
		Directories: hooks.Directories{
			Working:  releaseDir,
			Current:  filepath.Join(merged.ReleasesRoot, "current"),
			Releases: merged.ReleasesRoot,
			Shared:   merged.SharedRoot,
		},
		Env: envMap(),
	}
	hookEventFn := hookEmitter(emit)

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PreExtract, hookData, releaseDir, d.out, d.confirmFn, "pre_extract", hookEventFn); err != nil {
		return emitError("pre_extract", fmt.Errorf("pre_extract hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PreExtract) > 0 {
		n := len(merged.Hooks.PreExtract)
		sp.Step("pre_extract hooks", fmt.Sprintf("(%d/%d)", n, n))
	}

	// Extract artifact.
	info, err := os.Stat(opts.Artifact)
	if err != nil {
		return fmt.Errorf("stat artifact: %w", err)
	}

	if d.jsonMode() {
		emit.Emit(map[string]any{"event": "start", "step": "extract", "artifact": opts.Artifact})
	}
	extractStart := time.Now()
	updateProgress, doneProgress := tui.NewProgressBar(d.mode, info.Size(), "Extracting artifact", d.out)
	var jsonProgressFn func(n int64)
	if d.jsonMode() {
		var written int64
		total := info.Size()
		jsonProgressFn = func(n int64) {
			written += n
			emit.Emit(map[string]any{"event": "progress", "step": "extract", "bytes": written, "total": total})
		}
	}
	progressFn := func(n int64) {
		updateProgress(n)
		if jsonProgressFn != nil {
			jsonProgressFn(n)
		}
	}
	if err := Extract(ctx, opts.Artifact, releaseDir, progressFn); err != nil {
		return emitError("extract", fmt.Errorf("extracting artifact: %w", err))
	}
	doneProgress()
	d.debug("artifact extracted", "artifact", opts.Artifact, "bytes", info.Size())
	if d.jsonMode() {
		emit.Emit(map[string]any{"event": "done", "step": "extract", "duration_ms": time.Since(extractStart).Milliseconds()})
	}
	if !d.jsonMode() {
		sp.Step("Artifact extracted", fmt.Sprintf("(%.1fs)", time.Since(extractStart).Seconds()))
	}

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PostExtract, hookData, releaseDir, d.out, d.confirmFn, "post_extract", hookEventFn); err != nil {
		return emitError("post_extract", fmt.Errorf("post_extract hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PostExtract) > 0 {
		n := len(merged.Hooks.PostExtract)
		sp.Step("post_extract hooks", fmt.Sprintf("(%d/%d)", n, n))
	}

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PreLink, hookData, releaseDir, d.out, d.confirmFn, "pre_link", hookEventFn); err != nil {
		return emitError("pre_link", fmt.Errorf("pre_link hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PreLink) > 0 {
		n := len(merged.Hooks.PreLink)
		sp.Step("pre_link hooks", fmt.Sprintf("(%d/%d)", n, n))
	}
```

- [x] **Step 3: Add `post_link` and rename `pre_enable_release` → `pre_activate`**

Replace (originally lines 196–219):

```go
	if err := runStep("link", func() (map[string]any, error) {
		err := LinkShared(merged.SharedDirs, merged.SharedFiles, releaseDir, merged.SharedRoot)
		return map[string]any{"dirs": merged.SharedDirs, "files": merged.SharedFiles}, err
	}); err != nil {
		return fmt.Errorf("linking shared resources: %w", err)
	}
	if !d.jsonMode() {
		sp.Step("Shared directories linked", fmt.Sprintf("(%d)", len(merged.SharedDirs)))
		for _, dir := range merged.SharedDirs {
			tui.PrintDetail(d.mode, d.out, dir)
		}
		sp.Step("Shared files linked", fmt.Sprintf("(%d)", len(merged.SharedFiles)))
		for _, f := range merged.SharedFiles {
			tui.PrintDetail(d.mode, d.out, f)
		}
	}

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PreEnableRelease, hookData, releaseDir, d.out, d.confirmFn, "pre_enable_release", hookEventFn); err != nil {
		return emitError("pre_enable_release", fmt.Errorf("pre_enable_release hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PreEnableRelease) > 0 {
		n := len(merged.Hooks.PreEnableRelease)
		sp.Step("pre_enable_release hooks", fmt.Sprintf("(%d/%d)", n, n))
	}
```

with:

```go
	if err := runStep("link", func() (map[string]any, error) {
		err := LinkShared(merged.SharedDirs, merged.SharedFiles, releaseDir, merged.SharedRoot)
		return map[string]any{"dirs": merged.SharedDirs, "files": merged.SharedFiles}, err
	}); err != nil {
		return fmt.Errorf("linking shared resources: %w", err)
	}
	if !d.jsonMode() {
		sp.Step("Shared directories linked", fmt.Sprintf("(%d)", len(merged.SharedDirs)))
		for _, dir := range merged.SharedDirs {
			tui.PrintDetail(d.mode, d.out, dir)
		}
		sp.Step("Shared files linked", fmt.Sprintf("(%d)", len(merged.SharedFiles)))
		for _, f := range merged.SharedFiles {
			tui.PrintDetail(d.mode, d.out, f)
		}
	}

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PostLink, hookData, releaseDir, d.out, d.confirmFn, "post_link", hookEventFn); err != nil {
		return emitError("post_link", fmt.Errorf("post_link hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PostLink) > 0 {
		n := len(merged.Hooks.PostLink)
		sp.Step("post_link hooks", fmt.Sprintf("(%d/%d)", n, n))
	}

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PreActivate, hookData, releaseDir, d.out, d.confirmFn, "pre_activate", hookEventFn); err != nil {
		return emitError("pre_activate", fmt.Errorf("pre_activate hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PreActivate) > 0 {
		n := len(merged.Hooks.PreActivate)
		sp.Step("pre_activate hooks", fmt.Sprintf("(%d/%d)", n, n))
	}
```

- [x] **Step 4: Rename `post_enable_release` → `post_activate`**

Replace (originally lines 232–238):

```go
	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PostEnableRelease, hookData, releaseDir, d.out, d.confirmFn, "post_enable_release", hookEventFn); err != nil {
		return emitError("post_enable_release", fmt.Errorf("post_enable_release hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PostEnableRelease) > 0 {
		n := len(merged.Hooks.PostEnableRelease)
		sp.Step("post_enable_release hooks", fmt.Sprintf("(%d/%d)", n, n))
	}
```

with:

```go
	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PostActivate, hookData, releaseDir, d.out, d.confirmFn, "post_activate", hookEventFn); err != nil {
		return emitError("post_activate", fmt.Errorf("post_activate hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PostActivate) > 0 {
		n := len(merged.Hooks.PostActivate)
		sp.Step("post_activate hooks", fmt.Sprintf("(%d/%d)", n, n))
	}
```

- [x] **Step 5: Add `pre_purge` and `post_purge` around the purge step**

Replace (originally lines 240–262):

```go
	// PurgePlan failure is non-fatal: Purge carries its own error path.
	purgePlan, purgePlanErr := PurgePlan(merged.ReleasesRoot, filepath.Base(releaseDir), merged.Settings.ReleasesToKeep)
	if purgePlanErr != nil {
		purgePlan = nil
	}
	if err := runStep("purge", func() (map[string]any, error) {
		extras := map[string]any{"purged": purgePlan, "kept": merged.Settings.ReleasesToKeep}
		purge := func() error { return Purge(merged.ReleasesRoot, merged.Settings.ReleasesToKeep) }
		if d.jsonMode() {
			return extras, purge()
		}
		// Spin while purging (human + TTY) then resolve to "✓ Old releases purged — (kept N)".
		return extras, sp.Run("Old releases purged", func() (forgeui.Result, error) {
			return forgeui.Result{Detail: fmt.Sprintf("(kept %d)", merged.Settings.ReleasesToKeep)}, purge()
		})
	}); err != nil {
		return fmt.Errorf("purging old releases: %w", err)
	}
	if !d.jsonMode() {
		for _, r := range purgePlan {
			tui.PrintDetail(d.mode, d.out, r+" deleted")
		}
	}
```

with:

```go
	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PrePurge, hookData, releaseDir, d.out, d.confirmFn, "pre_purge", hookEventFn); err != nil {
		return emitError("pre_purge", fmt.Errorf("pre_purge hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PrePurge) > 0 {
		n := len(merged.Hooks.PrePurge)
		sp.Step("pre_purge hooks", fmt.Sprintf("(%d/%d)", n, n))
	}

	// PurgePlan failure is non-fatal: Purge carries its own error path.
	purgePlan, purgePlanErr := PurgePlan(merged.ReleasesRoot, filepath.Base(releaseDir), merged.Settings.ReleasesToKeep)
	if purgePlanErr != nil {
		purgePlan = nil
	}
	if err := runStep("purge", func() (map[string]any, error) {
		extras := map[string]any{"purged": purgePlan, "kept": merged.Settings.ReleasesToKeep}
		purge := func() error { return Purge(merged.ReleasesRoot, merged.Settings.ReleasesToKeep) }
		if d.jsonMode() {
			return extras, purge()
		}
		// Spin while purging (human + TTY) then resolve to "✓ Old releases purged — (kept N)".
		return extras, sp.Run("Old releases purged", func() (forgeui.Result, error) {
			return forgeui.Result{Detail: fmt.Sprintf("(kept %d)", merged.Settings.ReleasesToKeep)}, purge()
		})
	}); err != nil {
		return fmt.Errorf("purging old releases: %w", err)
	}
	if !d.jsonMode() {
		for _, r := range purgePlan {
			tui.PrintDetail(d.mode, d.out, r+" deleted")
		}
	}

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PostPurge, hookData, releaseDir, d.out, d.confirmFn, "post_purge", hookEventFn); err != nil {
		return emitError("post_purge", fmt.Errorf("post_purge hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PostPurge) > 0 {
		n := len(merged.Hooks.PostPurge)
		sp.Step("post_purge hooks", fmt.Sprintf("(%d/%d)", n, n))
	}
```

- [x] **Step 6: Update the doc comment in `internal/hooks/runner.go`**

In `internal/hooks/runner.go`, replace:

```go
// lifecycle is the name of the hook group (e.g. "pre_enable_release"). eventFn may be nil.
```

with:

```go
// lifecycle is the name of the hook group (e.g. "pre_activate"). eventFn may be nil.
```

- [x] **Step 7: Build the two affected packages**

Run: `go build ./internal/strategy/... ./internal/hooks/...`
Expected: PASS. `go build ./...` still fails on `internal/cmd` — fixed in Tasks 6–7. Do
not commit yet.

---

### Task 6: Update the `deploy --dry-run` preview in `internal/cmd/deploy.go`

**Files:**
- Modify: `internal/cmd/deploy.go:110-153` (`deployDryRun`)

- [x] **Step 1: Replace `deployDryRun`'s body**

Replace:

```go
	_, _ = fmt.Fprintln(out, "DRY RUN — no changes will be made")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "  Would create   %s\n", releaseDir)
	_, _ = fmt.Fprintf(out, "  Would extract  %s  →  %s\n", artifact, releaseDir)
	for _, h := range merged.Hooks.PostExtract {
		hookLine("post_extract", h)
	}
	for _, h := range merged.Hooks.PreLink {
		hookLine("pre_link", h)
	}
	for _, rel := range merged.SharedDirs {
		_, _ = fmt.Fprintf(out, "  Would link     %s  →  %s\n",
			filepath.Join(releaseDir, rel), filepath.Join(merged.SharedRoot, rel))
	}
	for _, rel := range merged.SharedFiles {
		_, _ = fmt.Fprintf(out, "  Would link     %s  →  %s\n",
			filepath.Join(releaseDir, rel), filepath.Join(merged.SharedRoot, rel))
	}
	for _, h := range merged.Hooks.PreEnableRelease {
		hookLine("pre_enable_release", h)
	}
	_, _ = fmt.Fprintf(out, "  Would update   %s  →  %s\n", currentLink, relBase)
	for _, h := range merged.Hooks.PostEnableRelease {
		hookLine("post_enable_release", h)
	}
	if candidates, err := atomic.PurgePlan(merged.ReleasesRoot, relBase, merged.Settings.ReleasesToKeep); err == nil && len(candidates) > 0 {
		_, _ = fmt.Fprintf(out, "  Would purge    %s  (keeping %d)\n", strings.Join(candidates, ", "), merged.Settings.ReleasesToKeep)
	}
	return nil
}
```

with:

```go
	_, _ = fmt.Fprintln(out, "DRY RUN — no changes will be made")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "  Would create   %s\n", releaseDir)
	for _, h := range merged.Hooks.PreExtract {
		hookLine("pre_extract", h)
	}
	_, _ = fmt.Fprintf(out, "  Would extract  %s  →  %s\n", artifact, releaseDir)
	for _, h := range merged.Hooks.PostExtract {
		hookLine("post_extract", h)
	}
	for _, h := range merged.Hooks.PreLink {
		hookLine("pre_link", h)
	}
	for _, rel := range merged.SharedDirs {
		_, _ = fmt.Fprintf(out, "  Would link     %s  →  %s\n",
			filepath.Join(releaseDir, rel), filepath.Join(merged.SharedRoot, rel))
	}
	for _, rel := range merged.SharedFiles {
		_, _ = fmt.Fprintf(out, "  Would link     %s  →  %s\n",
			filepath.Join(releaseDir, rel), filepath.Join(merged.SharedRoot, rel))
	}
	for _, h := range merged.Hooks.PostLink {
		hookLine("post_link", h)
	}
	for _, h := range merged.Hooks.PreActivate {
		hookLine("pre_activate", h)
	}
	_, _ = fmt.Fprintf(out, "  Would update   %s  →  %s\n", currentLink, relBase)
	for _, h := range merged.Hooks.PostActivate {
		hookLine("post_activate", h)
	}
	for _, h := range merged.Hooks.PrePurge {
		hookLine("pre_purge", h)
	}
	if candidates, err := atomic.PurgePlan(merged.ReleasesRoot, relBase, merged.Settings.ReleasesToKeep); err == nil && len(candidates) > 0 {
		_, _ = fmt.Fprintf(out, "  Would purge    %s  (keeping %d)\n", strings.Join(candidates, ", "), merged.Settings.ReleasesToKeep)
	}
	for _, h := range merged.Hooks.PostPurge {
		hookLine("post_purge", h)
	}
	return nil
}
```

- [x] **Step 2: Build `internal/cmd`**

Run: `go build ./internal/cmd/...`
Expected: errors remain only in `internal/cmd/release/activate.go` and `rollback.go` —
fixed in Task 7. Do not commit yet.

---

### Task 7: Rename hook references in `release activate` and `release rollback`

**Files:**
- Modify: `internal/cmd/release/activate.go:117-127`
- Modify: `internal/cmd/release/rollback.go:84-94`

- [x] **Step 1: Update `activate.go`**

Replace:

```go
			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PreEnableRelease, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("pre_enable_release hooks: %w", err)
			}

			if err := atomic.SetCurrent(merged.ReleasesRoot, releaseDir); err != nil {
				return fmt.Errorf("updating current symlink: %w", err)
			}

			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PostEnableRelease, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("post_enable_release hooks: %w", err)
			}
```

with:

```go
			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PreActivate, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("pre_activate hooks: %w", err)
			}

			if err := atomic.SetCurrent(merged.ReleasesRoot, releaseDir); err != nil {
				return fmt.Errorf("updating current symlink: %w", err)
			}

			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PostActivate, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("post_activate hooks: %w", err)
			}
```

- [x] **Step 2: Update `rollback.go`**

Replace:

```go
			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PreEnableRelease, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("pre_enable_release hooks: %w", err)
			}

			if err := atomic.SetCurrent(merged.ReleasesRoot, releaseDir); err != nil {
				return fmt.Errorf("updating current symlink: %w", err)
			}

			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PostEnableRelease, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("post_enable_release hooks: %w", err)
			}
```

with:

```go
			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PreActivate, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("pre_activate hooks: %w", err)
			}

			if err := atomic.SetCurrent(merged.ReleasesRoot, releaseDir); err != nil {
				return fmt.Errorf("updating current symlink: %w", err)
			}

			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PostActivate, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("post_activate hooks: %w", err)
			}
```

- [x] **Step 3: Build and run unit tests**

Run: `go build ./... && go test ./...`
Expected: PASS — full build green, all unit tests pass (180+ tests across all packages).

- [x] **Step 4: Lint**

Run: `mise run lint:check`
Expected: PASS — 0 issues.

- [x] **Step 5: Commit**

```bash
git add internal/config/schema.go internal/config/schema_test.go \
  internal/config/loader.go internal/config/loader_test.go \
  internal/config/merge.go internal/config/merge_test.go \
  internal/config/flatgen_test.go \
  internal/strategy/atomic/deployer.go internal/hooks/runner.go \
  internal/cmd/deploy.go \
  internal/cmd/release/activate.go internal/cmd/release/rollback.go \
  testdata/bifrost-full.yml testdata/bifrost-flat.yml
git commit -m "feat(hooks): expand hook lifecycle to 8 symmetric pre/post stages

Renames pre_enable_release/post_enable_release to pre_activate/post_activate
to match the release activate/rollback verb, and adds pre_extract, post_link,
pre_purge, post_purge for full deploy pipeline coverage. Breaking change: old
hook names are rejected by strict YAML parsing with no alias."
```

---

### Task 8: Extend the deploy hooks integration test to cover all 8 hook points and verify ordering

**Files:**
- Modify: `testdata/bifrost-deploy-hooks-test.yml`
- Modify: `testdata/bifrost-deploy-sudo-test.yml`
- Modify: `internal/cmd/deploy_integration_test.go` (`TestDeployCmd_Hooks`)

- [x] **Step 1: Update `testdata/bifrost-deploy-sudo-test.yml`**

Replace:

```yaml
strategy: atomic
paths:
  releases_root: /var/releases
  shared_root: /var/shared
settings:
  releases_to_keep: 5
environments:
  test:
    applications:
      app:
        hooks:
          post_enable_release:
            - cmd: systemctl reload nginx
              sudo: true
```

with:

```yaml
strategy: atomic
paths:
  releases_root: /var/releases
  shared_root: /var/shared
settings:
  releases_to_keep: 5
environments:
  test:
    applications:
      app:
        hooks:
          post_activate:
            - cmd: systemctl reload nginx
              sudo: true
```

- [x] **Step 2: Write the failing test — update the hooks fixture**

Replace the contents of `testdata/bifrost-deploy-hooks-test.yml`:

```yaml
strategy: atomic
paths:
  releases_root: /var/releases
  shared_root: /var/shared
settings:
  releases_to_keep: 5
hooks:
  pre_extract:
    - cmd: "echo pre_extract >> /tmp/hook_order.log"
  post_extract:
    - cmd: "echo post_extract >> /tmp/hook_order.log"
  pre_link:
    - cmd: "echo pre_link >> /tmp/hook_order.log"
  post_link:
    - cmd: "echo post_link >> /tmp/hook_order.log"
  pre_activate:
    - cmd: "echo pre_activate >> /tmp/hook_order.log"
  post_activate:
    - cmd: "echo post_activate >> /tmp/hook_order.log"
  pre_purge:
    - cmd: "echo pre_purge >> /tmp/hook_order.log"
  post_purge:
    - cmd: "echo post_purge >> /tmp/hook_order.log"
environments:
  test:
    applications:
      app:
        paths:
          shared:
            directories:
              - var/log
            files:
              - .env
```

- [x] **Step 3: Write the failing test — update `TestDeployCmd_Hooks`**

Replace `TestDeployCmd_Hooks` in `internal/cmd/deploy_integration_test.go`:

```go
func TestDeployCmd_Hooks(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-hooks-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost-hooks.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost-hooks.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "hooks-r1",
		"--init",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "deploy output:\n%s", result.Output)

	// All 8 hooks must have fired, in pipeline order.
	res, err := c.Exec(ctx, []string{"cat", "/tmp/hook_order.log"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "hook order log should exist")

	want := "pre_extract\npost_extract\npre_link\npost_link\npre_activate\npost_activate\npre_purge\npost_purge"
	assert.Equal(t, want, strings.TrimSpace(res.Output), "hooks must fire in pipeline order")
}
```

- [x] **Step 4: Run the test to verify it passes**

Run: `go test -tags integration ./internal/cmd/... -run TestDeployCmd_Hooks -v`
Expected: PASS — all 8 hook markers appear in `/tmp/hook_order.log` in pipeline order
(this confirms Task 5's implementation is correct end-to-end; requires Docker).

- [x] **Step 5: Run the full integration suite to check for regressions**

Run: `go test -tags integration ./...`
Expected: PASS — all integration tests green, including `TestDeployCmd_E2E`,
`TestDeployCmd_JSONOutput`, `TestDeployCmd_Purge`, `TestDeployCmd_DryRun_SudoHook`
(still asserts `"systemctl reload nginx"` and `"(sudo)"` — those substrings are
unchanged by the `post_enable_release` → `post_activate` rename in the fixture).

- [x] **Step 6: Commit**

```bash
git add testdata/bifrost-deploy-hooks-test.yml testdata/bifrost-deploy-sudo-test.yml \
  internal/cmd/deploy_integration_test.go
git commit -m "test(cmd): cover all 8 hook lifecycle points and verify deploy ordering"
```

---

### Task 9: Update `bifrost.schema.json` for the 8-hook shape

**Files:**
- Modify: `bifrost.schema.json` (`#/definitions/hooks`)

- [x] **Step 1: Replace the `hooks` definition**

In `bifrost.schema.json`, find the `hooks` definition (under `definitions`):

```json
"hooks": {
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "post_extract": {
      "$ref": "#/definitions/hookList"
    },
    "pre_link": {
      "$ref": "#/definitions/hookList"
    },
    "pre_enable_release": {
      "$ref": "#/definitions/hookList"
    },
    "post_enable_release": {
      "$ref": "#/definitions/hookList"
    }
  }
}
```

Replace its `properties` with:

```json
"hooks": {
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "pre_extract": {
      "$ref": "#/definitions/hookList"
    },
    "post_extract": {
      "$ref": "#/definitions/hookList"
    },
    "pre_link": {
      "$ref": "#/definitions/hookList"
    },
    "post_link": {
      "$ref": "#/definitions/hookList"
    },
    "pre_activate": {
      "$ref": "#/definitions/hookList"
    },
    "post_activate": {
      "$ref": "#/definitions/hookList"
    },
    "pre_purge": {
      "$ref": "#/definitions/hookList"
    },
    "post_purge": {
      "$ref": "#/definitions/hookList"
    }
  }
}
```

- [x] **Step 2: Validate the JSON is well-formed**

Run: `python3 -c "import json; json.load(open('bifrost.schema.json'))" && echo OK`
Expected: `OK`

- [x] **Step 3: Commit**

```bash
git add bifrost.schema.json
git commit -m "docs(schema): update bifrost.schema.json for 8-hook lifecycle"
```

---

### Task 10: Update `config init` scaffold

**Files:**
- Modify: `internal/cmd/config/init.go` (`scaffold` const)

- [x] **Step 1: Check for scaffold content assertions**

Run: `/usr/bin/grep -rn "pre_enable_release\|post_enable_release\|hooks:" internal/cmd/config/*_test.go`

If any test asserts on the literal old hook names in the scaffold output, note the file —
it will need the same renames applied in Step 3.

- [x] **Step 2: Update the `hooks:` block in `scaffold`**

In `internal/cmd/config/init.go`, replace:

```go
hooks:
  post_extract: [] # After extraction, before shared linking
  pre_link: [] # Before shared resource linking
  pre_enable_release: [] # Before current symlink update
  post_enable_release: [] # After current symlink update
```

with:

```go
hooks:
  pre_extract: [] # Before artifact extraction
  post_extract: [] # After extraction, before shared linking
  pre_link: [] # Before shared resource linking
  post_link: [] # After shared resource linking
  pre_activate: [] # Before current symlink update
  post_activate: [] # After current symlink update
  pre_purge: [] # Before old-release purge
  post_purge: [] # After old-release purge
```

- [x] **Step 3: Update the example app hooks block in `scaffold`**

In the same file, replace:

```go
        hooks:
          post_enable_release:
            - cmd: "systemctl reload php-fpm"
              sudo: true
              priority: 10
```

with:

```go
        hooks:
          post_activate:
            - cmd: "systemctl reload php-fpm"
              sudo: true
              priority: 10
```

- [x] **Step 4: Run config tests**

Run: `go test ./internal/cmd/config/...`
Expected: PASS

- [x] **Step 5: Run `go build` and `mise run lint:check`**

Run: `go build ./... && mise run lint:check`
Expected: PASS

- [x] **Step 6: Commit**

```bash
git add internal/cmd/config/init.go
git commit -m "feat(config): update config init scaffold for 8-hook lifecycle"
```

---

### Task 11: Update `docs/specs/02-configuration.md`

**Files:**
- Modify: `docs/specs/02-configuration.md` ("Full schema" hooks block, "Hook lifecycle"
  table, "Example config")

- [x] **Step 1: Update the "Full schema" hooks block**

Replace (lines 74–78):

```yaml
hooks:
  post_extract: []        # Runs after extraction, before pre_link — raw release dir available
  pre_link: []            # Runs before shared resource linking
  pre_enable_release: []  # Runs before current symlink is updated
  post_enable_release: [] # Runs after current symlink is updated
```

with:

```yaml
hooks:
  pre_extract: []    # Runs before artifact extraction
  post_extract: []   # Runs after extraction, before pre_link — raw release dir available
  pre_link: []       # Runs before shared resource linking
  post_link: []      # Runs after shared resource linking
  pre_activate: []   # Runs before current symlink is updated
  post_activate: []  # Runs after current symlink is updated
  pre_purge: []      # Runs before old-release purge
  post_purge: []     # Runs after old-release purge
```

- [x] **Step 2: Update the "Hook lifecycle" table**

Replace (lines 110–118):

```markdown
## Hook lifecycle

| List | When it runs |
|---|---|
| `post_extract` | After artifact extraction, before `pre_link` — raw release dir available |
| `pre_link` | After `post_extract`, before shared resource linking |
| `pre_enable_release` | After shared linking, before `current` symlink update |
| `post_enable_release` | After `current` symlink update |
```

with:

```markdown
## Hook lifecycle

The `deploy` pipeline fires all 8 hooks, in this order:

```
pre_extract → [extract]      → post_extract
pre_link    → [link shared]  → post_link
pre_activate → [switch current] → post_activate
pre_purge   → [purge]        → post_purge
```

| List | When it runs |
|---|---|
| `pre_extract` | Before artifact extraction |
| `post_extract` | After artifact extraction, before `pre_link` — raw release dir available |
| `pre_link` | After `post_extract`, before shared resource linking |
| `post_link` | After shared resource linking |
| `pre_activate` | After `post_link`, before `current` symlink update |
| `post_activate` | After `current` symlink update |
| `pre_purge` | Before old-release purge |
| `post_purge` | After old-release purge |

`release activate` and `release rollback` only fire `pre_activate` / `post_activate`,
around their existing `current` symlink update — they don't extract or purge, so the
other six hooks don't apply to those commands.
```

- [x] **Step 3: Update the "Example config"**

Replace the `hooks:` block in the example config (lines 146–156):

```yaml
        hooks:
          post_extract:
            - cmd: "composer install --no-dev --optimize-autoloader"
              cmd_dir: "{{ .Directories.Working }}"
          post_enable_release:
            - cmd: "systemctl reload php-fpm"
              sudo: true
              priority: 10
            - cmd: "systemctl reload nginx"
              sudo: true
              priority: 20
```

with:

```yaml
        hooks:
          post_extract:
            - cmd: "composer install --no-dev --optimize-autoloader"
              cmd_dir: "{{ .Directories.Working }}"
          post_activate:
            - cmd: "systemctl reload php-fpm"
              sudo: true
              priority: 10
            - cmd: "systemctl reload nginx"
              sudo: true
              priority: 20
```

- [x] **Step 4: Commit**

```bash
git add docs/specs/02-configuration.md
git commit -m "docs(specs): update spec 02 for 8-hook lifecycle"
```

---

### Task 12: Rewrite `docs/specs/05-hooks.md` lifecycle table

**Files:**
- Modify: `docs/specs/05-hooks.md` ("Lifecycle points" table)

- [x] **Step 1: Replace the "Lifecycle points" section**

Replace:

```markdown
## Lifecycle points

| Hook list | Shared symlinks present? | `current` updated? | Typical use |
|---|---|---|---|
| `post_extract` | No | No | Compile assets, modify extracted files |
| `pre_link` | No | No | Remove dirs that will be replaced by symlinks, create placeholder subdirs |
| `pre_enable_release` | **Yes** | No | DB migrations, config cache — needs shared files, must run before traffic hits release |
| `post_enable_release` | Yes | **Yes** | Reload services, notify monitoring |

`pre_link` and `pre_enable_release` differ in one critical way: shared dirs/files
(`var/log`, `.env`, etc.) are **not yet symlinked** at `pre_link` time, but are
**fully in place** by `pre_enable_release`. Run anything that requires shared config
files (migrations, cache warming) in `pre_enable_release`, not `pre_link`.
```

with:

```markdown
## Lifecycle points

The `deploy` command fires all 8 hooks below, in order. `release activate` and
`release rollback` only fire `pre_activate` / `post_activate`, around their existing
`current` symlink update.

```
pre_extract → [extract]      → post_extract
pre_link    → [link shared]  → post_link
pre_activate → [switch current] → post_activate
pre_purge   → [purge]        → post_purge
```

| Hook list | Shared symlinks present? | `current` updated? | Typical use |
|---|---|---|---|
| `pre_extract` | No | No | Prepare the empty release directory before extraction |
| `post_extract` | No | No | Compile assets, modify extracted files |
| `pre_link` | No | No | Remove dirs that will be replaced by symlinks, create placeholder subdirs |
| `post_link` | **Yes** | No | DB migrations, config cache — needs shared files, must run before traffic hits release |
| `pre_activate` | Yes | No | Final checks immediately before the symlink switch |
| `post_activate` | Yes | **Yes** | Reload services, notify monitoring |
| `pre_purge` | Yes | Yes | Snapshot or archive a release directory before it may be purged |
| `post_purge` | Yes | Yes | Cleanup or notification after old releases are removed |

`pre_link` and `post_link` differ in one critical way: shared dirs/files (`var/log`,
`.env`, etc.) are **not yet symlinked** at `pre_link` time, but are **fully in place**
by `post_link`. Run anything that requires shared config files (migrations, cache
warming) in `post_link`, not `pre_link`.
```

- [x] **Step 2: Commit**

```bash
git add docs/specs/05-hooks.md
git commit -m "docs(specs): rewrite spec 05 hook lifecycle table for 8 hooks"
```

---

### Task 13: Update `docs/bifrost.sample.yml`

**Files:**
- Modify: `docs/bifrost.sample.yml` (hooks section, lines 103–116; example app hooks,
  lines 162–172)

- [x] **Step 1: Update the annotated `hooks:` block**

Replace (lines 103–116):

```yaml
hooks:
  # post_extract — runs after artifact extraction, before pre_link.
  # The raw release directory is available.
  post_extract: []

  # pre_link — runs after post_extract, before shared resource linking.
  pre_link: []

  # pre_enable_release — runs after shared linking, before the current
  # symlink is updated.
  pre_enable_release: []

  # post_enable_release — runs after the current symlink is updated.
  post_enable_release: []
```

with:

```yaml
hooks:
  # pre_extract — runs before artifact extraction.
  pre_extract: []

  # post_extract — runs after artifact extraction, before pre_link.
  # The raw release directory is available.
  post_extract: []

  # pre_link — runs after post_extract, before shared resource linking.
  pre_link: []

  # post_link — runs after shared resource linking. Shared dirs/files
  # (var/log, .env, etc.) are in place — use this for DB migrations,
  # cache warming, anything that needs shared config.
  post_link: []

  # pre_activate — runs after post_link, before the current symlink
  # is updated.
  pre_activate: []

  # post_activate — runs after the current symlink is updated.
  post_activate: []

  # pre_purge — runs before old releases are purged.
  pre_purge: []

  # post_purge — runs after old releases are purged.
  post_purge: []
```

- [x] **Step 2: Update the example app `hooks:` block**

Replace (lines 162–172):

```yaml
        hooks:
          post_extract:
            - cmd: "composer install --no-dev --optimize-autoloader"
              cmd_dir: "{{ .Directories.Working }}"
          post_enable_release:
            - cmd: "systemctl reload php-fpm"
              sudo: true
              priority: 10
            - cmd: "systemctl reload nginx"
              sudo: true
              priority: 20
```

with:

```yaml
        hooks:
          post_extract:
            - cmd: "composer install --no-dev --optimize-autoloader"
              cmd_dir: "{{ .Directories.Working }}"
          post_activate:
            - cmd: "systemctl reload php-fpm"
              sudo: true
              priority: 10
            - cmd: "systemctl reload nginx"
              sudo: true
              priority: 20
```

- [x] **Step 3: Commit**

```bash
git add docs/bifrost.sample.yml
git commit -m "docs: update bifrost.sample.yml for 8-hook lifecycle"
```

---

### Task 14: Write ADR-0012

**Files:**
- Create: `docs/adr/0012-hook-lifecycle-granularity.md`
- Modify: `docs/adr/README.md` (index table)

- [x] **Step 1: Create the ADR**

Create `docs/adr/0012-hook-lifecycle-granularity.md`:

```markdown
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

- Design spec: [`docs/specs/2026-06-12-hook-lifecycle-design.md`](../specs/2026-06-12-hook-lifecycle-design.md)
- [Spec 02 — Configuration](../specs/02-configuration.md)
- [Spec 05 — Hooks](../specs/05-hooks.md)
```

- [x] **Step 2: Add ADR-0012 to the index**

In `docs/adr/README.md`, add a row after the ADR-0011 row:

```markdown
| [0012](0012-hook-lifecycle-granularity.md) | Hook lifecycle granularity — 8 symmetric pre/post stages | Accepted |
```

- [x] **Step 3: Commit**

```bash
git add docs/adr/0012-hook-lifecycle-granularity.md docs/adr/README.md
git commit -m "docs(adr): record hook lifecycle granularity decision (ADR-0012)"
```

---

### Task 15: Add roadmap milestone M18

**Files:**
- Modify: `docs/tasks/v1-roadmap.md` (append new milestone after M17)

- [x] **Step 1: Append M18 to `docs/tasks/v1-roadmap.md`**

After the M17 section (ends with "Deliverable: CI passes on all v1 integration tests;
no known edge cases unhandled."), append:

```markdown

---

## M18 — Hook lifecycle granularity

Replace the 4-hook lifecycle (`post_extract`, `pre_link`, `pre_enable_release`,
`post_enable_release`) with 8 symmetric `pre_<stage>`/`post_<stage>` hooks covering
the full `deploy` pipeline, and rename the activation pair to align with the
`activate` verb used by `release activate`/`release rollback`.

- [x] `internal/config/schema.go` — `Hooks` struct: rename `pre_enable_release` →
  `pre_activate`, `post_enable_release` → `post_activate`; add `pre_extract`,
  `post_link`, `pre_purge`, `post_purge`
- [x] `internal/config/loader.go` — `applyHookDefaults` sets default priority for all
  8 hook fields
- [x] `internal/config/merge.go` — `Merge()` threads all 8 hook fields through
  `sortedHooks(...)`
- [x] `internal/strategy/atomic/deployer.go` — fire all 8 hooks around the `deploy`
  pipeline stages (extract, link, activate, purge); update `deployStepTotal`
- [x] `internal/cmd/deploy.go` — `--dry-run` preview reflects all 8 hooks
- [x] `internal/cmd/release/activate.go`, `rollback.go` — rename
  `PreEnableRelease`/`PostEnableRelease` → `PreActivate`/`PostActivate`
- [x] Integration test: all 8 hooks fire in pipeline order for `deploy`
- [x] `bifrost.schema.json`, `config init` scaffold, `docs/specs/02-configuration.md`,
  `docs/specs/05-hooks.md`, `docs/bifrost.sample.yml` updated for the new 8-hook
  shape
- [x] ADR-0012 records the naming and coverage decision

Spec references: [ADR-0012](../adr/0012-hook-lifecycle-granularity.md),
[design spec](../specs/2026-06-12-hook-lifecycle-design.md), [Spec 02](../specs/02-configuration.md),
[Spec 05](../specs/05-hooks.md).

Deliverable: `bifrost deploy` fires all 8 hooks in pipeline order; `release
activate`/`rollback` use the renamed `pre_activate`/`post_activate`; old hook names
are rejected with a clear config error.
```

- [x] **Step 2: Commit**

```bash
git add docs/tasks/v1-roadmap.md
git commit -m "docs(tasks): mark M18 hook lifecycle granularity complete"
```

Per the project's roadmap workflow (`.claude/rules/workflow.md`), this final commit
should be the last step of the overall task — implementation and roadmap update
together.
