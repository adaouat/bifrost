# Spec 02 — Configuration

## File location

Bifrost resolves the config file path using the following priority chain (first match wins):

| Priority | Source | Notes |
|---|---|---|
| 1 | `--config <path>` CLI flag | Always wins when set |
| 2 | `BIFROST_FILE` environment variable | Useful in CI/CD; whitespace-only is treated as unset |
| 3 | `.config/bifrost.yml` | Used if the file exists on disk |
| 4 | `.bifrost.yml` | Default fallback |

If the resolved file does not exist, loading fails with an error.

### `config init` write destination

The `config init` command applies the same priority chain for the output path, except
step 3 checks whether the `.config/` **directory** exists (not the file, since the file is
being created). See [ADR-0008](../adr/0008-config-path-resolution.md) for rationale.

## Hierarchy and merge rules

Configuration is defined at three levels merged from lowest to highest priority:

```
Global (root of YAML) < Environment < Application
```

| Type | Merge behavior |
|---|---|
| Maps / objects | Deep merge — all keys kept, child value wins on conflict |
| Lists (hooks) | Concatenated across all levels, then sorted by `priority` ascending |
| Scalars | Child overrides parent |

After merging, `paths.releases_root` and `paths.shared_root` are required. If either
is missing the command fails with exit code 2 and a specific error message.

## Full schema

```yaml
# ── Strategy ───────────────────────────────────────────────────────────────────
strategy: atomic                 # Deployment strategy. Default: atomic. See ADR-0006.
                                 # v0 valid values: atomic
                                 # v4/v5: docker, k8s

# ── Servers (v1) ───────────────────────────────────────────────────────────────
# Named SSH target servers. Referenced by name from environments / applications.
# If no servers are referenced, Bifrost runs in local mode (v0 behaviour).
servers:
  <server_name>:
    host: 192.168.1.10           # REQUIRED. Hostname or IP address.
    port: 22                     # Default: 22
    user: deploy                 # REQUIRED. SSH username.
    key_file: ~/.ssh/id_rsa      # Optional. Private key path. ~ is expanded.
                                 # Falls back to SSH agent (SSH_AUTH_SOCK) if omitted.
    staging_dir: /tmp            # Optional. Base dir for the remote staging folder.
                                 # Full path: {staging_dir}/bifrost-{uuid}/

# ── Global defaults ────────────────────────────────────────────────────────────
paths:
  releases_root: /var/www/releases  # REQUIRED after merge
  shared_root: /var/www/shared      # REQUIRED after merge
  shared:
    directories: []                 # Relative paths symlinked into each release
    files: []                       # Relative paths symlinked into each release

settings:
  releases_to_keep: 10             # Number of old releases to retain (default: 10)

variables:                         # String key-value pairs available in hook templates
  key: value                       # Values must be strings (quote non-string YAML scalars)

hooks:
  post_extract: []        # Runs after extraction, before pre_link — raw release dir available
  pre_link: []            # Runs before shared resource linking
  pre_enable_release: []  # Runs before current symlink is updated
  post_enable_release: [] # Runs after current symlink is updated

# ── Environments ───────────────────────────────────────────────────────────────
environments:
  <env_name>:
    name: "Human-readable name"    # Optional
    servers: [<server_name>, ...]  # Optional (v1). Default for all apps in this env.
    applications:                  # REQUIRED — at least one application
      <app_name>:
        name: "Human-readable name"
        servers: [<server_name>]   # Optional (v1). Overrides env-level servers list.
        paths: { ... }             # Override / extend global + env paths
        settings: { ... }
        variables: { ... }
        hooks: { ... }
    paths: { ... }
    settings: { ... }
    variables: { ... }
    hooks: { ... }
```

## Hook entry schema

```yaml
cmd: "systemctl restart nginx"   # REQUIRED. Supports Go text/template syntax.
priority: 99999                  # Execution order, ascending. Default: 99999.
sudo: false                      # Wrap command in: sudo sh -c "<cmd>"
cmd_dir: ~                       # Working directory. Default: release directory.
allow_fail: false                # Continue deployment on non-zero exit code.
interactive: false               # Prompt user for confirmation before running.
```

## Hook lifecycle

| List | When it runs |
|---|---|
| `post_extract` | After artifact extraction, before `pre_link` — raw release dir available |
| `pre_link` | After `post_extract`, before shared resource linking |
| `pre_enable_release` | After shared linking, before `current` symlink update |
| `post_enable_release` | After `current` symlink update |

## Example config

```yaml
paths:
  releases_root: /var/www/webroot/ROOT
  shared_root: /var/nas/shared

settings:
  releases_to_keep: 30

variables:
  app_env: production

environments:
  prod:
    name: Production
    applications:
      web:
        name: Web Application
        paths:
          shared:
            directories:
              - var/log
              - var/cache
            files:
              - .env
              - config/secrets.yml
        hooks:
          post_extract:
            - cmd: "composer install --no-dev --optimize-autoloader"
              cmd_dir: "{{ .Directories.Working }}"
          post_enable_release:
            - cmd: "systemctl reload php-fpm"
              sudo: true
              priority: 10
            - cmd: "systemctl reload nginx"
              sudo: true
              priority: 20
```
