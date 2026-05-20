# Version Roadmap

High-level feature scope per major version. Implementation milestones within each version
are tracked separately (see `roadmap.md` for v0).

---

## v0 — Capistrano-style artifact deployment (current)

Like the current deployer, with fixes and enhancements. Runs **on the server** — CI/CD
handles the SSH transport (copy artifact, copy binary, call binary remotely).

**Deployment model:**
- Single binary, no SSH management
- Artifact strategy only: extract archive → link shared dirs/files → flip `current` symlink → run hooks
- Single server per invocation

**Fixes over v1 deployer:**
- Hook commands run via `sh -c` — supports `&&`, `|`, redirects, args with spaces
  (v1 used `.split(' ')` which broke all of these)
- `sudo` + shell operators work correctly
- Extraction errors always surfaced (not hidden)

**Features:**
- `bifrost artifact` — full deploy flow
- `bifrost release list / enable / rollback` — release management
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

**New deployment model:**
- Bifrost runs **client-side** (local machine or CI runner)
- Opens SSH connections to one or more target servers
- Sends the artifact via SFTP
- Copies itself to the remote server (`/tmp/bifrost-agent`)
- Executes the agent remotely in `--output json` mode
- Streams output back and renders it locally in human mode
- Supports parallel deployment to multiple servers

**Why agent model over pure SSH orchestration:**
- Hooks run with full server-local context (env vars, filesystem, installed tools)
- Single SSH connection setup, then local operations are fast
- The agent binary is architecture-matched automatically at deploy time

**New features:**
- SSH connection config in `.bifrost.yml` (host, port, user, key, known_hosts)
- Multi-server deployment (parallel or sequential, configurable)
- `strategy.go` interface formalized — the artifact strategy becomes the first
  implementation of a typed interface
- Deployment result aggregation across servers

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

New deployment strategies alongside the existing artifact strategy. Selected via the
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
