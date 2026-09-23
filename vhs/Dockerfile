# Recording environment for the README demo tapes: the official VHS image
# plus byobu/tmux for the sessions demo and git for the profiles demo. Linux
# keeps the terminal component a no-op so `tars pull` records cleanly against
# a scratch HOME. Pinned: v0.12.0 reports success but writes no GIF.
FROM ghcr.io/charmbracelet/vhs:v0.10.0
RUN apt-get update --allow-releaseinfo-change \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y byobu tmux git \
  && rm -rf /var/lib/apt/lists/*
# gh stand-in: `tars p use` shells out to it and the demo has no GitHub login.
COPY bin/gh /usr/local/bin/gh
RUN chmod +x /usr/local/bin/gh
