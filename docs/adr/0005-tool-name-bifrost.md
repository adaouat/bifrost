# ADR-0005: Name the tool "Bifrost"

**Status:** Accepted

## Context

The tool needed a name. "Deployer v2" is a working title, not a proper name. A good CLI
tool name should be:

- Memorable and distinctive — easy to recall, easy to Google
- Short enough to type comfortably as a binary name
- Meaningful in relation to what the tool does
- Not already taken by a major product or service in the same space

The team has an affinity for names drawn from geek culture (DC/Marvel comics, Tolkien).

## Candidates evaluated

| Name | Universe | Concept | Rejected because |
|---|---|---|---|
| **Bifrost** | Marvel / Norse mythology | The rainbow bridge connecting realms | — (chosen) |
| Anduril | Tolkien | The reforged sword (Narsil → Andúril) | Anduril Industries is a well-known defense company |
| Terminus | Asimov (Foundation) | The planet that coordinates the galaxy's future | Too close to "terminal", ambiguous |
| Zeta | DC (Adam Strange) | Zeta-Beam: instant teleportation to another world | Less universally recognizable |
| Warp | DC (Teen Titans) | Wormhole creation between any two points | Too generic, high collision risk |
| Blink | Marvel (X-Men) | Portal creation, instant teleportation | Extremely generic, ungoogleable |

## Decision

Name the tool **Bifrost** (binary: `bifrost`).

## Rationale

In Norse mythology — and the Marvel Cinematic Universe — the Bifrost is the rainbow bridge
that connects Asgard to the other Nine Realms. It is the mechanism by which things are
transported from one world to another.

This maps directly onto what the tool does:

> You have an artifact on your machine (or CI). Bifrost carries it to the server, unpacks
> it into a new release, links the shared state, flips the `current` pointer, and runs the
> lifecycle hooks. One world to another.

Additional reasons the name holds up:

- **Recognizable without being a brand.** The Bifrost appears in the Thor films, making it
  familiar to a broad audience — not just comics readers. Unlike Anduril or Gwaihir, no
  prior knowledge is required.
- **No namespace collision.** No major CLI tool, cloud service, or company uses this name
  in the deployment/DevOps space.
- **Works as a binary name.** `bifrost artifact push`, `bifrost release enable` — the name
  reads naturally as a command.
- **Sounds serious.** It is a real word with weight, not a cute pun or acronym.

## Consequences

- The binary is named `bifrost` (lowercase, as is convention for CLI tools).
- The project is referred to as "Bifrost" in documentation and conversation.
- The Go module path and repository name should use `bifrost`.
- `CLAUDE.md` and all specs refer to it as Bifrost, not "Deployer v2".
