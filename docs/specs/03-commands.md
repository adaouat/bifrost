# Spec 03 — Commands

## Global flags

Available on all commands:

| Flag | Default | Description |
|---|---|---|
| `--config` | `./.deployer.yml` | Path to configuration file |
| `--output` | `human` | Output mode: `human`, `json`, `plain` |
| `--dry-run` | `false` | Show what would happen without making changes |

---

## `config`

Display and validate the effective configuration after merging.

```
deployer config [--environment <env>] [--application <app>]
```

| Behavior | Condition |
|---|---|
| Print full merged config as JSON | No `--environment` or `--application` |
| Print merged app config + validate required fields | Both flags provided |

Exits non-zero (code 2) with a specific message for each missing required field.
With `--environment` + `--application`, validates all environments/applications
in a single pass and reports all errors before exiting.

---

## `artifact`

Deploy an application from a compiled artifact archive.

```
deployer artifact \
  --environment <env> \
  --application <app> \
  --artifact <path> \
  [--release-name <name>] \
  [--init]
```

| Flag | Required | Description |
|---|---|---|
| `--environment` | Yes | Target environment key |
| `--application` | Yes | Target application key |
| `--artifact` | Yes | Path to artifact file (`.tar.gz`, `.zip`, etc.) |
| `--release-name` | No | Override auto-generated timestamp name (e.g. `v2.1.3`) |
| `--init` | No (flag) | Create `releases_root` and `shared_root` if they do not exist |

**Deployment flow:**

1. Load and validate config → exit 2 on error
2. Verify `releases_root` and `shared_root` exist → create if `--init`, else exit 3
3. Verify artifact file exists and is readable → exit 3
4. Create release directory: `{releases_root}/{YYYYMMDD-HHMMSS}` (or `--release-name`)
5. Extract artifact into release directory (spinner + progress bar)
6. Run `pre_artifact` hooks
7. Link shared directories and files (see spec 04)
8. Run `pre_enable_release` hooks
9. Update `{releases_root}/current` symlink → new release directory
10. Run `post_enable_release` hooks
11. Purge old releases, keeping `settings.releases_to_keep` most recent

---

## `release list`

List all available releases for an application.

```
deployer release list --environment <env> --application <app>
```

Output: list of release directory names sorted newest-first, with the active release
marked. Excludes the `current` symlink itself.

In `--output json` mode: JSON array of objects with `name` and `active` fields.

---

## `release enable`

Activate a previously deployed release.

```
deployer release enable \
  --environment <env> \
  --application <app> \
  [--release <name>]
```

| Flag | Required | Description |
|---|---|---|
| `--environment` | Yes | Target environment key |
| `--application` | Yes | Target application key |
| `--release` | No | Release name to activate. If omitted, shows interactive selector. |

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
deployer release rollback --environment <env> --application <app>
```

Equivalent to `release enable` with the second-most-recent release. Exits with code 3
if there is no previous release to roll back to.

---

## `init` (planned)

First-time setup of a deployment target.

```
deployer init --environment <env> --application <app>
```

Creates `releases_root` and `shared_root` directories. Validates config. Does not deploy.
Equivalent to `artifact --init` without the artifact.
