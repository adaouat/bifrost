# bifrost

Atomic deployment CLI.

## Install

### `go install`

```bash
go install github.com/adaouat/bifrost/cmd/bifrost@latest
```

### mise

```bash
mise use github:adaouat/bifrost
```

Or declare it in your `.mise.toml` / `mise.toml`:

```toml
[tools]
"github:adaouat/bifrost" = "latest"
```

### Prebuilt binary

Download the raw binary for your platform from the
[releases page](https://github.com/adaouat/bifrost/releases). Assets are named
`bifrost_<version>_<os>_<arch>` (no archive wrapper), alongside a `checksums.txt`
for verification.

```bash
# example: macOS arm64 — replace <version> with the release tag
curl -L -o bifrost "https://github.com/adaouat/bifrost/releases/download/<version>/bifrost_<version>_darwin_arm64"
chmod +x bifrost
./bifrost --version
```

> **macOS / Gatekeeper:** if the binary is blocked after download, clear the quarantine flag:
> ```bash
> xattr -d com.apple.quarantine bifrost
> ```

bifrost prints a one-line upgrade hint when a newer release exists; re-run your install
method (`mise upgrade bifrost`, `go install …@latest`, or the curl command) to upgrade.
