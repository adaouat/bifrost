# Version Roadmap

High-level feature scope per major version. Implementation milestones within each version
are tracked separately (see `roadmap.md` for v0).

---

## v0 — Atomic deployment (current)

Like the current deployer, with fixes and enhancements. Runs **on the server** — CI/CD
handles the SSH transport (copy artifact, copy binary, call binary remotely).

**Deployment model:**
- Single binary, no SSH management
- Atomic strategy only: extract archive → link shared dirs/files → flip `current` symlink atomically → run hooks
- Single server per invocation

**Fixes over v1 deployer:**
- Hook commands run via `sh -c` — supports `&&`, `|`, redirects, args with spaces
  (v1 used `.split(' ')` which broke all of these)
- `sudo` + shell operators work correctly
- Extraction errors always surfaced (not hidden)

**Features:**
- `bifrost deploy` — full deploy flow
- `bifrost release list / activate / rollback` — release management
- `bifrost config` — config validation and display
- Three-level config merge: global < environment < application
- Hook system: pre/post lifecycle hooks with template variables, priority ordering, sudo,
  allow-fail, interactive confirm
- Three output modes: human (TUI), json (event stream), plain (CI-friendly)
- Dry-run mode
- `--output json` event stream for machine consumption

---

## v1 — SSH orchestration + agent model

Bifrost manages SSH itself. CI/CD no longer needs to handle transport.
v0 local mode is preserved: no `servers:` config = runs on the server as before.

**New deployment model:**
- Bifrost runs **client-side** (local machine or CI runner)
- Opens SSH connections to one or more target servers
- Detects remote OS/arch via `uname`, downloads the matching Bifrost binary from GitHub
  Releases (cached at `{os.UserCacheDir()}/bifrost/agents/{version}/{os}_{arch}/bifrost`)
- Uploads binary + artifact + flat merged config to `/tmp/bifrost-{uuid}/` via SFTP
- Executes itself as the agent in `--output json` mode
- Streams JSON events back and renders them locally
- Supports sequential deployment to multiple servers (parallel planned for v2+)

**Why self-agent over a dedicated agent binary:**
- Single binary to distribute; agent always version-matches client
- Existing `--output json` is the protocol — no new wire format
- All commands (`release list`, `activate`, `rollback`) automatically work over SSH

**New features:**
- Top-level `servers:` map in `.bifrost.yml`; env/app reference server names
- Auth: key file → SSH agent → `BIFROST_SSH_PASSWORD` env var
- `--agent-binary <path>` flag for air-gapped or pre-release builds
- `release activate` multi-server TUI: per-server release selector (huh form)
- `strategy.go` interface formalized — atomic strategy becomes the first typed implementation
- `release list` / `activate` / `rollback` work over SSH

**Implementation milestones:** see [`v1-roadmap.md`](v1-roadmap.md) (M11–M17).

**Known limitation carried forward:**
- Windows: `sh -c` in hooks requires Git Bash or WSL in PATH; symlinks require Developer
  Mode or elevated privileges. Builds provided on best-effort basis.

---

## v2 — Plugin system

Extend Bifrost's behavior without modifying the core binary.

**Plugin model:** out-of-process executables implementing a defined protocol
(stdin/stdout JSON or gRPC). This is the same model used by Terraform, Vault, and Packer —
avoids Go's native `plugin.Open` which requires identical build flags and Go version.

**Built-in plugin types (hook runners):**
- `shell` — current behavior, `sh -c`
- `http` — POST to a webhook URL with deploy context
- Custom executables — any binary in PATH implementing the protocol

**Extensible via plugins:**
- Custom hook runners
- Custom notification backends
- Custom secret backends (see secrets management below)

---

## v3 — Daemon mode + fleet management

Bifrost can run as a long-lived process on servers, exposing an API for a central
controller or web UI.

**Approach TBD at v3** — two options to evaluate then:
- `bifrost serve` — same binary, daemon mode flag
- `bifrostd` — separate binary with its own `cmd/bifrostd/` entry point

The decision will depend on how different the dependency graph is (HTTP server, no TUI,
persistent state). If the daemon needs significantly different dependencies, a separate
binary is cleaner.

**Features:**
- REST or gRPC API for deployment operations
- Web UI for fleet management (likely a separate repository)
- Deployment history and status
- Centralized config management

---

## v4/v5 — Additional strategies

New deployment strategies alongside the existing atomic strategy. Selected via the
`strategy:` config field.

**docker strategy:**
- SSH to server (uses v1 transport layer)
- `docker pull <image>`
- `docker compose up -d [services]`
- No artifact, no symlinks — just remote command execution
- Config: `compose_file`, `services`, image references

**k8s strategy:**
- `kubectl set image` or `helm upgrade`
- No SSH required — uses kubeconfig
- Config: cluster context, namespace, chart/manifest references

**Deployment strategies (blue/green, canary):**
- Blue/green: deploy to idle environment, switch traffic at load balancer level
- Canary: graduated traffic shift with metric gates
- Requires external system integration (nginx, HAProxy, AWS ALB, Kubernetes ingress)
- Exact scope and config schema TBD

---

## Unscheduled features

Features discussed but not yet assigned to a version. To be placed when scope is clearer.

| Feature | Notes |
|---|---|
| **Notifications** | Slack, webhook, PagerDuty on deploy success/failure. Likely v1 or v2 (via plugin system). |
| **Secrets management** | Vault, AWS SSM, 1Password integration. Likely v2 (plugin-based secret backend). |
| **Multi-app orchestration** | Deploy a stack of apps in dependency order. Likely v1 or v2. |
