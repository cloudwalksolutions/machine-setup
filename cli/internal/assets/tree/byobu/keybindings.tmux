set -g prefix2 F12

unbind-key -T root F2
unbind-key -T root F3
unbind-key -T root F4
unbind-key -T root F5
unbind-key -T root F6
unbind-key -T root F7

bind y set-window-option synchronize-panes

# Make splitting panes more intuitive
bind v split-window -h -c "#{pane_current_path}"
bind h split-window -v -c "#{pane_current_path}"

# Resize panes using the home row too!
bind -r H resize-pane -L 5
bind -r J resize-pane -D 5
bind -r K resize-pane -U 5
bind -r L resize-pane -R 5

# Make copy/paste same as vim
bind Escape copy-mode
unbind p
bind p paste-buffer
bind-key -T copy-mode-vi 'v' send-keys -X begin-selection
bind-key -T copy-mode-vi 'y' send-keys -X copy-selection-and-cancel

unbind-key -n C-s
unbind-key -n C-a
set -g prefix ^A
set -g prefix2 F12
bind a send-prefix

# Rename via tmux's brace form; byobu's quoted-string `%%` breaks under tmux 3.6+
bind-key A     command-prompt -I "#W" { rename-window "%%" }
unbind-key -n F8
bind-key -n F8 command-prompt -p "(rename-window) " -I "#W" { rename-window "%%" }

# Free Option/Alt+arrows so the shell gets them for word-by-word movement
unbind-key -n M-Left
unbind-key -n M-Right
unbind-key -n M-Up
unbind-key -n M-Down

# Shift+Option/Alt+arrows: switch windows (left/right) and sessions (up/down)
# Overrides byobu's pane-resize default; resizing stays on prefix H/J/K/L.
bind-key -n M-S-Left previous-window
bind-key -n M-S-Right next-window
bind-key -n M-S-Up switch-client -p
bind-key -n M-S-Down switch-client -n
