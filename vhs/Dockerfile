# Recording environment for the README demo tapes: the official VHS image
# plus byobu/tmux for the sessions demo. Linux keeps the terminal component a
# no-op so `tars pull` records cleanly against a scratch HOME.
FROM ghcr.io/charmbracelet/vhs
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y byobu tmux \
  && rm -rf /var/lib/apt/lists/*
