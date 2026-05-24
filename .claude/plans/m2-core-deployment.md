# Plan: M2 — Core Deployment

## Context

M2 implements the core deployment engine: archive extraction, shared resource linking, release
directory management, and wires them into the `bifrost deploy` command. No hooks (those are M3).
Deliverable: `bifrost deploy` deploys a real archive end-to-end.

The command name is `deploy` (confirmed by user). The existing stub in `internal/cmd/deploy.go`
already has the correct flags.

---

## Task order (one commit each, [x] + roadmap per commit)

### Task 1 — `internal/strategy/atomic/deploy.go`: archive extraction

**Add dependency first:** `go get github.com/mholt/archives` (pin exact version in go.mod).

Function signature:
```go
func Extract(ctx context.Context, artifactPath, destDir string, onProgress func(n int64)) error
```

- Opens artifact, calls `archives.Identify` to detect format (tar.gz, zip, etc.)
- Asserts format implements `archives.Extractor`; returns error if not
- Extracts each entry via `archives.FileSystemExtractor{Root: destDir}`
- `onProgress` wraps the reader via a counting `io.Reader` if non-nil (task 6 will wire it)

**Unit tests** (`deploy_test.go`, no build tag):
- `TestExtract_ArtifactNotFound` — non-existent path → error wrapping `os.ErrNotExist`
- `TestExtract_UnsupportedFormat` — a plain `.txt` file → error about unsupported format

Commit: `feat(strategy/atomic): implement archive extraction`

---

### Task 2 — `internal/strategy/atomic/shared.go`: shared resource linking

Function signature:
```go
func LinkShared(dirs []string, files []string, releaseDir, sharedRoot string) error
```

Implements spec 04 verbatim, calling `linkOne(relPath, releaseDir, sharedRoot, isDir)`:

```
target = sharedRoot/relPath
link   = releaseDir/relPath

if target does not exist:
    if link exists: os.Rename(link, target)       // seed from release
    else if isDir:  os.MkdirAll(target, 0755)
    else:           touch target (create empty)
os.RemoveAll(link)
os.MkdirAll(parent(link), 0755)
os.Symlink(target, link)                          // absolute path
```

**Tests:** no standalone unit tests (all FS ops); covered by task 4 integration test.

Commit: `feat(strategy/atomic): implement shared resource linking`

---

### Task 3 — `internal/strategy/atomic/release.go`: release management

Three functions:

```go
// CreateReleaseDir creates {releasesRoot}/{name} (or timestamp if name=="").
func CreateReleaseDir(releasesRoot, name string) (string, error)

// SetCurrent atomically updates {releasesRoot}/current via tmp-symlink + rename.
func SetCurrent(releasesRoot, releaseDir string) error

// Purge deletes oldest releases, keeping keepN most recent (by dir name sort).
func Purge(releasesRoot string, keepN int) error
```

`Purge` sorts dir names lexicographically (timestamp names sort chronologically),
skips `current`, deletes from the front until `len(releases) <= keepN`.

**Tests:** no standalone unit tests; covered by task 4 integration test.

Commit: `feat(strategy/atomic): implement release directory management`

---

### Task 4 — `deploy` command + integration test

**Write the failing integration test first**, then implement to make it pass.

**Test file:** `internal/cmd/deploy_integration_test.go` (`//go:build integration`)

```
TestMain  → testutil.BuildBifrostBinary()
TestDeployCmd_E2E:
  1. Start container with bifrost binary
  2. CopyFile config fixture → /tmp/bifrost.yml
  3. CopyFile testdata/release.tar.gz → /tmp/release.tar.gz
  4. RunBifrost "deploy --config /tmp/bifrost.yml --env test --app app
                         --artifact /tmp/release.tar.gz --release-name test-r1 --init"
  5. Assert exit 0
  6. Assert /var/releases/test-r1/public/index.html exists
  7. Assert readlink /var/releases/current == /var/releases/test-r1
  8. Assert /var/releases/test-r1/var/log is symlink → /var/shared/var/log
  9. Assert /var/releases/test-r1/.env is symlink → /var/shared/.env
```

**New testdata fixture:** `testdata/bifrost-deploy-int-test.yml`
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
        paths:
          shared:
            directories:
              - var/log
            files:
              - .env
```

**Enhance testutil:** add `CopyFile(ctx, content []byte, containerPath string, mode int64) error`
wrapping `c.inner.CopyToContainer`.

**Updated `internal/cmd/deploy.go`:**
```go
RunE: func(cmd *cobra.Command, _ []string) error {
    // validate required flags
    // 1. Load + merge + validate config (exit 2 on validation error)
    // 2. Ensure roots exist (exit 3 if missing and --init not set)
    // 3. Stat artifact (exit 3 if not found)
    // 4. CreateReleaseDir
    // 5. Extract (onProgress=nil for now)
    // 6. LinkShared
    // 7. SetCurrent
    // 8. Purge
}
```

Config is read via `resolveConfigPath(cmd.Root())` (existing helper).

Commit: `feat(cmd/deploy): implement deploy command — full flow, no hooks`

---

### Task 5 — `internal/tui/progress.go`: spinner + progress bar wrappers

**Add dependencies:** `github.com/charmbracelet/huh` (for spinner), `github.com/charmbracelet/bubbles`
(for progress bar). Pin exact versions.

Two exported helpers:
```go
// RunWithSpinner runs fn wrapped in a spinner (human+TTY only; falls back to plain fn() otherwise).
func RunWithSpinner(title string, fn func() error) error

// NewProgressBar returns a progress bar model and a writer that reports progress.
func NewProgressBar(total int64) (*progress.Model, io.Writer)
```

TTY detection: `golang.org/x/term.IsTerminal(int(os.Stdout.Fd()))` — no new dep (already transitively
available via charmbracelet/x/term in go.sum).

**Tests:** none (pure TUI code).

Commit: `feat(tui): add spinner and progress bar wrappers`

---

### Task 6 — Wire progress bar to extraction

- In `deploy.go` command: pass an `onProgress` callback to `Extract` that feeds the bubbles
  progress bar (updates the model, writes to stdout).
- Guard with output-mode + TTY check via `tui` package.
- The integration test implicitly covers this (deploy still succeeds with or without TTY).

Commit: `feat(cmd/deploy): wire progress bar to extraction byte stream`

---

## Critical files

| File | Action |
|---|---|
| `internal/strategy/atomic/deploy.go` | Create |
| `internal/strategy/atomic/deploy_test.go` | Create (unit) |
| `internal/strategy/atomic/shared.go` | Create |
| `internal/strategy/atomic/release.go` | Create |
| `internal/cmd/deploy.go` | Rewrite stub |
| `internal/cmd/deploy_integration_test.go` | Create |
| `internal/testutil/container.go` | Add `CopyFile` method |
| `internal/tui/progress.go` | Create |
| `testdata/bifrost-deploy-int-test.yml` | Create |
| `go.mod` / `go.sum` | Add mholt/archives, huh, bubbles |
| `docs/tasks/roadmap.md` | Mark tasks [x] per commit |

---

## Verification

```bash
# Unit tests (tasks 1-3 error cases)
mise run test

# Integration test (tasks 1-4 full flow)
go test -tags integration ./internal/cmd/...

# Lint
mise run lint:check

# Manual smoke test
mise run build
./bifrost deploy --config testdata/bifrost-full.yml --env prod --app web \
  --artifact testdata/release.tar.gz --release-name smoke-test --init
```
