---
description: Land the current work through a pull request and watch its checks
argument-hint: "[title]"
---
Land the current work as a pull request${1:+ titled "$1"}.

1. Confirm the full test suite and lint are green locally; stop and report if not.
2. Create a branch off the default branch if we are still on it. Commit in logical chunks with conventional-commit messages; never `git stash`, never force-push, never commit straight to the default branch.
3. Push and open the PR with a body covering what changed, why, and how it was verified.
4. Watch the PR checks until every one finishes. On a red check, read its log, fix the cause with the same TDD rules, commit, push, and watch again.
5. Report the PR URL and the final state of each check. Do not merge unless I say so.
