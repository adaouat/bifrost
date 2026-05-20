# Spec 02 — Configuration

## File location

Default: `./.bifrost.yml` in the working directory.
Override with the global `--config <path>` flag.

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

After merging, `paths.roots.releases` and `paths.roots.shared` are required. If either
is missing the command fails with exit code 2 and a specific error message.

## Full schema

```yaml
# ── Global defaults ────────────────────────────────────────────────────────────
paths:
  roots:
    releases: /var/www/releases    # REQUIRED after merge
    shared: /var/www/shared        # REQUIRED after merge
  shared:
    directories: []                # Relative paths symlinked into each release
    files: []                      # Relative paths symlinked into each release

settings:
  releases_to_keep: 10             # Number of old releases to retain (default: 10)

variables:                         # Arbitrary key-value pairs available in hook templates
  key: value

hooks:
  pre_artifact: []                 # Runs after extraction, before shared linking
  pre_enable_release: []           # Runs before current symlink is updated
  post_enable_release: []          # Runs after current symlink is updated

# ── Environments ───────────────────────────────────────────────────────────────
environments:
  <env_name>:
    name: "Human-readable name"    # Optional
    applications:                  # REQUIRED — at least one application
      <app_name>:
        name: "Human-readable name"
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
| `pre_artifact` | After artifact extraction, before shared resource linking |
| `pre_enable_release` | After shared linking, before `current` symlink update |
| `post_enable_release` | After `current` symlink update |

## Example config

```yaml
paths:
  roots:
    releases: /var/www/webroot/ROOT
    shared: /var/nas/shared

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
          pre_artifact:
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
