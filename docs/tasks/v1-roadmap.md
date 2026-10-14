# v1 Roadmap — SSH orchestration + agent model

Milestones in implementation order. Each builds on the previous.

Same working process as v0: two-step commit flow, TDD, `hk check` before every commit.
See [`roadmap.md`](roadmap.md) for the process rules.

---

## M11 — Strategy interface

Introduce the `strategy.go` interface that ADR-0006 deferred to v1. The atomic strategy
becomes the first typed implementation. No behaviour change.

- [x] `internal/strategy/strategy.go` — `Deployer` interface with `Deploy(ctx, opts) error`
- [x] `internal/strategy/atomic/` implements `Deployer`
- [x] `internal/cmd/deploy.go` uses the interface, not the concrete type

Deliverable: `go build` passes; no behaviour change; interface is the extension point for
future strategies.

---

## M12 — Server config schema

Add SSH server definitions and server references to the config layer. No SSH connections yet.

- [x] `internal/config/schema.go` — add `ServerConfig` struct; add `Servers []string` to
  `EnvConfig` and `AppConfig`
- [x] `internal/config/loader.go` — flat config path: if no `environments:` key, load
  top-level directly; skip `--env`/`--app` requirement
- [x] `internal/config/loader.go` — validate server references: every name in `servers:`
  lists must exist in the top-level `servers:` map (exit code 2)
- [x] `internal/config/merge.go` — server resolution: app-level `servers` overrides
  env-level; `nil` inherits from parent
- [x] `internal/cmd/deploy.go` — detect client mode: resolved `servers` list non-empty
- [x] Update `config show` and `config check` to display/validate server entries
- [x] Update `testdata/` fixtures and unit tests

Spec references: [Spec 02](../specs/02-configuration.md), [Spec 07](../specs/07-ssh-transport.md),
[ADR-0010](../adr/0010-server-config-schema.md).

Deliverable: config with `servers:` block loads, validates, and displays correctly.
Local mode (no servers) is unchanged.

---

## M13 — SSH transport layer

Low-level SSH and SFTP plumbing. No deploy logic yet.

- [x] `internal/transport/ssh.go` — SSH client wrapper: connect, exec (returns stdout/stderr
  readers + exit code), close. Auth: key file → SSH agent → `BIFROST_SSH_PASSWORD` env var.
  Known-hosts: system `~/.ssh/known_hosts`, strict.
- [x] `internal/transport/sftp.go` — SFTP wrapper: upload file, mkdir, chmod
- [x] `internal/transport/staging.go` — staging dir lifecycle: create (`/tmp/bifrost-{uuid}/`
  or configured base), upload contents, cleanup (`rm -rf`)
- [x] `internal/transport/agent.go` — arch detection (`uname -s/m` via SSH), platform key
  mapping, binary download from GitHub Releases, local cache at
  `{os.UserCacheDir()}/bifrost/agents/{version}/{os}_{arch}/bifrost`
- [x] `--agent-binary <path>` flag wired into all SSH-capable commands (bypass download)
- [x] Unit tests for arch mapping and cache key logic (no real SSH)

Dependencies: `golang.org/x/crypto/ssh`, `github.com/pkg/sftp`

Spec references: [Spec 07](../specs/07-ssh-transport.md),
[ADR-0011](../adr/0011-agent-binary-distribution.md).

Deliverable: transport package compiles and unit tests pass. No integration yet.

---

## M14 — Client deploy flow

Wire the transport layer into the `deploy` command for the single-server case.

- [x] `internal/config/flatgen.go` — flat config generator: merge global < env < app,
  strip `environments:` and `servers:` keys, write YAML
- [x] `internal/cmd/deploy.go` — client mode branch:
  - Generate release name once (before server loop)
  - For each server: upload staging dir, exec agent, stream JSON → render, cleanup
- [x] JSON event stream reader: read agent stdout line by line, decode events, forward to
  existing TUI renderer with server header prefix
- [x] `--output json` re-emit with added `"server"` field
- [x] Integration test: testcontainers Linux container with OpenSSH server, deploy via
  client mode, assert filesystem state

Spec references: [Spec 07](../specs/07-ssh-transport.md),
[Spec 08](../specs/08-multi-server.md), [ADR-0009](../adr/0009-agent-model.md).

Deliverable: `bifrost deploy --env prod --app web --artifact app.tar.gz` works end-to-end
over SSH to a single server.

---

## M15 — Multi-server deploy

Extend M14 to the multi-server sequential loop.

- [x] Server loop in `deploy` iterates over all resolved servers sequentially
- [x] Failure on any server skips remaining servers; exits with agent's exit code
- [x] Human / plain mode: server header line before each server's output block
- [x] Integration test: two containers, sequential deploy, assert both filesystems

Spec references: [Spec 08](../specs/08-multi-server.md).

Deliverable: `bifrost deploy` works against multiple servers; output has per-server sections.

