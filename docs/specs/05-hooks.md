# Spec 05 — Hooks

Hooks are shell commands that run at defined points in the deployment lifecycle.

## Lifecycle points

| Hook list | When |
|---|---|
| `pre_artifact` | After extraction, before shared resource linking |
| `pre_enable_release` | After shared linking, before `current` symlink update |
| `post_enable_release` | After `current` symlink update |

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
| `interactive` | bool | false | Prompt user for confirmation before running. |

## Error handling

- If a hook exits non-zero and `allow_fail: false` (default): deployment stops, exit
  code 3. Hook stdout and stderr are shown.
- If `allow_fail: true`: a warning is logged, deployment continues.
- `interactive: true` prompts before each run. In non-interactive mode (`--output json`
  or `plain`, or no TTY): interactive hooks are skipped with a warning.
