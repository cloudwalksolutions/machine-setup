# AGENTS.md

Guidance for coding agents (Claude Code, pi, Gemini CLI) when working in {{.Name}}.
`CLAUDE.md` and `GEMINI.md` import this file; edit here only.

## Project Overview

{{.Description}}

## Development Philosophy

- Strict outside-in TDD: one failing test, red, minimal green, refactor.
- Simplest change that works; no speculative abstractions.
- Comments are for users, tests are for maintainers: names and tests carry intent, and there are no explanatory comments.

## The Agent's Role and Constraints

- Land work as small, focused pull requests; CI proves validity.
- The user's dev servers are theirs; start and own separate instances for testing.
- Edit files only with the Edit/Write tools; never stream edits.
- Verify behavior with tests, not by hand; browser MCP proves visuals only.

## Commands

```
{{.TestCommand}}
```

Always use the project's canonical targets above, never raw tooling.

## Architecture

<!-- backend / frontend / infra layout -->

## Testing and Verification

<!-- test layers, coverage bar, e2e conventions, visual gate -->

## Configuration and Secrets

Secrets come from the environment or git-ignored local files, never inline.

## Sub-project AGENTS.md Files

<!-- list nested AGENTS.md files; read them there, not here -->

## Gotchas and Known Issues

<!-- surprises worth knowing before touching the code -->

## Maintaining This File

Capture user preferences and corrections here as they come up so they are not repeated.