---

## M16 — Release commands over SSH

Port `release list`, `release activate`, and `release rollback` to client mode.

- [x] `release list` — SSH to each server, render per-server section (human) or events
  with `"server"` field (JSON)
- [x] `release activate` — non-interactive path: `--release <name>` runs activate on all
  servers; interactive path: query release lists → local huh multi-select form (one select
  per server) → run activate per selection
- [x] `release rollback` — SSH to each server, run rollback, render per-server result
- [x] Non-TTY guard for `release activate` without `--release` (exit code 1)
- [x] Integration tests for each command (single-server and multi-server)

Spec references: [Spec 08](../specs/08-multi-server.md).

Deliverable: all release management commands work over SSH; multi-server activate TUI
renders correctly.

---

## M17 — Quality and hardening

Fill test gaps and harden edge cases introduced by v1.

- [x] Unit tests for flat config generator (various merge scenarios)
- [x] Unit tests for server resolution (app overrides env, nil inheritance) — covered by `merge_test.go` (M12 TDD)
- [x] Unit tests for staging dir path construction — covered by `staging_test.go` (M13 TDD)
- [x] Integration test: `--agent-binary` flag with a pre-built binary (no download)
- [x] Integration test: SSH auth failure → clean error message, exit code 3
- [x] Integration test: unknown remote arch → clean error message, exit code 3
- [x] Integration test: agent exits non-zero mid-deploy → cleanup still runs
- [x] Review and update specs 01, 02, 03, 07, 08 for any gaps found during implementation
- [x] `bifrost.schema.json` — JSON Schema for `.bifrost.yml` (full v1 schema); update
  `config init` template to include `# yaml-language-server: $schema=` comment for IDE validation

Deliverable: CI passes on all v1 integration tests; no known edge cases unhandled.

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
[design notes](../../.claude/plans/m18-hook-lifecycle-design.md), [Spec 02](../specs/02-configuration.md),
[Spec 05](../specs/05-hooks.md).

Deliverable: `bifrost deploy` fires all 8 hooks in pipeline order; `release
activate`/`rollback` use the renamed `pre_activate`/`post_activate`; old hook names
are rejected with a clear config error.

---

## M19 — v1 hardening

Close the critical and important gaps surfaced by the full code/doc review (2026-06-13).
The defects cluster in the v1 SSH/agent path (M11–M18), which CI never exercises. Same
working process as the rest of v1: two-step commit flow, TDD (failing test first),
`hk check` before every commit. One task at a time.

### Agent binary download (Critical)

`.goreleaser.yml` ships raw binaries (`formats: [binary]`, asset
`bifrost_{version}_{os}_{arch}`, no archive), but `internal/transport/agent.go` requests a
`.tar.gz`, untars it, and re-prefixes an already-`v`-prefixed version — so auto-download
404s on every release and only `--agent-binary` works.

- [x] `internal/transport/agent.go` — `downloadURL` points at the raw binary asset (no
  `.tar.gz`); remove `extractBinary`/gzip+tar; cache the downloaded bytes directly
- [x] `internal/transport/agent.go` — normalise the version (strip a leading `v`) so the
  release-tag path and asset filename match goreleaser's `{{ .Tag }}` / `{{ .Version }}`
- [x] Test `downloadURL`/`downloadAgent` against a local HTTP server (no `--agent-binary`
  bypass), asserting the raw binary bytes are returned (the cache write is a local-FS op
  excluded by the no-write test rule)
- [x] Update [ADR-0011](../adr/0011-agent-binary-distribution.md) and
  [Spec 07](../specs/07-ssh-transport.md) to describe raw-binary distribution

### Agent binary integrity verification

The downloaded binary is cached and executed (often under `sudo`) with no integrity check;
goreleaser already publishes `checksums.txt`.

- [x] `internal/transport/agent.go` — fetch `checksums.txt` and verify the asset's SHA-256
  before writing it to cache; fail with a clear error on mismatch
- [x] Unit test: a tampered binary is rejected
- [x] Record the integrity decision in [ADR-0011](../adr/0011-agent-binary-distribution.md)

### Quote remote command arguments

Agent invocation strings interpolate `env`, `app`, `release-name`, and the artifact basename
straight into a remote shell, unquoted — they break on spaces and are injectable.

- [x] Add a POSIX shell-quoting helper in `internal/transport`
- [x] Apply it to every interpolated value in `internal/cmd/client_deploy.go`,
  `internal/cmd/release/client_activate.go`, `client_list.go`, `client_rollback.go`, and the
  `rm -rf` / `chmod` calls in `internal/transport/staging.go`
- [x] Integration test: a value containing a space and `;`, quoted by `ShellQuote`,
  round-trips as one literal argument over a real SSH shell and does not inject

### Run integration tests in CI

