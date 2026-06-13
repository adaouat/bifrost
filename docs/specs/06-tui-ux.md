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
  ┌──────────────────────────────────────┐
  │  Environment   prod › web            │
  │  Release       20260520-141500       │
  └──────────────────────────────────────┘

  ✔ Config loaded and validated
  ✔ Release directory created
      - Root path: /var/www/releases
      - Shared path: /var/nas/shared
  ✔ pre_extract hooks               (1/1)
  Extracting artifact   ████████░░░░░░░  52%
  ✔ Artifact extracted              (1.2s)
  ✔ post_extract hooks              (1/1)
  ✔ pre_link hooks                  (1/1)
  ✔ Shared directories linked       (2)
      - var/log
      - var/cache
  ✔ Shared files linked             (1)
      - .env
  ✔ post_link hooks                 (1/1)
  ✔ pre_activate hooks              (2/2)
  ✔ current symlink updated
      - /var/www/releases/20260520-141500
  ✔ post_activate hooks             (1/1)
  ✔ pre_purge hooks                 (1/1)
  ⠸ Purging 1 old release...
  ✔ Old releases purged             (kept 10)
      - 20260101-120000 deleted
  ✔ post_purge hooks                (1/1)

  Deployed in 4.3s  →  20260520-141500
```

Hook step lines (`pre_extract hooks`, `post_extract hooks`, `pre_link hooks`,
`post_link hooks`, `pre_activate hooks`, `post_activate hooks`, `pre_purge hooks`,
`post_purge hooks`) are only shown when at least one hook is configured for that
lifecycle. Spinner for purge only shown when there are releases to remove.

## Human mode — interactive release selection

When `release activate` is called without `--release`:

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
  $ bifrost deploy --environment prod --application web \
      --artifact ./release.tar.gz --dry-run

  DRY RUN — no changes will be made

  Would create   /var/www/releases/20260520-141500/
  Would run      [pre_extract]         echo "starting deploy"
  Would extract  ./release.tar.gz  →  /var/www/releases/20260520-141500/
  Would run      [post_extract]        composer install
  Would run      [pre_link]            cp .env.example .env
  Would link     .../20260520-141500/var/log  →  /var/nas/shared/var/log
  Would link     .../20260520-141500/.env     →  /var/nas/shared/.env
  Would run      [post_link]           php artisan config:cache
  Would run      [pre_activate]        echo "deploying"
  Would update   /var/www/releases/current  →  20260520-141500
  Would run      [post_activate]       systemctl restart nginx  (sudo)
  Would run      [pre_purge]           echo "cleaning up old releases"
  Would purge    20250101-120000, 20250515-083045  (keeping 10)
  Would run      [post_purge]          echo "cleanup done"
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
{"event":"hook","lifecycle":"pre_extract","index":0,"cmd":"echo starting deploy","exit_code":0}
{"event":"start","step":"extract","artifact":"/tmp/release.tar.gz"}
{"event":"progress","step":"extract","bytes":524288,"total":1048576}
{"event":"done","step":"extract","duration_ms":1234}
{"event":"hook","lifecycle":"post_extract","index":0,"cmd":"composer install","exit_code":0}
{"event":"hook","lifecycle":"pre_link","index":0,"cmd":"cp .env.example .env","exit_code":0}
{"event":"start","step":"link"}
{"event":"done","step":"link","duration_ms":45,"dirs":["var/log","var/cache"],"files":[".env"]}
{"event":"hook","lifecycle":"post_link","index":0,"cmd":"php artisan config:cache","exit_code":0}
{"event":"hook","lifecycle":"pre_activate","index":0,"cmd":"echo deploying","exit_code":0}
{"event":"start","step":"current_symlink"}
{"event":"done","step":"current_symlink","duration_ms":2,"path":"/var/www/releases/20260520-141500"}
{"event":"hook","lifecycle":"post_activate","index":0,"cmd":"systemctl restart nginx","exit_code":0}
{"event":"hook","lifecycle":"pre_purge","index":0,"cmd":"echo cleaning up old releases","exit_code":0}
{"event":"start","step":"purge"}
{"event":"done","step":"purge","duration_ms":123,"purged":["20260101-120000"],"kept":10}
{"event":"hook","lifecycle":"post_purge","index":0,"cmd":"echo cleanup done","exit_code":0}
{"event":"done","step":"deploy","release":"20260520-141500","duration_ms":4312}
```

On error:

```json
{"event":"error","step":"extract","message":"extracting archive: ...","exit_code":1}
```

Hook lifecycles emitted as `hook` events: `pre_extract`, `post_extract`, `pre_link`,
`post_link`, `pre_activate`, `post_activate`, `pre_purge`, `post_purge`.

## release list output

**Human:**

```
  Releases for prod › web  (3 total)

    20260520-141500  ← current
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
