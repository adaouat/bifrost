# Spec 04 — Shared Resources

Shared resources are files and directories that must persist across deployments.
They are stored in `{shared_root}` and symlinked into each release at deploy time.

Configured under `paths.shared` in the YAML:

```yaml
paths:
  shared:
    directories:
      - var/log
      - var/cache
      - public/uploads
    files:
      - .env
      - config/secrets.yml
```

## Linking algorithm

Run once for each entry in `directories` and `files`, in order.

Let `target = {shared_root}/{relative_path}` and `link = {release_dir}/{relative_path}`.

```
if target does not exist:
    if link exists (release shipped its own copy):
        move link → target          # seed shared from first release
    else:
        if entry is a directory: mkdir -p target
        if entry is a file:      touch target  (create empty)

delete link (or the directory subtree at link path)
mkdir -p parent(link)               # ensure parent directories exist
create symlink: link → target       # absolute path
```

## Behavior notes

- **First deploy:** if the artifact contains content at the shared path (e.g. a default
  `.env.example` shipped as `.env`), it is moved to `{shared_root}` to seed the shared
  copy. Subsequent deploys always use the shared copy.

- **Directory vs file:** the algorithm is the same. For directories, the whole tree is
  moved/created. For files, a single file is moved/created.

- **Parent directories:** `mkdir -p` is used before creating the symlink. If a shared
  path is `a/b/c`, `{release_dir}/a/b/` is created before symlinking `c`.

- **Symlink target:** uses the absolute path to `{shared_root}/{relative_path}`, not a
  relative symlink. This avoids broken links if the working directory changes.

- **Idempotent:** running the same deploy twice on the same release directory is safe.
  The second run deletes and re-creates the symlinks.