The reusable `forge/go-ci.yml` runs `go test ./...` without `-tags integration`, and
bifrost's `ci.yml` only builds — so the entire testcontainers suite is unguarded (this is
how the broken download path shipped).

- [x] `.github/workflows/ci.yml` — add a Docker-enabled job running
  `go test -tags integration ./...`
- [x] Pin any new action to a full commit SHA with a version comment
- [x] Fix the test-infra race the CI job exposed: `NewSSHContainer` waited on
  `wait.ForListeningPort`, which returns before the setup script's `useradd` on Linux CI, so
  `chown deploy:deploy` failed; readiness now gates on a `/tmp/.setup-complete` marker
  (`internal/testutil/ssh_container.go`, commit `e837a96`)

### Hook output over SSH / JSON mode

In `--output json` mode the deployer writes raw hook stdout/stderr to the same stream as the
JSON events (`internal/strategy/atomic/deployer.go`, `internal/hooks/runner.go`); over SSH
the client's NDJSON parser drops it (`internal/tui/stream.go`), so hook output is invisible
remotely.

- [x] Emit hook output as JSON events (`hook_output`) instead of raw writes to the
  protocol stream; render them in `internal/tui/stream.go`
- [x] Verify NDJSON stays valid and hook output surfaces: unit tests for the writer and
  `ForwardStream` rendering, plus a deployer test that runs the real pipeline in JSON mode
  with a fake runner (local shell exec is forbidden by the test rules), asserting every
  stream line is valid JSON and a `hook_output` event carries the hook's stdout. The container
  E2E for the SSH leg landed in M20 (S11)

### Exit-code classification for deployer failures

Hook, template, extraction, link, symlink, and purge failures return bare `fmt.Errorf` and
resolve to exit 1 (Usage), but specs 03/05 promise exit 3 (Runtime) for hook/template
failures.

- [x] Classify genuine runtime failures in `internal/strategy/atomic/` and
  `internal/hooks/` as `cmderr.Runtime` (exit 3); `emitError` now reports the resolved code
  in the JSON error event instead of a hardcoded 1
- [x] Test a failing hook (`allow_fail: false`) resolves to exit 3 — unit tests in the hooks
  runner plus a deployer test with a fake runner (local shell exec is forbidden by the test
  rules); template errors also resolve to 3
- [x] Specs 01/03/05 already documented exit 3 for hook/template/runtime failures; the fix
  aligns the code to them (no spec change needed)

### Documentation reconciliation

Standalone doc drift found during the review, not tied to a code change above.

- [x] [Spec 05](../specs/05-hooks.md) — remove the stale "interactive hooks not supported in
  v0" claim (the code supports them)
- [x] [`docs/specs/README.md`](../specs/README.md) — add Spec 07 (SSH Transport) and Spec 08
  (Multi-server) to the index
- [x] [Spec 03](../specs/03-commands.md) + `.claude/rules/coding.md` — correct the
  `--config` default to the real chain (`.config/bifrost.yml` → `.bifrost.yml`)
- [x] [Spec 07](../specs/07-ssh-transport.md) — align the `release list`/`activate`/`rollback`
  agent-invocation examples with the actual client commands
- [x] [`versions.md`](versions.md) — update the milestone range to include M18–M19
- [x] Qualify the `ADR-0018` reference in `.goreleaser.yml` as a forge ADR
- [x] Move working-doc artifacts out of `docs/`: the superpowers plan →
  `.claude/plans/m18-hook-lifecycle.md` and the design notes →
  `.claude/plans/m18-hook-lifecycle-design.md` (matching the existing `m0`/`m2` plans);
  the durable links in the M18 entry and ADR-0012 are repointed

Spec references: full code/doc review 2026-06-13; [Spec 03](../specs/03-commands.md),
[Spec 05](../specs/05-hooks.md), [Spec 07](../specs/07-ssh-transport.md),
[ADR-0011](../adr/0011-agent-binary-distribution.md).

Deliverable: `bifrost deploy` to a server without `--agent-binary` downloads, verifies, and
runs the matching agent; remote commands are injection-safe; hook output is visible over SSH;
runtime failures exit 3; CI runs the integration suite; specs match the code.

---

## M20 — Cleanup

Lower-priority simplifications and hardening from the 2026-06-13 review. Quality only, no
behaviour change expected (except where noted). TDD where behaviour changes; pure refactors
keep the existing tests green. Same two-step commit flow, one task at a time.

- [x] Deduplicate the flat-detect → merge → validate block shared by `deploy`, `release list`,
  `release activate`, `release rollback`; share `writeTempFlatConfig`, the env-map helper, and
  the interactive-confirm helper across the `cmd` and `release` packages (S1)
- [x] Collapse the 8 near-identical hook-stage blocks in
  `internal/strategy/atomic/deployer.go` into a single `runHookStage` helper (S2)
