# ADR-0001: Use Go as implementation language

**Status:** Accepted

## Context

Deployer v1 was written in Dart. The v2 rewrite has these hard requirements:

- Single native binary, no runtime on the target server or CI host
- Cross-compilation to Linux amd64/arm64 from a macOS dev machine
- A rich TUI: spinners, progress bars, interactive forms
- Accessible to contributors who are learning systems programming

The main candidates evaluated were **Go** and **Rust**.

## Decision

Use **Go**.

## Rationale

### Go vs Rust

| Criterion | Go | Rust |
|---|---|---|
| Learning curve | Low — small language spec | High — ownership, lifetimes, borrow checker |
| Cross-compilation | Trivial (`GOOS`/`GOARCH` env vars) | Requires toolchain setup (`cross` or cargo targets) |
| Compile speed | Fast | Slow on first build |
| TUI ecosystem | Charmbracelet suite (see ADR-0003) — purpose-built | Ratatui — powerful but low-level, no wizard/form equivalent |
| Binary size | ~5–10 MB | ~2–4 MB |
| Runtime | GC (acceptable for CLI) | No GC |
| Stdlib | Comprehensive for CLI/IO tasks | Comprehensive |

Rust's memory safety guarantees are its main advantage over Go. For a deployment CLI that is not handling untrusted input at high throughput, this advantage does not outweigh the friction cost.

### Go vs Dart (v1)

Dart was kept for v1 because the team was familiar with it. For v2:

- The Dart TUI ecosystem is minimal.
- `dart compile exe` produces native binaries but the compilation story is less standard.
- The Go toolchain is more widely available on CI runners and developer machines.

## Consequences

- Contributors need Go installed (managed via mise, see tasks roadmap).
- Error handling is explicit (`if err != nil`) — verbose but clear.
- The Charmbracelet TUI ecosystem becomes available (see ADR-0003).
- Cross-compilation is a one-liner: `GOOS=linux GOARCH=amd64 go build`.
- Hook template syntax changes from Liquid to Go `text/template` (see ADR-0004).
