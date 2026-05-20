# Spec 06 — TUI & UX

## Output modes

Controlled by the global `--output` flag.

| Mode | Description |
|---|---|
| `human` (default) | Colored, formatted output. Spinners and progress bar on real TTY. |
| `json` | One JSON object per event written to stdout. No TUI components. |
| `plain` | No colors, no spinners. Suitable for CI log capture. |

In `human` mode, if stdout is not a TTY (e.g. piped), spinners and interactive forms are
disabled automatically. Output is still colored unless `NO_COLOR` is set.

## Human mode — deploy output

```
  ┌─ Deployer ──────────────────────────────────────────────┐
  │  Environment   prod › web                               │
  │  Release       20260520-141500                          │
  └─────────────────────────────────────────────────────────┘

  ✔ Config loaded and validated
  ✔ Release directory created
  ⠸ Extracting artifact...          ████████░░░░░░░  52%
  ✔ Artifact extracted              (1.2s)
  ✔ pre_artifact hooks              (1/1)
  ✔ Shared directories linked       (3)
  ✔ Shared files linked             (1)
  ✔ pre_enable_release hooks        (2/2)
  ✔ current symlink updated
  ✔ post_enable_release hooks       (1/1)
  ✔ Old releases purged             (kept 10)

  Deployed in 4.3s  →  20260520-141500
```

## Human mode — interactive release selection

When `release enable` is called without `--release`:

```
  Which release would you like to enable?

  > 20260520-141500  (current)
    20260519-093012
    20260517-160041
    20260515-083045

```

When the selected release is already current:

```
  ⚠  20260520-141500 is already the active release. Enable anyway? (y/N)
```

## Human mode — dry run

```
  $ bifrost artifact --environment prod --application web \
      --artifact ./release.tar.gz --dry-run

  DRY RUN — no changes will be made

  Would create   /var/www/releases/20260520-141500/
  Would extract  ./release.tar.gz  →  /var/www/releases/20260520-141500/
  Would run      [pre_artifact]           composer install
  Would link     .../20260520-141500/var/log  →  /var/nas/shared/var/log
  Would link     .../20260520-141500/.env     →  /var/nas/shared/.env
  Would run      [pre_enable]             echo "deploying"
  Would update   /var/www/releases/current  →  20260520-141500
  Would run      [post_enable]            systemctl restart nginx  (sudo)
  Would purge    20250101-120000, 20250515-083045  (keeping 10)
```

## Human mode — styled errors (fang)

```
  Error  configuration error

  paths.roots.releases is required but not set for prod › web.
  Check your .bifrost.yml and ensure the application or its
  environment defines paths.roots.releases.
```

## JSON mode event stream

Each step emits one JSON line to stdout:

```json
{"event":"start","step":"extract","artifact":"/tmp/release.tar.gz"}
{"event":"progress","step":"extract","bytes":524288,"total":1048576}
{"event":"done","step":"extract","duration_ms":1234}
{"event":"hook","lifecycle":"pre_enable","index":0,"cmd":"echo hello","exit_code":0}
{"event":"done","step":"deploy","release":"20260520-141500","duration_ms":4312}
```

On error:

```json
{"event":"error","step":"extract","message":"...","exit_code":3}
```

## release list output

**Human:**

```
  Releases for prod › web  (3 total)

  * 20260520-141500  ← current
    20260519-093012
    20260517-160041
```

**JSON:**

```json
[
  {"name":"20260520-141500","active":true},
  {"name":"20260519-093012","active":false},
  {"name":"20260517-160041","active":false}
]
```