- [x] `internal/cmd/deploy.go` — pass `cmd.Context()` instead of `context.Background()` so
  SIGINT cancels the deploy (S3). Required forge v0.17.0, which wires the signal context via
  `fang.WithNotifySignal` (`cmd.Context()` was inert before); cancellation reaches `Extract`
  — the other deploy steps don't observe ctx, so fuller cancellation is future work
- [x] Pick one error-inspection idiom: `errors.AsType` (`ssh.go`) vs `errors.As`
  (`runner.go`) (S4)
- [x] `internal/config` — reject `strategy` values other than `atomic` in v0 (S5)
- [x] `internal/config` — reject an empty hook `cmd` at load time (S6)
- [x] `internal/tui/stream.go` — replace the 64 KB `bufio.Scanner` with a larger buffer or
  `bufio.Reader` so large event lines don't truncate the stream (S7)
- [x] Use the existing `SFTP.Chmod` instead of shelling out `chmod +x` for the agent (S8)
- [x] `internal/transport/agent.go` — give the download an `http.Client` with a timeout (S9)
- [x] `internal/strategy/atomic/deployer.go` — log `purgePlanErr` instead of silently
  discarding it (S10)
- [x] Extract the SSH stage-and-deploy scaffolding from the client E2E tests into
  `internal/testutil` (`DeployOverSSH`), then add a container E2E asserting hook output is
  visible over SSH (closes the M19 deferral; complements the unit + fake-runner routing tests) (S11)
- [x] Fix the known_hosts test-infra race S11 exposed on CI: `WriteKnownHosts` wrote to a
  shared `/tmp/bifrost-test-known-hosts` that the parallel CMD/RELEASE packages clobbered
  ("knownhosts: key is unknown"); it now writes a unique per-call path
  (`internal/testutil/ssh_container.go`)
- [x] Harden the `nocolor` integration tests against a testcontainers exec exit-code race
  under CI load: confirm the deploy errored via stderr (read reliably) rather than the racy
  container-exec exit code (`internal/cmd/nocolor_integration_test.go`)

Deliverable: duplication removed, deploy honours cancellation, config rejects invalid
strategy/empty hooks, and the remaining review nits are closed. No functional regressions.

---

## M21 — Deploy cancellation guard

M20 (S3) wired `cmd.Context()` into `deploy` so SIGINT cancels extraction (via
`ctx.Done()`, checked by `Extract`/`archives.Extract`). Every step after extraction
(`LinkShared`, `SetCurrent`, hooks, `Purge`) ignores `ctx` entirely and already runs to
completion regardless of cancellation — but silently, with no feedback to the user. Add a
one-time warning so a post-extraction Ctrl+C doesn't look like it was ignored.

- [x] `internal/tui/styles.go` — add a warning color/style
- [x] `internal/tui/deploy.go` — `PrintWarning(mode, out, msg)` for human/plain output
- [x] `internal/strategy/atomic/deployer.go` — after `Extract` succeeds, check `ctx.Err()`
  at the start of each remaining stage (`runHookStage`/`runStep`); on first detection,
  print `"Deploy in progress — cannot be cancelled, continuing..."` (human/plain) or emit
  a JSON `{"event":"warning",...}` exactly once, and continue the pipeline unchanged
- [x] Unit test: a context cancelled during a post-extraction hook still lets `Deploy`
  return `nil`, and the warning appears in the output exactly once
- [x] Integration test: send SIGINT mid-deploy after extraction completes, assert the
  deploy still finishes successfully and the warning is printed

Deliverable: Ctrl+C during extraction still cancels cleanly (M20 S3 behaviour preserved);
Ctrl+C after extraction prints a one-time warning and the deploy runs to completion.

---

## M22 — `--interactive` deploy flag

Add a `--interactive` flag to `deploy` for troubleshooting: pause after every numbered
pipeline step — the 7 base steps (config loaded, release dir created, artifact extracted,
shared dirs linked, shared files linked, current symlink updated, old releases purged)
plus any configured hook-group steps (`deployStepTotal`'s "+1 per configured hook group",
e.g. `[4/8] post_extract hooks`) — so the operator can manually inspect server state
before continuing. This is independent of the existing per-hook `interactive: true`
confirm (`internal/hooks/runner.go`), which is unaffected and unforced.

- [x] `internal/cmd/deploy.go` — add `--interactive` bool flag
- [x] Require human mode + a real TTY when `--interactive` is set; otherwise fail fast
  with a clear `cmderr.Usage` error (matches the non-TTY guard pattern used by
  `release activate`)
- [x] `internal/tui/confirm.go` — add a step-continue confirm prompt (huh, bifrost theme),
  e.g. `Title("Continue to next step?")`, default Yes
- [x] `internal/strategy/atomic/deployer.go` — after every numbered step (base steps and
  hook-group steps alike), if interactive mode is on, show the prompt; answering "no"
  aborts the deploy with `cmderr.Runtime` and a message naming the step
