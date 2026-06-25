# Spec 03 — Commands

## Global flags

Available on all commands:

| Flag | Default | Description |
|---|---|---|
| `--config` | `.config/bifrost.yml`, then `.bifrost.yml` | Config file path (full resolution chain in [Spec 02](02-configuration.md)) |
| `--output` | `human` | Output mode: `human`, `json`, `plain` |
| `--dry-run` | `false` | Show what would happen without making changes |
| `--verbose` | `false` | Enable verbose logging |

---

## `config`

Parent command for inspecting, validating, and scaffolding configuration. Running
`bifrost config` with no subcommand prints help listing the three subcommands below.

### `config show`

Display the effective configuration after merging, as JSON.

```
bifrost config show [--environment/--env <env>] [--application/--app <app>]
```

| Flag | Aliases | Required | Description |
|---|---|---|---|
| `--environment` | `--env` | No | Target environment key |
| `--application` | `--app` | No | Target application key |

With no flags, prints the full merged config. When `--env` and `--app` are given
(they must be used together), prints the resolved config for that target.

### `config check`

Validate required fields for a deployment target.

```
bifrost config check --environment/--env <env> --application/--app <app>
```

| Flag | Aliases | Required | Description |
|---|---|---|---|
| `--environment` | `--env` | Yes | Target environment key |
| `--application` | `--app` | Yes | Target application key |

Validates the merged config and reports every missing required field in a single pass.
Exits with code 2 if any field is missing.

### `config init`

Scaffold a default `.bifrost.yml` with inline comments covering every field.

```
bifrost config init [--force]
```

| Flag | Required | Description |
|---|---|---|
| `--force` | No | Overwrite an existing config file (otherwise init refuses) |

The write destination follows the config-path chain from [Spec 02](02-configuration.md):
`--config` → `BIFROST_FILE` → `.config/` directory → `.bifrost.yml`.

---

## `deploy`

Deploy an application from a compiled artifact archive.

```
bifrost deploy \
  --environment/--env <env> \
  --application/--app <app> \
  --artifact <path> \
  [--release-name <name>] \
  [--init] \
  [--agent-binary <path>]
```

| Flag | Aliases | Required | Description |
|---|---|---|---|
| `--environment` | `--env` | Yes | Target environment key |
| `--application` | `--app` | Yes | Target application key |
| `--artifact` | | Yes | Path to artifact file (`.tar.gz`, `.zip`, etc.) |
| `--release-name` | | No | Override auto-generated timestamp name (e.g. `v2.1.3`) |
| `--init` | | No | Create `releases_root` and `shared_root` if they do not exist |
| `--interactive` | | No | Pause for confirmation after each deploy step (human mode + TTY only) |
| `--agent-binary` | | No | Path to a prebuilt agent binary; skips download in client mode (v1) |

If the resolved config references one or more `servers`, `deploy` runs in **client mode**
over SSH instead of locally. See [Spec 07](07-ssh-transport.md) and
[Spec 08](08-multi-server.md) for the transport, agent, and multi-server flow. The
local flow below is what the agent runs on each target server.

**Deployment flow (local / per-server):**

1. Load and validate config → exit 2 on error
2. Verify `releases_root` and `shared_root` exist → create if `--init`, else exit 3
3. Verify artifact file exists and is readable → exit 3
4. Create release directory: `{releases_root}/{YYYYMMDD-HHMMSS}` (or `--release-name`)
5. Run `pre_extract` hooks
6. Extract artifact into release directory (spinner + progress bar)
7. Run `post_extract` hooks
8. Run `pre_link` hooks
9. Link shared directories and files (see spec 04)
10. Run `post_link` hooks
11. Run `pre_activate` hooks
12. Update `{releases_root}/current` symlink → new release directory
13. Run `post_activate` hooks
14. Run `pre_purge` hooks
15. Purge old releases, keeping `settings.releases_to_keep` most recent
16. Run `post_purge` hooks

**`--interactive` mode:**

When `--interactive` is set, the deploy pauses after every numbered step (all 7 base steps
plus any configured hook-group step) and shows a huh confirm prompt:

```
  Continue to next step?
  Artifact extracted
  > Yes / No
```

Answering **No** aborts the deploy immediately with exit code 3. The prompt is only
available in `human` mode on a real TTY; passing `--interactive` with `--output json`,
`--output plain`, or when stdout is not a TTY is rejected with a usage error (exit 1)
before any steps run.

`--interactive` is local-only: it is rejected with a usage error (exit 1) when the resolved
config references one or more `servers`, because in client mode the agent runs remotely in
`--output json` mode and cannot show per-step prompts.

---

## `release list`

List all available releases for an application.

```
bifrost release list --environment/--env <env> --application/--app <app>
```

| Flag | Aliases | Required | Description |
|---|---|---|---|
| `--environment` | `--env` | Yes | Target environment key |
| `--application` | `--app` | Yes | Target application key |
| `--agent-binary` | | No | Path to a prebuilt agent binary; skips download in client mode (v1) |

Output: list of release directory names sorted newest-first. The active release is shown
with a `← current` suffix. Excludes the `current` symlink itself.

In `--output json` mode: JSON array of objects with `name` and `active` fields.

---

## `release activate`

Activate a previously deployed release.

```
bifrost release activate \
  --environment/--env <env> \
  --application/--app <app> \
  [--release <name>]
```

| Flag | Aliases | Required | Description |
|---|---|---|---|
| `--environment` | `--env` | Yes | Target environment key |
| `--application` | `--app` | Yes | Target application key |
| `--release` | | No | Release name to activate. If omitted, shows interactive selector. |
| `--agent-binary` | | No | Path to a prebuilt agent binary; skips download in client mode (v1) |

If `--release` is omitted and stdout is a TTY, presents an interactive list (huh select)
showing all releases with the current one highlighted. If not a TTY, exits with code 1
and an error asking for `--release`.

**Flow:**

1. Validate config
2. Verify release directory exists
3. If release is already `current`: prompt for confirmation (interactive) or exit 3 (non-interactive with `--no-confirm`)
4. Link shared directories and files
5. Run `pre_activate` hooks
6. Update `current` symlink
7. Run `post_activate` hooks

---

## `release rollback`

Activate the release immediately preceding the current one.

```
bifrost release rollback --environment/--env <env> --application/--app <app>
```

| Flag | Aliases | Required | Description |
|---|---|---|---|
| `--environment` | `--env` | Yes | Target environment key |
| `--application` | `--app` | Yes | Target application key |
| `--agent-binary` | | No | Path to a prebuilt agent binary; skips download in client mode (v1) |

Equivalent to `release activate` with the second-most-recent release. Exits with code 3
if there is no previous release to roll back to.

---

## `release init`

First-time setup of a deployment target. (The top-level `init` command from M4 was
removed in M5 in favour of this subcommand under the `release` group.)

```
bifrost release init --environment/--env <env> --application/--app <app>
```

| Flag | Aliases | Required | Description |
|---|---|---|---|
| `--environment` | `--env` | Yes | Target environment key |
| `--application` | `--app` | Yes | Target application key |

Creates `releases_root` and `shared_root` directories. Validates config. Does not deploy.
Equivalent to `deploy --init` without the artifact. This command is local-only and takes
no `--agent-binary` flag.
