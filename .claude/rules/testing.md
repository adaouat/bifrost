# Testing rules

## Hard constraints — never break these

1. **No filesystem writes on the local host.** Forbidden in test code: `os.MkdirAll`,
   `os.Mkdir`, `os.Create`, `os.WriteFile`, `os.Symlink`, `os.Remove`, `os.RemoveAll`,
   `t.TempDir()`. Two exceptions: writing the compiled binary to `/tmp`; and deployer
   FS-logic unit tests (`internal/strategy/atomic/`), which may use `t.TempDir()` for a real
   local filesystem because the deployer relies on symlinks and afero cannot model them.

2. **No shell execution on the local host.** `sh -c` and any `exec.Command` that runs
   real commands must never execute on the local machine during tests. All shell execution
   happens inside containers.

## Two test categories

### Unit tests — `go test ./...`

- No Docker required.
- Pure logic only: config parsing, merge, validation, template rendering.
- `exec.Command` is mocked — never real.
- Fixtures are read from `testdata/` (read-only, committed to repo).
- No filesystem writes of any kind.

### Integration tests — `go test -tags integration ./...`

- Requires Docker (Docker Desktop / Colima on macOS, pre-installed on GitHub Actions).
- All filesystem operations and shell execution happen inside Linux containers.
- Uses **testcontainers-go** for container lifecycle management.
- Every integration test file carries `//go:build integration` at the top.
- `TestMain` cross-compiles the `bifrost` binary once (`GOOS=linux GOARCH=amd64`) to `/tmp`
  and shares it across all tests in the package.
- Container is started fresh per test (or per suite) and destroyed after.
- Container readiness gates on a setup-complete marker the startup script writes last (e.g.
  `/tmp/.setup-complete`), **not** `wait.ForListeningPort` — Docker's port proxy can report
  the port open before the script finishes (e.g. `useradd`), racing test setup on CI. See
  `internal/testutil/ssh_container.go`.
- Assertions inspect the container filesystem state via container exec or file copy.

## TDD

Write the failing test before writing implementation code. No exceptions.
The test defines the expected behaviour. Implementation makes it pass, nothing more.

## No afero

Do not use `spf13/afero` or any in-memory filesystem abstraction. Bifrost's core relies on
symlinks; `MemMapFs` does not support them. Integration tests use a real Linux filesystem
inside containers.

## Test helpers

Shared container setup, binary building, and assertion helpers live in `internal/testutil/`.
**Never duplicate container boilerplate across test files.** If a second test needs the
same setup, it goes into `internal/testutil/` first, then both tests call it. No exceptions.

## Fixtures

`testdata/` holds read-only input fixtures: YAML config files, small `.tar.gz` archives.
Nothing is written to `testdata/` during test runs.
