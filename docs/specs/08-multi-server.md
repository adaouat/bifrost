# Spec 08 — Multi-server commands

## Overview

When a resolved (env, app) config references multiple servers, Bifrost runs the requested
command on each server in sequence. This spec covers the per-command behaviour, output
rendering, and the multi-server TUI for release management.

For SSH transport details (auth, staging, agent invocation) see
[Spec 07 — SSH Transport](07-ssh-transport.md).

## Execution model (v1)

Servers are processed **sequentially** in the order they appear in the resolved `servers`
list. If a server fails, the remaining servers are skipped and Bifrost exits with the
failing server's exit code.

The code is designed so that parallel execution can be added in a later version without
restructuring the per-server logic. Each server's work is self-contained; only the outer
loop needs to change.

## `deploy` — multiple servers

Before the server loop:
1. Generate the release name once (timestamp or `--release-name` value).
2. Resolve the agent binary for each unique remote arch (may be the same for all servers).

Server loop (sequential):
```
for each server:
  1. Upload staging dir (agent binary, flat config, artifact)
  2. Execute agent deploy with --release-name <generated-name>
  3. Stream output (see Output rendering below)
  4. Cleanup staging dir
  5. On failure: skip remaining servers, exit with agent's exit code
```

Using the same release name on every server ensures all servers end up at the same
directory name, making cross-server `release rollback` meaningful.

## `release list` — multiple servers

For each server, run `release list --output json` via the agent and render the result.
Servers are processed sequentially. Each server's output is shown in a separate section.

**Human / plain mode — per-server section:**

```
Releases for prod › web  ─  web-01 (192.168.1.10)
  20260601-120000  ← current
  20260531-090000
  20260530-143012

Releases for prod › web  ─  web-02 (192.168.1.11)
  20260601-120000  ← current
  20260531-090000
```

**JSON mode** — each event gets a `"server"` field:

```json
{"event":"list","server":"web-01","releases":[...]}
{"event":"list","server":"web-02","releases":[...]}
```

A tab-based view (one tab per server, navigable with keyboard) is a planned enhancement
for a future version. The current sequential per-section layout is intentional for v1.

## `release activate` — multiple servers

### Selection TUI (client-side)

Before running any SSH command, the client queries all servers sequentially for their
release lists. It then presents a local `huh` form with one `Select` field per server:

```
┌─ Activate Release — prod › web ─────────────────────────────┐
│                                                              │
│  web-01 (192.168.1.10)                                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ > 20260601-120000 (current)                          │   │
│  │   20260531-090000                                    │   │
│  │   20260530-143012                                    │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  web-02 (192.168.1.11)                                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ > 20260601-120000 (current)                          │   │
│  │   20260531-090000                                    │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│                     [ Cancel ]  [ Activate ]                │
└──────────────────────────────────────────────────────────────┘
```

The user may select a different release per server (e.g. to stage a staggered rollback).
After confirming, the client runs `release activate <name> --output json` on each server
sequentially with the server's selected release name.

### Non-interactive path

When `--release <name>` is provided, the selection form is skipped. The given name is
used for all servers. If a server does not have that release, the agent exits non-zero
and the remaining servers are skipped.

When stdout is not a TTY and `--release` is not set, Bifrost exits with code 1.

### JSON mode

```json
{"event":"activate","server":"web-01","release":"20260531-090000","status":"done"}
{"event":"activate","server":"web-02","release":"20260601-120000","status":"done"}
```

## `release rollback` — multiple servers

Runs `release rollback --output json` on each server sequentially. No selection TUI —
rollback always activates the release immediately preceding the current one on each server.

If servers are at different current releases (common after a failed mid-fleet deploy),
each server rolls back to its own previous release independently.

**Human / plain mode:**

```
Rolling back prod › web

  web-01  ✔  Rolled back to 20260531-090000
  web-02  ✔  Rolled back to 20260530-143012
```

**JSON mode:**

```json
{"event":"rollback","server":"web-01","release":"20260531-090000","status":"done"}
{"event":"rollback","server":"web-02","release":"20260530-143012","status":"done"}
```

## Output rendering — general rules

### Human / plain mode (multi-server)

Each server's output block is preceded by a server header line:

```
── web-01 (192.168.1.10) ──────────────────────────────────────
  ✔ Config loaded
  ✔ Release directory created  (20260601-120000)
  ...

── web-02 (192.168.1.11) ──────────────────────────────────────
  ✔ Config loaded
  ...
```

### JSON mode (multi-server)

Every JSON event emitted by the agent is re-emitted by the client with an added
`"server"` field (server name as defined in the config):

```json
{"event":"start","step":"extract","server":"web-01"}
{"event":"done","step":"extract","server":"web-01","duration_ms":1240}
{"event":"start","step":"extract","server":"web-02"}
```

Single-server deploys also include the `"server"` field in JSON mode for consistency.

## Single-server behaviour

When only one server is configured, all multi-server rendering rules apply (server
header, `"server"` field in JSON). This keeps the output format consistent regardless
of fleet size and avoids special-casing single-server configs.

## Future: parallel execution + tabbed TUI

The sequential model is the baseline for v1. The following are explicitly out of scope
but should be kept in mind when designing the internal loop:

- **Parallel execution**: deploy to all servers simultaneously; aggregate results.
- **Tabbed TUI**: one tab per server in human mode, switchable with arrow keys or number
  keys (charmbracelet `bubbles/tabs`). Each tab shows the live output for that server.
- **Partial-failure strategy**: configurable behaviour when one server fails mid-fleet
  (stop all, continue, rollback completed).

These will be addressed in a future milestone once sequential is stable.
