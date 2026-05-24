# Claude behaviour rules

## Architectural decisions

- Challenge design choices when something seems wrong, suboptimal, or inconsistent with
  the stated goals — even if the user proposed it. State the concern clearly and explain
  why, then let the user decide.
- Never silently accept a decision that contradicts an ADR or a rule in this project.
  Surface the conflict and ask for clarification.

## Task discipline

- Never implement more than one roadmap task per session without explicit user approval.
- Never implement anything not on the current roadmap without asking first.
- If a task feels too large to implement safely in one step, break it down and propose
  the breakdown before starting.
- Always follow the two-step commit flow (see workflow rules). No shortcuts.

## TDD discipline

- Always write the failing test before writing implementation code.
- Never write implementation first and tests after.
- If the user asks to skip tests, push back and explain why the tests are needed.

## Code discipline

- Never implement more than what the current task requires.
- Never refactor, clean up, or improve surrounding code as part of a task unless the task
  explicitly asks for it.
- Never add features "while we're here".
- If linting or tests fail, fix the root cause — never silence the error or skip the check.

## Roadmap discipline

- Before starting a task: read the current roadmap, confirm the task is `[ ]`.
- After completing a task: mark it `[x]`, commit implementation and roadmap together.
- If a new task is discovered mid-implementation, add it to the roadmap — do not silently
  implement it.

## Irreversible operations

- Always ask before: force-pushing, deleting files/branches, resetting git history,
  modifying CI/CD pipelines, pushing to remote.
- Confirm scope explicitly — approval for one action does not imply approval for similar
  actions in other contexts.

## Memory

- Save important architectural decisions, user preferences, and non-obvious constraints
  to memory so they survive context resets.
- Verify memory before acting on it — stale memory is worse than no memory.
