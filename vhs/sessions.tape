# tars sessions demo — declarative byobu sessions from a simple yaml config.
# Runs against a scratch HOME; byobu/tmux live inside the recording container.
Output vhs/sessions.gif

Require tars
Require byobu

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
Type "rm -rf /tmp/tars-demo && mkdir -p /tmp/tars-demo/work/api /tmp/tars-demo/work/infra /tmp/tars-demo/.config/tars && cd /tmp/tars-demo && printf 'sessions:\n  - name: cloudwalk\n    dirs:\n      - ~/work/api\n      - ~/work/infra\n' > /tmp/tars-demo/.config/tars/sessions.yaml && clear"
Enter
Show

Type "tars sessions list"
Enter
Sleep 2s

# create-or-attach: one window per dir, idempotent
Type "tars s o cloudwalk"
Enter
Sleep 5s

Hide
Type "tars"
Enter
