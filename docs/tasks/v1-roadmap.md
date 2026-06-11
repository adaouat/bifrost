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
- [ ] Integration test: unknown remote arch → clean error message, exit code 3
- [ ] Integration test: agent exits non-zero mid-deploy → cleanup still runs
- [ ] Review and update specs 01, 02, 03, 07, 08 for any gaps found during implementation
- [ ] `bifrost.schema.json` — JSON Schema for `.bifrost.yml` (full v1 schema); update
  `config init` template to include `# yaml-language-server: $schema=` comment for IDE validation

Deliverable: CI passes on all v1 integration tests; no known edge cases unhandled.
