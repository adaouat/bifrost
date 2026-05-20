# ADR-0004: Use Go text/template for hook command templates

**Status:** Accepted

## Context

Hook `cmd` fields support variable interpolation so deployments can reference dynamic
values (release path, shared path, environment variables, user-defined variables).

Deployer v1 used **Liquid templates** via the `liquify` Dart package with syntax like:

```yaml
cmd: "ln -s {{ directories.working }}/public /var/www/html"
cmd: "DB_HOST={{ envVars.DB_HOST }} php artisan migrate"
```

## Decision

Use Go's stdlib **`text/template`** package.

## Rationale

- **No extra dependency.** `text/template` is in the Go standard library.
- **Well-documented and stable.** It will not break between Go versions.
- **Dot-notation on structs** is idiomatic Go and works out of the box.
- **Security:** `text/template` does not auto-escape (unlike `html/template`), which is
  correct for shell command generation where we control the data.

The syntax change from v1 Liquid to Go templates is minor:

| v1 (Liquid) | v2 (Go template) |
|---|---|
| `{{ variables.key }}` | `{{ .Variables.key }}` |
| `{{ settings.releases_to_keep }}` | `{{ .Settings.ReleasesToKeep }}` |
| `{{ directories.working }}` | `{{ .Directories.Working }}` |
| `{{ envVars.DB_HOST }}` | `{{ .Env.DB_HOST }}` |

The leading dot and capitalized field names are the only differences. This is a **breaking
change** for existing `.bifrost.yml` config files from v1.

### Why not keep Liquid

- No maintained Go Liquid implementation with the same feature set.
- `osteele/liquid` exists but adds a dependency for minimal gain.
- Go templates are more flexible (custom functions, pipelines) if needed later.

### Why not `text/template` with a shim for v1 syntax

- Extra complexity for a migration aid.
- v2 is a clean rewrite; a migration guide in the docs is sufficient.

## Consequences

- **Breaking change:** existing v1 config files need template variables updated.
  A migration note will be added to the README.
- Template errors surface as deployment errors with the offending template and error shown.
- Custom template functions can be added later via `template.FuncMap` without changing the API.