- [x] Unit test: a fake confirm function that returns `false` on the 2nd step aborts
  `Deploy` after step 2 and never runs step 3 onward
- [x] Unit test: with a configured hook group, the prompt also fires after that
  hook-group step
- [x] Unit test: `--interactive` without a TTY (or in `json`/`plain` mode) fails fast with
  `cmderr.Usage`, no steps run
- [x] Integration test: `--interactive` with a scripted confirm sequence runs a full
  deploy to completion (regression check — confirming "yes" at every step behaves like a
  normal deploy)
- [x] Update [Spec 03](../specs/03-commands.md) and [Spec 06](../specs/06-tui-ux.md) to
  document the flag and its prompts

Deliverable: `bifrost deploy --interactive` pauses for confirmation between each of the 7
pipeline steps on a TTY; declining aborts the deploy; the flag is rejected outside
human+TTY.

---

## M23 — Reject `--interactive` for remote deploys (review fix)

A code review found two defects in the M22 `--interactive` guard. First, the human+TTY
guard runs before config load, so on a real TTY a `--interactive` deploy against a config
with `servers:` passed the guard and silently fell through to a client-mode SSH deploy —
the flag was accepted and ignored, because the remote agent runs in `--output json` mode
where per-step prompts are skipped. Second, [Spec 03](../specs/03-commands.md) documented
the guard's usage error as exit 2, but the code returns `cmderr.Usage` (exit 1).

- [x] `internal/tui/progress.go` — make `IsTTY` overridable via `SetIsTTY` (test seam,
  matching the `SetStatFile`/`SetInitWrite` pattern)
- [x] `internal/cmd/deploy.go` — after merge, reject `--interactive` with `cmderr.Usage`
  when `len(merged.Servers) > 0`, before the client-mode branch
- [x] Unit test: with TTY forced on, `--interactive` against a server config returns a
  `cmderr.Usage` error naming "remote" and never attempts an SSH connection
- [x] [Spec 03](../specs/03-commands.md) — correct the guard exit code to 1 and document
  the local-only restriction

Deliverable: `bifrost deploy --interactive` against a server config fails fast with a clear
usage error (exit 1) instead of silently ignoring the flag; the spec matches the code.

---

## M24 — Validate archive symlink targets (review fix)

A code review found that `extractHandler` validated entry *names* against path
traversal (absolute / `../`) but created symlink entries with an unvalidated,
attacker-controlled *target* (`os.Symlink(fi.LinkTarget, …)`). A crafted artifact
could ship a symlink whose target is absolute (`/etc`) or climbs out via `..`;
a later entry written under that link then escapes the release directory (the
symlink-based zip-slip variant). Severity is bounded — operators deploy their own
artifacts — but it is inconsistent with the existing path-traversal guard, so the
extractor should reject escaping symlink targets too.

- [x] `internal/strategy/atomic/deploy.go` — add `symlinkEscapes(destDir, linkPath,
  target)` and reject symlink entries whose target is absolute or resolves outside
  the release directory, before `os.Symlink`
- [x] Unit test: an archive with an absolute symlink target is rejected and the link
  is not created
- [x] Unit test: an archive with a `..`-escaping relative symlink target is rejected
- [x] Unit test: a relative symlink that resolves inside the release directory is still
  extracted normally (no over-restriction)
- [x] [Spec 03](../specs/03-commands.md) — document artifact extraction safety: rejected
  entry names and symlink targets, why absolute targets are rejected unconditionally, and
  the `paths.shared` alternative

Deliverable: archive extraction rejects symlink targets that would escape the release
directory, matching the existing entry-name traversal guard.

---

## M25 — Honor `--dry-run` for remote and rollback commands (review fix)

A code review found that `--dry-run` is silently ignored in several paths. `deploy` and
`release activate` both check the client-mode branch (`len(merged.Servers) > 0`) before
ever checking `--dry-run`, so a remote (SSH) invocation performs a real deployment or
activation even when the operator asked for a simulation. `release rollback` never checks
`--dry-run` at all, on either the local or client path.

- [x] `internal/cmd/client_deploy.go` — `runClientDeploy` checks `--dry-run` before looping
  servers; prints a per-server plan and returns without connecting to any server when set
- [x] `internal/cmd/release/client_activate.go` — `runClientReleaseActivate` checks
  `--dry-run` after resolving release selections, before the mutating loop; prints a
  per-server plan and returns without executing the remote agent
- [x] `internal/cmd/release/client_rollback.go` — `runClientReleaseRollback` checks
  `--dry-run` before looping servers; prints a per-server plan (the actual target release
  is only known once the remote agent runs) and returns without connecting
- [x] `internal/cmd/release/rollback.go` — local rollback path checks `--dry-run` after
  computing the rollback target, before `LinkShared`/hooks/`SetCurrent`; prints the target
  release and returns without mutating anything
