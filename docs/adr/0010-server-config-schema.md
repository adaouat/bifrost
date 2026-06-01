# ADR-0010: Server configuration schema

**Status:** Accepted

## Context

v1 needs SSH connection details for each target server. Multiple environments and
applications may share the same physical server (e.g. `web-01` serves both `staging/api`
and `staging/worker`). Duplicating connection details per application would be error-prone
and hard to maintain.

## Decision

Introduce a top-level `servers:` map of named server entries. Environments and applications
reference server names in a `servers:` list. Server resolution follows app > env priority:
the app-level list overrides the env-level list when set.

```yaml
servers:
  web-01:
    host: 192.168.1.10     # required
    port: 22               # default: 22
    user: deploy           # required
    key_file: ~/.ssh/id_rsa  # optional; falls back to SSH agent (SSH_AUTH_SOCK)
    staging_dir: /tmp      # optional; default: /tmp. Full staging path: {staging_dir}/bifrost-{uuid}/

environments:
  prod:
    servers: [web-01, web-02]   # env-level default
    applications:
      api:
        servers: [web-01]       # overrides env-level
      worker:
        # inherits env-level: [web-01, web-02]
```

If neither env-level nor app-level `servers:` is set, Bifrost runs in local mode (v0
behaviour). This preserves full backwards compatibility.

## Authentication model

Bifrost never stores passwords in the config file. Authentication is attempted in this
order:

1. `key_file` — explicit private key path from server config
2. SSH agent — `SSH_AUTH_SOCK` socket if present
3. `BIFROST_SSH_PASSWORD` environment variable — last resort, applies to all servers

The config file containing only host, port, user, and key path contains nothing secret
and can be committed to version control. Key material stays on disk or in the agent.
Password via env var is intentionally inconvenient — it signals that key-based auth
is the expected path.

Known-hosts validation uses the system `~/.ssh/known_hosts` file. Bifrost does not
implement a `--insecure` flag; host key verification is always strict. Secrets management
(per-server passwords, Vault, AWS SSM) is out of scope for v1 — see [versions.md](../tasks/versions.md) v2.

## Validation rules

- Every name in a `servers:` reference list must exist as a key in the top-level
  `servers:` map. Missing references are reported at config load time (exit code 2).
- A server entry with no `host` or no `user` is invalid (exit code 2).
- `port` defaults to 22 if omitted.
- `staging_dir` defaults to `/tmp` if omitted.

## Rationale

**Top-level `servers:` map over inline server blocks:**
Connection details are infrastructure facts that change independently of deployment
config. Centralising them avoids duplication and makes a server rename a single-line
edit.

**Reference lists over a single-server field:**
Deploys to multiple servers are a first-class v1 concern. A list field models the intent
directly; a single-server field would require a breaking change in v2.

**App overrides env, not merges:**
Server lists are not merged (unlike hooks or maps) because a partial override — deploying
the API to only `web-01` while workers go to both — must be expressible without surprising
concatenation behaviour.

## Consequences

- The config schema (`internal/config/schema.go`) gains a top-level `ServerConfig` map
  and a `servers []string` field on both `EnvConfig` and `AppConfig`.
- The config validator checks that all referenced server names exist in the top-level map.
- Local mode (no servers) continues to work exactly as in v0 with no config changes.
- See [ADR-0009](0009-agent-model.md) for how servers are used at deploy time.
