# tars sessions demo — live byobu sessions, one per project, laid out by its template.
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
Type "rm -rf /tmp/tars-demo && mkdir -p /tmp/tars-demo/Desktop/projects/api/.git /tmp/tars-demo/Desktop/projects/api/frontend /tmp/tars-demo/Desktop/projects/infra/.git /tmp/tars-demo/.config/tars && cd /tmp/tars-demo && printf 'projects:\n  api:\n    windows:\n      - {name: shell, command: ls}\n      - {name: web, dir: frontend, command: pwd}\n' > /tmp/tars-demo/.config/tars/projects.yaml && clear"
Enter
Show

# every git repo under projects_dir is a project
Type "tars s p"
Enter
Sleep 2.5s

# start api's session from its template (attaches if it is already running)
Type "tars s o api"
Enter
Sleep 2.5s

# the running sessions, and the manager over them
Type "tars s ls"
Enter
Sleep 2s
Type "tars s -i"
Enter
Sleep 2.5s
Down
Sleep 1s
Type "q"
Sleep 1s

Hide
Type "tmux kill-server"
Enter