- [x] Unit tests: dry-run in each of the three client-mode functions never invokes the
  per-server mutating function (`deployToServerFn` / `releaseActivateToServerFn` /
  `releaseRollbackToServerFn`)
- [x] Integration test: local `release rollback --dry-run` leaves `current` unchanged and
  prints the would-be target release
- [x] [Spec 03](../specs/03-commands.md) — document `--dry-run` behavior for `deploy`,
  `release activate`, and `release rollback`, including that a remote (client-mode)
  dry run prints a plan without connecting to any server

Deliverable: `--dry-run` never performs a real deployment, activation, or rollback on any
path — local or remote — for any of `deploy`, `release activate`, `release rollback`.

---

## M26 — Fix crash, hang, and masked-failure bugs in the SSH event pipeline (review fix)

A code review found several ways a remote deploy can crash the client, hang indefinitely,
or misreport success/failure, independent of any specific command's business logic.

- [x] `internal/tui/stream.go` — guard against a decoded event map being `nil` (a literal
  `null` JSON line) before writing into it, instead of panicking with "assignment to entry
  in nil map"
- [x] `internal/tui/stream.go` — a single large `hook_output` line (a verbose hook's entire
  stdout) must not fail `ForwardStream` on an otherwise-successful deploy; raise or remove
  the `bufio.Scanner`'s fixed token-size cap
- [x] `internal/cmd/client_deploy.go` — when `ForwardStream` returns early, drain the
  remaining SSH channel output instead of only closing the pipe reader, so the SSH copy
  goroutine can exit and `wg.Wait()` cannot hang forever
- [x] `internal/cmd/client_deploy.go` — check `streamErr` before `execErr` (or ignore
  `io.ErrClosedPipe` from `execErr`) so a real remote failure and its exit code are not
  masked by a pipe-close race triggered by the client's own cleanup
- [x] Unit tests covering each of the above

Deliverable: a remote deploy with verbose hook output, an unusual event payload, or a
mid-stream failure never panics or hangs the client, and always reports the agent's real
exit code.

---

## M27 — Reject a flat config that declares `servers:` (review fix)

Per [ADR-0009](../adr/0009-self-agent-model.md), a flat config (no `environments:` key) is
always agent mode, and `MergeFlat` correctly drops `Servers` — the agent runs locally on
the target server and must not re-open SSH to itself. But nothing warns a user who
hand-authors a flat `.bifrost.yml` with a top-level `servers:` block, expecting a remote
deploy: it silently deploys to the local machine instead, no error.

- [x] `internal/config/loader.go` — reject a flat config (`IsFlat(cfg)`) that has a
  non-empty `cfg.Servers` map, at `Load` time (exit 2, matching other config validation)
- [x] Unit test covering the rejection
- [x] [Spec 02](../specs/02-configuration.md) — document that `servers:` has no effect
  without an `environments:` block, and is rejected at load time in that combination

Deliverable: a flat config that declares `servers:` fails fast with a clear config error
instead of silently deploying locally.

---

## M28 — Fix JSON-mode safety in `release activate`/`rollback` (review fix)

A code review found that `release activate` and `release rollback` never got the M19 fix
that routes hook output through JSON events instead of raw stdout writes, and that their
interactive prompts (hook confirms, the release selector) gate on TTY alone rather than on
output mode — both `deployer.go` and `deploy.go`'s `--interactive` guard already do this
correctly; these commands never picked up the pattern.

- [x] `internal/tui/confirm.go` — `InteractiveHookConfirm` takes a `forgeui.Mode` and
  returns nil (hook skipped with a warning) unless mode is human and stdout is a TTY
- [x] `internal/cmd/deploy.go` — pass `mode` to `InteractiveHookConfirm`
- [x] `internal/cmd/release/activate.go` — pass `mode` to `InteractiveHookConfirm`; require
  `--release` (not just a TTY) before showing the interactive selector when mode isn't
  human; swap the hook writer for `tui.NewHookOutputWriter` when mode is JSON
- [x] `internal/cmd/release/rollback.go` — same two fixes: mode-gated hook confirm, and
  `tui.NewHookOutputWriter` for hooks in JSON mode
- [x] `internal/cmd/release/client_activate.go` — require `--release` (not just a TTY)
  before showing the per-server selector form when mode isn't human; remote `agentCmd`
  includes `--output json`, matching `deploy`/`release list`/`release rollback`
- [x] `internal/cmd/release/client_rollback.go` — discovered while testing the above: once
  the agent emits hook output as JSON events, its stdout is NDJSON (multiple lines), not a
  single object, so the whole-buffer `json.Unmarshal` of the rollback result broke on any
  configured hook; replaced with `parseRollbackEvent`, which scans for the `"rollback"`
  event line
- [x] Tests for each

Deliverable: `release activate`/`rollback --output json` never renders a blocking prompt or
leaks raw hook output into the event stream, on any path.

