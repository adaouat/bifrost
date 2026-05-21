# ADR-0006: Strategy-based deployment architecture

**Status:** Accepted

## Context

Bifrost v0 deploys applications by extracting an archive into a timestamped release
directory, linking shared resources, and flipping a `current` symlink (atomic deployment).
Without a strategy abstraction, this flow is hardcoded.

In practice, different projects use fundamentally different deployment models:

| Strategy | What happens |
|---|---|
| **atomic** | Extract archive, manage releases, flip `current` symlink atomically |
| **docker** | SSH to server, `docker pull`, restart compose services |
| **k8s** | `kubectl set image` or `helm upgrade` |

These are not variations on the same flow — they require different steps, different config
sections, and different success criteria. A hardcoded flow cannot accommodate them.

Both docker and k8s deployments are frequent real-world use cases that need to be supported.

## Decision

Introduce a top-level `strategy:` field in the config schema. Each strategy is a named
deployment model with its own lifecycle stages, config section, and implementation package
under `internal/strategy/<name>/`.

Hooks (`internal/hooks/`) are **shared across all strategies** — they attach to lifecycle
stage names defined by each strategy.

**v0 implements only the `atomic` strategy.** Docker and k8s strategies are planned for
v4/v5. The strategy interface will be formalized when a second strategy is added — for v0,
no interface abstraction is needed.

## Package layout

```
internal/
  strategy/
    atomic/     ← v0: extract archive, link shared, flip symlink atomically, purge releases
    docker/     ← v4/v5: pull image, restart compose services
    k8s/        ← v4/v5: kubectl / helm
  hooks/        ← shared: sh -c execution, sudo, template rendering, priority sort
```

## Config

`strategy:` is a top-level field defaulting to `atomic` when omitted, preserving
backward compatibility. Strategy-specific fields stay at the top level for v0; they
may migrate under a strategy key when a second strategy is added.

```yaml
strategy: atomic     # default; can be omitted in v0

paths:
  roots:
    releases: /var/www/releases
    shared: /var/www/shared
  shared:
    directories: []
    files: []
settings:
  releases_to_keep: 10
hooks:
  pre_artifact: []
  pre_enable_release: []
  post_enable_release: []
```

## Rationale

- **Strategies are orthogonal to transport.** The atomic strategy runs locally in v0 and
  remotely (via SSH agent) in v1. The strategy package has no knowledge of transport.
- **Hooks are strategy-agnostic by design.** A post-deploy hook that restarts nginx works
  the same whether the strategy is atomic or docker.
- **The package boundary makes adding strategies additive,** not disruptive to existing code.
- **No premature abstraction.** The strategy interface is deferred until v1 introduces a
  second strategy implementation. YAGNI for v0.

## Consequences

- `internal/deploy/` is removed in favour of `internal/strategy/atomic/` and `internal/hooks/`.
- The config schema gains a `strategy:` field (default: `atomic`).
- Every future strategy must define its lifecycle stage names so hooks can attach to them.
- The strategy interface (`internal/strategy/strategy.go`) will be introduced in v1.
