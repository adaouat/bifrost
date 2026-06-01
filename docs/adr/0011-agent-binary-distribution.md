# ADR-0011: Agent binary distribution via GitHub Releases with local cache

**Status:** Accepted

## Context

The self-agent model (see [ADR-0009](0009-agent-model.md)) requires uploading a Bifrost
binary to each target server. The binary must:

- Match the remote OS and architecture (the client and server may differ, e.g. macOS
  developer deploying to Linux amd64).
- Exactly match the running client version to guarantee protocol compatibility.
- Be available without requiring Go on the client machine.

Options considered:

| Option | Description |
|---|---|
| **Use the running binary** | Upload the client binary to the remote as-is |
| **Cross-compile at runtime** | Run `go build` for the target arch on the client |
| **Download from GitHub Releases** | Fetch the pre-built binary for the detected arch |
| **Require pre-installed agent** | Expect `bifrost` to already be on the server |

## Decision

**Download from GitHub Releases at the client's own version, cache locally by version and arch.**

### Flow

```
1. SSH: uname -s && uname -m           → detect remote OS and arch
2. Map to goreleaser naming:
     Linux  / x86_64   → linux_amd64
     Linux  / aarch64  → linux_arm64
     Darwin / x86_64   → darwin_amd64
     Darwin / arm64    → darwin_arm64
3. Check local cache:
     {os.UserCacheDir()}/bifrost/agents/{version}/{os}_{arch}/bifrost
4. Cache miss → download:
     https://github.com/adaouat/bifrost/releases/download/v{version}/
       bifrost_{version}_{os}_{arch}.tar.gz
   Extract binary, write to cache, chmod +x
5. SFTP upload cached binary → /tmp/bifrost-<uuid>/bifrost-agent
```

### Escape hatch

`--agent-binary <local-path>` flag on `deploy` and all `release` commands bypasses the
download entirely and uploads the specified binary. This covers:

- Pre-release or development builds (no GitHub Release exists yet).
- Air-gapped environments with no internet access from the CI runner.
- Custom builds with non-standard patches.

## Rationale

**Download from GitHub Releases over using the running binary:**
The client binary is built for the client OS/arch. A macOS arm64 binary cannot run on a
Linux amd64 server. The running binary cannot be used when client and server differ.

**Download from GitHub Releases over cross-compiling at runtime:**
Cross-compilation requires Go on the client — defeating the single-binary distribution
model. Pre-built release artifacts already exist for all supported platforms (goreleaser
produces them in the release workflow).

**Download from GitHub Releases over pre-installed agent:**
Requiring a pre-installed agent shifts version management responsibility onto the user and
risks version skew. Self-upload guarantees the agent always matches the client.

**Local cache:**
Release artifacts are immutable (a given version never changes). Caching by version + arch
means subsequent deploys to the same arch skip the download entirely. The cache uses the
standard OS cache directory (`os.UserCacheDir()`), which is automatically respected by
the OS's cache eviction policies.

## Consequences

- A new `internal/transport/agent.go` package handles arch detection, download, and
  caching.
- The `deploy` and `release` commands gain an `--agent-binary` flag.
- CI environments without internet access must use `--agent-binary`.
- If a GitHub Release does not exist for the current version (e.g. a dev build), Bifrost
  fails with a clear error message pointing to `--agent-binary`.
- Cache is never invalidated automatically for a given version; old versions accumulate
  until the user clears `{os.UserCacheDir()}/bifrost/agents/`.
