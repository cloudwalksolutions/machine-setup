# tars hero demo — runs against a scratch HOME with no repo clone (the
# dotfiles come from the binary's embedded assets). Nothing real is touched.
# Regenerate: make demos   (or: docker build -t tars-vhs vhs && see Makefile)
Output vhs/demo.gif

Require tars

Set Shell "bash"
Set FontSize 20
Set Width 1200
Set Height 650
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
Sleep 1.5s

Type "tars pull"
Enter
Wait+Screen@60s /Pull complete/
Sleep 1.5s

Type "ls -a ~ | head -8"
Enter
Sleep 2s

# Idempotent: a re-run reattaches to reality, changes nothing
Type "tars pull"
Enter
Wait+Screen@60s /Pull complete/
Sleep 3s

Hide
Type "rm -rf /tmp/tars-demo"
Enter
