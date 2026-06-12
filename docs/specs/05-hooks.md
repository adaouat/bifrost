# Spec 05 — Hooks

Hooks are shell commands that run at defined points in the deployment lifecycle.

## Lifecycle points

The `deploy` command fires all 8 hooks below, in order. `release activate` and
`release rollback` only fire `pre_activate` / `post_activate`, around their existing
`current` symlink update.

```
pre_extract → [extract]      → post_extract
pre_link    → [link shared]  → post_link
pre_activate → [switch current] → post_activate
pre_purge   → [purge]        → post_purge
```

| Hook list | Shared symlinks present? | `current` updated? | Typical use |
|---|---|---|---|
| `pre_extract` | No | No | Prepare the empty release directory before extraction |
| `post_extract` | No | No | Compile assets, modify extracted files |
| `pre_link` | No | No | Remove dirs that will be replaced by symlinks, create placeholder subdirs |
| `post_link` | **Yes** | No | DB migrations, config cache — needs shared files, must run before traffic hits release |
| `pre_activate` | Yes | No | Final checks immediately before the symlink switch |
| `post_activate` | Yes | **Yes** | Reload services, notify monitoring |
| `pre_purge` | Yes | Yes | Snapshot or archive a release directory before it may be purged |
| `post_purge` | Yes | Yes | Cleanup or notification after old releases are removed |

`pre_link` and `post_link` differ in one critical way: shared dirs/files (`var/log`,
`.env`, etc.) are **not yet symlinked** at `pre_link` time, but are **fully in place**
by `post_link`. Run anything that requires shared config files (migrations, cache
warming) in `post_link`, not `pre_link`.

## Execution

Each hook runs as:

```sh
sh -c "<rendered_cmd>"
```

With `sudo: true`:

```sh
sudo sh -c "<rendered_cmd>"
```

Using `sh -c` (rather than splitting on spaces) means:
- Arguments with spaces work correctly.
- Shell operators work: `&&`, `||`, `|`, redirects, subshells.
- This fixes BUG-1 and BUG-2 from v1.

**Working directory:** defaults to the release directory. Override per-hook with `cmd_dir`.

**stdout/stderr:** always captured and displayed. On failure, both are shown regardless
of the `--output` mode.

## Ordering

All hooks from all config levels (global + environment + application) for a given
lifecycle point are merged into a single list, then sorted by `priority` ascending.
Hooks with the same priority run in the order they appear after merging (global first,
then environment, then application).

Default priority is `99999`, meaning hooks without an explicit priority run last.

## Template rendering

`cmd` fields are rendered with Go `text/template` before execution. Available data:

```go
type HookData struct {
    Settings    Settings            // Merged settings
    Variables   map[string]string   // Merged variables
    Directories Directories         // Path helpers
    Env         map[string]string   // OS environment variables
}

type Directories struct {
    Working  string   // Release directory being activated
    Current  string   // Path to the `current` symlink (not its target)
    Releases string   // Releases root directory
    Shared   string   // Shared root directory
}
```

Examples:

```yaml
cmd: "echo deploying {{ .Variables.app_env }}"
cmd: "ln -nfs {{ .Directories.Working }}/public /var/www/html"
cmd: "DB_HOST={{ .Env.DB_HOST }} php artisan migrate --force"
cmd: "echo 'keeping {{ .Settings.ReleasesToKeep }} releases'"
```

Template errors cause the deployment to fail immediately (exit code 3) with the
offending template and error displayed.

## Hook options

| Field | Type | Default | Description |
|---|---|---|---|
| `cmd` | string | — | REQUIRED. Shell command. Supports templates. |
| `priority` | int | 99999 | Execution order, ascending. |
| `sudo` | bool | false | Wrap with `sudo sh -c`. |
| `cmd_dir` | string | release dir | Working directory for the command. |
| `allow_fail` | bool | false | Continue deployment on non-zero exit. |
| `interactive` | bool | false | Prompt user for confirmation before running. Not supported in v0 — deployment errors if set to `true`. |

## Error handling

- If a hook exits non-zero and `allow_fail: false` (default): deployment stops, exit
  code 3. Hook stdout and stderr are shown.
- If `allow_fail: true`: a warning is logged, deployment continues.
- `interactive: true` prompts before each run. In non-interactive mode (`--output json`
  or `plain`, or no TTY): interactive hooks are skipped with a warning.
