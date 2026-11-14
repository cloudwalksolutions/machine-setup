Output vhs/pull.gif

Require tars

Set Shell "bash"
Set FontSize 18
Set Width 1400
Set Height 900
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set WindowBar Colorful

Env HOME "/tmp/tars-pull"
Env COLORTERM "truecolor"

Hide
Type "rm -rf /tmp/tars-pull && mkdir -p /tmp/tars-pull && cd /tmp/tars-pull && tars pull -y >/dev/null"
Enter
Wait@60s
Type "printf '\nalias gs=git-status\nexport EDITOR=nvim\n' >> ~/.zshrc && echo 'set relativenumber' >> ~/.vimrc"
Enter
Type "rm ~/.byobu/datetime.tmux && echo 'return {}' > ~/.config/nvim/lua/scratch.lua && clear"
Enter
Show

Type "tars pull"
Enter
Wait+Screen@60s /Review pull/
Sleep 2.5s

Down@600ms 2
Sleep 1.5s
Type@800ms "nn"
Sleep 1.5s

Tab
Sleep 500ms
Down@300ms 6
Sleep 1.5s
Tab
Sleep 500ms

Down@500ms 4
Sleep 1s
Space
Sleep 2s

Enter
Wait@60s
Sleep 3s

Type "tars pull"
Enter
Wait+Screen@60s /Review pull/
Sleep 2s
Enter
Wait@60s
Sleep 1.5s

Type "tars pull"
Enter
Wait@60s
Sleep 3s

Hide
Type "rm -rf /tmp/tars-pull"
Enter
