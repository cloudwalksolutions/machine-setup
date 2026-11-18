---
name: pr
description: Land the current work through a pull request and watch its checks
argument-hint: "[title]"
disable-model-invocation: true
---
Land the current work as a pull request. Title, if given: $ARGUMENTS

1. Confirm the full test suite and lint are green locally; stop and report if not.
2. Create a branch off the default branch if we are still on it, and commit in logical chunks.
3. Push and open the PR with the description the rules ask for.
4. Watch the PR checks until every one finishes. On a red check, read its log, fix the cause test-first, commit, push, and watch again.
5. Hand back with the PR link, its live state, and the final state of each check. Never merge it yourself.
