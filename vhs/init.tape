# tars init — scratch HOME, no repo clone (embedded dotfiles). Ctrl+C at the
# pickers skips installs and agent setups, so nothing outside HOME is touched.
# Regenerate: make demos
Output vhs/init.gif

Require tars

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 820
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set WindowBar Colorful

Env HOME "/tmp/tars-init"

Hide
Type "rm -rf /tmp/tars-init && mkdir -p /tmp/tars-init && cd /tmp/tars-init && clear"
Enter
Show

Type "tars init"
Enter
Wait+Screen@30s /Welcome to tars/
Sleep 1.5s
Enter

# One category per page, every tool listed; walk the editors, then skip installing
Wait+Screen@60s /Agentic Coding Tools/
Sleep 1.5s
Enter
Wait+Screen@10s /Terminal Utilities/
Sleep 1s
Down@300ms 8
Sleep 1s
Ctrl+C

Wait+Screen@180s /Agent setups/
Sleep 1.5s
Ctrl+C

Wait+Screen@60s /Setup complete/
Sleep 3s

Hide
Type "rm -rf /tmp/tars-init"
Enter