---

## M29 — Extraction and agent-binary security hardening (review fix)

- [x] `internal/strategy/atomic/deploy.go` — reject an archive entry whose destination path
  has a symlink anywhere in its ancestor chain (not just a lexical check on the entry's own
  target), closing a two-link chain that defeats the M24 guard by writing through a
  previously-extracted symlink
- [x] `internal/transport/agent.go` — write the cached agent binary atomically (temp file +
  rename) and record its verified SHA-256 in a local sidecar file; on a cache hit, recompute
  the cached file's digest against the sidecar before trusting it (missing/mismatched
  sidecar falls back to a fresh download + verify)
- [x] `internal/transport/ssh.go` — close the SSH-agent unix socket dialed in `authMethods`
  when the `Client` is closed, instead of leaking one fd/connection per server per deploy
- [x] Unit tests: a two-link symlink-chain archive is rejected; the sidecar digest
  comparison used to verify a cache hit (a tampered/mismatched digest is rejected);
  `authMethods` returns the dialed agent conn for `Client.Close` to close. (The full
  download/cache disk-I/O path in `ResolveAgentBinary` remains untested at the unit level —
  this project's testing rules forbid local-host filesystem writes outside
  `internal/strategy/atomic/`, and it had no prior coverage either.)

Deliverable: extraction closes the ancestor-symlink escape class, the agent binary cache
never trusts unverified bytes, and SSH connections don't leak file descriptors.

---

## M30 — Exit-code and behavior consistency batch (review fix)

- [x] `internal/cmd/client_deploy.go` — forward `--init` to the remote agent command when
  set, so first-time remote deploys can create `releases_root`/`shared_root` on the target
- [x] `internal/cmd/client_deploy.go`, `internal/cmd/release/client_list.go` — replace
  `fmt.Sprintf("...: %v", err)`-into-`Message` with `%w`-wrapped errors passed through
  `cmderr.Wrap`, so `errors.Is`/`errors.As` work across the transport boundary and every
  failure resolves to the documented exit code (3) instead of defaulting to 1
- [x] `internal/cmd/release/list.go` — wrap `listReleases`' directory-read failure with
  `cmderr.Wrap(cmderr.Runtime, ...)` so `release list`/`rollback` exit 3 for a missing
  releases root, matching `deploy`'s classification of the same failure class
- [x] `internal/config/merge.go` — `MergeFlat` sorts hooks by priority like `Merge` does,
  instead of leaving them in file order
- [x] `internal/hooks/runner.go` — render `cmd_dir` through the same template pass as `cmd`,
  matching the documented example in spec 02 and the sample config
- [x] `internal/strategy/atomic/release.go`, `internal/cmd/release/list.go` — order releases
  by directory modification time instead of a lexicographic name sort, so a custom
  `--release-name` (documented as accepting e.g. `v2.1.3`) can't cause `Purge` to delete the
  newest release or `rollback` to miss a valid target. `client_release_integration_test.go`'s
  `seedReleases` helper now stamps each seeded release with a distinct, increasing mtime
  (a single bulk `mkdir` previously gave them effectively-tied timestamps)
- [x] [Spec 06](../specs/06-tui-ux.md) — correct the `--interactive` guard's documented exit
  code (1, not 2) and the JSON error event's documented `exit_code` (resolved via
  `cmderr.Resolve`, not a hardcoded 1), matching spec 03 and the M19 fix
- [x] Unit/integration tests for each

Deliverable: exit codes are consistent between local and client-mode paths for the same
failure class, remote `--init` works, flat-config hook ordering matches merged-config
ordering, `cmd_dir` templating works as documented, and release ordering is immune to
custom release names.

---

## M31 — Revert premature strategy interface (review fix)

A code review found that `internal/strategy.Deployer` is a strategy-selection interface
with exactly one implementation (`atomic.Deployer`), contradicting
`.claude/rules/coding.md`: "No strategy interface in v0 — it will be introduced in v1 when
a second strategy exists." We're past v0, but the precondition still isn't met — a second
strategy (docker/k8s) isn't scheduled until v4/v5 (`versions.md`). The interface's shape is
guessed from a single implementation and loses the concrete `*atomic.Deployer`'s chainable
`With*` methods behind the interface boundary for no current benefit.

- [x] `internal/strategy/strategy.go` — delete; move `DeployOptions` into
  `internal/strategy/atomic/deployer.go` as a concrete, unqualified type
- [x] `internal/strategy/atomic/deployer.go` — drop the `strategy.Deployer` interface
  satisfaction assertion and the `strategy` import; `Deploy` takes the local
  `DeployOptions` directly
- [x] `internal/cmd/deploy.go` — call `ad.Deploy(...)` on the concrete `*atomic.Deployer`
  directly, no interface variable
