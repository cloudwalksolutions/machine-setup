---
description: Survey the code and propose a plan with no open unknowns
argument-hint: "<task>"
---
Plan this task before touching any file: $@

1. Survey the relevant code, tests, and docs. Use subagents for wide searches and report only conclusions.
2. List the decisions that are mine to make and ask them now, one question each, with a recommended option.
3. Write the plan: context (why), the files to change, the existing functions to reuse, the TDD order (one red spec per step), and how the result is verified end to end.
4. Stop and wait for my go. Do not start implementing.
