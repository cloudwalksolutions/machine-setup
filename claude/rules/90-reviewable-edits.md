## Reviewable edits only
Change files with the Edit tool (exact string replacement) or the Write tool (full file) so every change is reviewable as a diff. Never `sed -i`, `awk -i inplace`, `perl -i`, interpreter heredocs or one-liners that write files, or shell redirects onto source files. A hook enforces this; do not try to route around it.
