# Spec 01 — Overview

## What Bifrost does

Bifrost is a CLI tool for managing application deployments on a server. It implements
**atomic deployment**:

- Every deployment creates a new timestamped directory.
- A `current` symlink always points to the active release.
- Shared files and directories (logs, `.env`, uploads) live outside the releases and are
  symlinked into each release on deploy.
- Old releases are automatically pruned.

The activation step (flipping the `current` symlink) is an atomic OS operation — users
never see a partially deployed version. Rollback is instant: re-point `current` to any
previous release.

## Directory layout

```
{releases_root}/
├── 20260101-120000/          ← old release (purgeable)
├── 20260519-093012/          ← old release (purgeable)
├── 20260520-141500/          ← active release
│   ├── public/
│   ├── var/
│   │   └── log -> {shared_root}/var/log    ← shared dir symlink
│   └── .env -> {shared_root}/.env          ← shared file symlink
└── current -> 20260520-141500              ← active release symlink

{shared_root}/
├── var/
│   └── log/                  ← persisted across all releases
└── .env                      ← persisted across all releases
```

## Key concepts

**Release** — A directory named `YYYYMMDD-HHMMSS` (UTC) containing the extracted artifact.
Its name is lexicographically sortable, making "most recent" trivial to compute.

**current** — A symlink at `{releases_root}/current` pointing to the active release directory name (not an absolute path). Switching releases = updating this symlink.

**Shared resource** — A file or directory that must persist across releases (logs, secrets,
uploaded files). Stored in `{shared_root}` and symlinked into each release at deploy time.

**Hook** — A shell command that runs before or after a release is activated. Supports
template variable interpolation, priority ordering, sudo, and conditional allow-fail.

**Environment / Application** — Configuration is scoped: global settings are overridden
by environment settings, which are overridden by application settings. A single config
file can manage multiple environments and applications.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Usage error (bad flags or arguments) |
| 2 | Configuration error (invalid YAML, missing required fields) |
| 3 | Runtime error (artifact not found, directory missing, hook failed) |
| 70 | Internal software error (unexpected panic or unhandled condition) |
