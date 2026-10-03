# tars hero demo — runs against a scratch HOME with no repo clone (the
# dotfiles come from the binary's embedded assets). Nothing real is touched.
# Regenerate: make demos
Output vhs/demo.gif

Require tars

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 1080
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set WindowBar Colorful

Env HOME "/tmp/tars-demo"

Hide
Type "rm -rf /tmp/tars-demo && mkdir -p /tmp/tars-demo && cd /tmp/tars-demo && clear"
Enter
Show

Type "tars --version"
Enter
Sleep 1.2s

# Preview first: every write it would make, and the backup each one gets
Type "tars pull --dry-run"
Enter
Wait+Screen@60s /Dry run complete/
Sleep 3s

Hide
Type "clear"
Enter
Show

Type "tars pull"
Enter
Wait+Screen@60s /Pull complete/
Sleep 2s

Hide
Type "clear"
Enter
Show

# Idempotent: a second look finds nothing left to do
Type "tars pull --dry-run"
Enter
Wait+Screen@60s /Dry run complete/
Sleep 3s

Hide
Type "rm -rf /tmp/tars-demo"
Enter
