## Pull requests and CI
The default way to land work is a branch and a pull request, never a direct push to a protected branch. Prove validity with tests that run in CI wherever possible; a green local run is not proof. Never `git stash`, never rewrite history that has been pushed, and never claim done while CI is red or unverified.
