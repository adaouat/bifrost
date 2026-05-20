# ADR-0007: Testing strategy

**Status:** Accepted

## Context

Bifrost is a deployment tool whose core operations are filesystem writes (directory
creation, symlink management, archive extraction) and shell command execution (hooks via
`sh -c`). Testing these operations naively would:

- Create and delete directories and files on the developer's local machine
- Execute arbitrary shell commands on the local machine
- Make tests environment-dependent and potentially destructive

Two hard constraints drive this strategy:

1. **No filesystem writes on the local host during test execution** — no `os.MkdirAll`,
   `os.Create`, `os.Symlink`, `os.Remove`, no `t.TempDir()`.
   Writing a compiled binary to `/tmp` (cross-compilation artifact) is acceptable.
2. **No shell command execution on the local host during test execution** — `sh -c` from
   hook execution must never run on the local machine.

## Decision

Split tests into two categories separated by a build tag:

| Category | Build tag | Requires Docker | Constraint |
|---|---|---|---|
| **Unit tests** | _(none)_ | No | No filesystem writes, no shell execution |
| **Integration tests** | `integration` | Yes | All filesystem and shell ops run inside containers |

### Unit tests — local, zero side effects

Run with `go test ./...`. Must contain **only** pure logic:

- Config loading (reads `testdata/` fixture files — read-only, allowed)
- Config merge and validation
- Template rendering for hook commands
- Hook runner: template expansion only (`exec.Command` is mocked)

No `t.TempDir()`. No `os.*` write calls. No `exec.Command` for real.

### Integration tests — testcontainers-go

Run with `go test -tags integration ./...`. Require Docker.

All tests that touch the filesystem or execute shell commands use
**[testcontainers-go](https://github.com/testcontainers/testcontainers-go)** to spin up
an isolated Linux container. Operations happen inside the container, never on the host.

#### Why testcontainers over in-memory filesystem (afero)

`spf13/afero` `MemMapFs` does not support symlinks. Symlinks are not an edge case in
Bifrost — the `current` pointer and all shared resource links are symlinks. Testing with
a fake filesystem would provide false confidence. A real Linux filesystem (inside a
container) is required.

#### Binary delivery — Option B (cross-compile on host, copy into container)

`TestMain` cross-compiles the `bifrost` binary for `linux/amd64` once before any test
runs:

```go
func TestMain(m *testing.M) {
    bin = buildBifrostBinary() // GOOS=linux GOARCH=amd64 go build → /tmp/bifrost-test-<hash>
    code := m.Run()
    os.Remove(bin)
    os.Exit(code)
}
```

The compiled binary path is shared across all container tests. Each test copies it into
a fresh container. This avoids rebuilding per test and avoids any in-project filesystem
writes (output goes to `/tmp`).

#### Test structure

Each integration test:
1. Starts a Linux container (alpine or debian-slim)
2. Copies the `bifrost` binary and a fixture archive into the container
3. Writes a test `.bifrost.yml` into the container
4. Executes `bifrost <command>` inside the container
5. Asserts on container filesystem state (dirs exist, symlinks point correctly, etc.)
6. Container is automatically destroyed after the test

#### Fixture archives

Small `.tar.gz` archives for testing are committed to `testdata/`. They are read-only
inputs — never generated during test runs.

### v1 — SSH agent model

When v1 introduces SSH transport, integration tests will use a two-container setup:
- `server` container: runs `sshd`, acts as the deployment target
- Bifrost runs locally, SSHs into `server` using testcontainers-go's exposed ports

testcontainers-go supports multi-container scenarios natively.

## Build tag convention

```go
//go:build integration
```

All files containing integration tests carry this tag. The `integration` tag is never
set in production code.

## CI

```yaml
# Unit tests — always run
- run: go test ./...

# Integration tests — requires Docker (available on ubuntu runners)
- run: go test -tags integration ./...
```

## Consequences

- `go test ./...` always works without Docker — fast feedback loop for pure logic changes.
- `go test -tags integration ./...` requires Docker (Docker Desktop or Colima on macOS,
  pre-installed on GitHub Actions ubuntu runners).
- `exec.Command` in the hook runner must be injectable / mockable for unit tests.
- All test helpers that spin up containers live in `internal/testutil/`.
- `testdata/` holds read-only fixtures (YAML configs, small archives). Nothing is written
  there during test runs.