- [x] Update the 6 test files referencing `strategy.DeployOptions`/`strategy.Deployer` to
  use `atomic.DeployOptions` and drop the interface-satisfaction test
- [x] All existing tests pass unchanged in behavior (pure structural revert, no new
  behavior — verified by the existing suite, not a new test)

Deliverable: the interface is gone; re-introduce it in v4/v5 when a second strategy
actually exists, per the rule.

---

## M32 — Unify the duplicated agent-session staging sequence (review fix)

A code review found the connect+detect+resolve+stage+upload-agent+upload-config+chmod
sequence independently reimplemented three times — `client_list.go`'s
`stageAgentSession` (shared by `release list`/`activate`/`rollback`), `client_deploy.go`'s
`deployToServer`, and `testutil/deploy.go`'s `DeployOverSSH` — and the copies had already
drifted: `deployToServer` wrote its cleanup warning to stdout instead of stderr (corrupting
`--output json`), and `DeployOverSSH` still had the pre-M26 premature-pipe-close pattern
that can hang draining a large/erroring event stream. The per-server flat config upload
also round-tripped through a local temp file (`cmdutil.WriteTempFlatConfig`) instead of
`Staging.UploadBytes`, which `testutil/deploy.go` already used directly from memory.

- [ ] `internal/transport/session.go` (new) — `AgentSession` + `StageAgentSession`:
  connect, resolve/upload the agent binary, generate and upload the flat config via
  `config.GenerateFlatConfig` + `Staging.UploadBytes` (no local temp file), chmod; `Close`
  warns to stderr on a cleanup failure
- [ ] `internal/cmd/release/client_list.go` — delete `stageAgentSession`; `fetchReleaseEntries`
  uses `transport.StageAgentSession`, classifying its error with `cmderr.Wrap(cmderr.Runtime,
  ...)` once (every staging step gets the same classification, not just the first few)
- [ ] `internal/cmd/release/client_activate.go`, `client_rollback.go` — use
  `transport.StageAgentSession`/`AgentSession`
- [ ] `internal/cmd/client_deploy.go` — `deployToServer` uses `transport.StageAgentSession`,
  then uploads the artifact into the same session before building the agent command
- [ ] `internal/testutil/deploy.go` — `DeployOverSSH` uses `transport.StageAgentSession`,
  and its event-stream handling drains instead of closing early on a stream error (matching
  `execAndForward`'s M26 fix)
- [ ] `internal/cmd/cmdutil/config.go` — delete `WriteTempFlatConfig` once nothing calls it
- [ ] Existing unit/integration suites pass unchanged in behavior; the hang/warning-stream
  fixes land as natural consequences of deduplication, not separate changes to verify

Deliverable: one staging implementation instead of three, no local temp file for the flat
config, and the cleanup-warning-to-stdout and premature-pipe-close bugs can't recur in a
fourth copy because there's only one copy.

---

## M33 — Efficiency cleanup batch (review fix)

Independent, low-risk performance fixes identified by the code review, bundled into one
milestone (matching the M20 pattern).

- [ ] `internal/transport/sftp.go` — enable `sftp.UseConcurrentWrites` on the client so
  uploads pipeline writes instead of waiting for each 32 KB chunk's round-trip before
  sending the next
- [ ] `internal/strategy/atomic/deployer.go` — throttle extraction progress emission
  (JSON events and the human-mode progress bar) to roughly once per 1% of total bytes or
  50 ms, whichever comes first, instead of once per ~32 KB read chunk
- [ ] `internal/hooks/runner.go` — `runOne` now returns the rendered command string it
  executed; `RunWithEvents` reports that string in the event instead of rendering the
  template a second time (which also silently discarded its own render error)
- [ ] `internal/strategy/atomic/release.go`, `deployer.go` — new `PurgeNames` removes
  exactly a given name list with no scan/sort; the deployer passes `PurgePlan`'s
  already-computed candidates into it directly instead of letting `Purge` re-scan and
  re-sort the releases directory it was just computed from (falling back to `Purge`'s own
  scan only if `PurgePlan` itself failed, preserving that resilience)
- [ ] `internal/cmd/deploy.go` — hoisted the local `os.Stat(artifact)` existence check
  above the client-mode branch, so a bad `--artifact` path fails immediately instead of
  after a full SSH connect + agent stage
- [ ] `internal/tui/stream.go` — in JSON passthrough mode, `spliceServerField` inserts the
  `server` key into the raw NDJSON line instead of decoding into `map[string]any` and
  re-marshaling the whole object (which would also alphabetize every other key)
- [ ] Unit tests for each fix that doesn't require a real SSH/container round-trip,
  including one proving the splice preserves original key order (a re-marshal wouldn't);
  integration tests where the change is only observable end-to-end (artifact-check
  ordering, verified via the full `deploy` command against an unreachable server)

Deliverable: none of these change observable behavior — same output, same ordering,
same errors — only less redundant work per deploy.
