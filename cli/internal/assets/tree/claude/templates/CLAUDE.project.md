# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in {{.Name}}.

## Project Overview

{{.Description}}

## Development Philosophy

- Strict outside-in TDD: one failing test, red, minimal green, refactor.
- Simplest change that works; no speculative abstractions.
- Names carry intent; no explanatory comments.

## Claude's Role and Constraints

- Land work through branches and pull requests; CI proves validity.
- The user owns dev servers; never start, stop, or build over them.
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

## Sub-project CLAUDE.md Files

<!-- list nested CLAUDE.md files; read them there, not here -->

## Gotchas and Known Issues

<!-- surprises worth knowing before touching the code -->

## Maintaining This File

Capture user preferences and corrections here as they come up so they are not repeated.
