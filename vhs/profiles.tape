# tars profiles demo — one machine, several git/GitHub identities, chosen by
# project dir. Runs against a scratch HOME; `gh` is the stand-in in vhs/bin so
# `use` shows its real success line without a GitHub login.
Output vhs/profiles.gif

Require tars
Require git
Require gh

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
Type "rm -rf /tmp/tars-demo && mkdir -p /tmp/tars-demo/.config/tars /tmp/tars-demo/.ssh /tmp/tars-demo/projects && cd /tmp/tars-demo && touch .ssh/id_rsa.cws .ssh/id_rsa.oss && git init -q projects/cloudwalk/api && git init -q projects/oss/tool && printf 'projects_dir: ~/projects\nfull_name: Ada Lovelace\nprofiles:\n  - name: cloudwalk\n    alias: cws\n    email: ada@cloudwalk.example\n    github: ada-cws\n  - name: oss\n    alias: oss\n    email: ada@oss.example\n    github: ada-oss\n' > .config/tars/profiles.yaml && clear"
Enter
Show

Type "tars p list"
Enter
Sleep 2s

# use: renders the gitconfigs, records the default, switches gh
Type "tars p use cws"
Enter
Sleep 2s

Hide
Type "clear"
Enter
Show

Type "tars p list"
Enter
Sleep 1.5s

# git picks the identity from the directory — no per-repo config
Type "git -C projects/oss/tool config user.email"
Enter
Sleep 1.5s

Type "git -C ~ config user.email"
Enter
Sleep 2s

Type "tars p show"
Enter
Sleep 3s

Hide
Type "rm -rf /tmp/tars-demo"
Enter
