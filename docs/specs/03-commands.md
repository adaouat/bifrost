# Spec 03 — Commands

## Global flags

Available on all commands:

| Flag | Default | Description |
|---|---|---|
| `--config` | `./.bifrost.yml` | Path to configuration file |
| `--output` | `human` | Output mode: `human`, `json`, `plain` |
| `--dry-run` | `false` | Show what would happen without making changes |
| `--verbose` | `false` | Enable verbose logging |

---

## `config`

Display and validate the effective configuration after merging.

```
bifrost config [--environment/--env <env>] [--application/--app <app>]
```

| Flag | Aliases | Required | Description |
|---|---|---|---|
| `--environment` | `--env` | No | Target environment key |
| `--application` | `--app` | No | Target application key |

| Behavior | Condition |
|---|---|
| Print full merged config as JSON | No `--env` or `--app` |
| Print merged app config + validate required fields | Both flags provided |

Exits non-zero (code 2) with a specific message for each missing required field.
With `--env` + `--app`, validates all environments/applications
in a single pass and reports all errors before exiting.

---

## `deploy`

Deploy an application from a compiled artifact archive.

```
bifrost deploy \
  --environment/--env <env> \
  --application/--app <app> \
  --artifact <path> \
  [--release-name <name>] \
  [--init]
```

| Flag | Aliases | Required | Description |
|---|---|---|---|
| `--environment` | `--env` | Yes | Target environment key |
| `--application` | `--app` | Yes | Target application key |
| `--artifact` | | Yes | Path to artifact file (`.tar.gz`, `.zip`, etc.) |
| `--release-name` | | No | Override auto-generated timestamp name (e.g. `v2.1.3`) |
| `--init` | | No | Create `releases_root` and `shared_root` if they do not exist |

**Deployment flow:**

1. Load and validate config → exit 2 on error
2. Verify `releases_root` and `shared_root` exist → create if `--init`, else exit 3
3. Verify artifact file exists and is readable → exit 3
4. Create release directory: `{releases_root}/{YYYYMMDD-HHMMSS}` (or `--release-name`)
5. Extract artifact into release directory (spinner + progress bar)
6. Run `post_extract` hooks
7. Run `pre_link` hooks
8. Link shared directories and files (see spec 04)
9. Run `pre_enable_release` hooks
10. Update `{releases_root}/current` symlink → new release directory
11. Run `post_enable_release` hooks
12. Purge old releases, keeping `settings.releases_to_keep` most recent

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

If `--release` is omitted and stdout is a TTY, presents an interactive list (huh select)
showing all releases with the current one highlighted. If not a TTY, exits with code 1
and an error asking for `--release`.

**Flow:**

1. Validate config
2. Verify release directory exists
3. If release is already `current`: prompt for confirmation (interactive) or exit 3 (non-interactive with `--no-confirm`)
4. Link shared directories and files
5. Run `pre_enable_release` hooks
6. Update `current` symlink
7. Run `post_enable_release` hooks

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

Equivalent to `release activate` with the second-most-recent release. Exits with code 3
if there is no previous release to roll back to.

---

## `init` (planned — M4)

First-time setup of a deployment target.

```
bifrost init --environment/--env <env> --application/--app <app>
```

| Flag | Aliases | Required | Description |
|---|---|---|---|
| `--environment` | `--env` | Yes | Target environment key |
| `--application` | `--app` | Yes | Target application key |

Creates `releases_root` and `shared_root` directories. Validates config. Does not deploy.
Equivalent to `deploy --init` without the artifact.
