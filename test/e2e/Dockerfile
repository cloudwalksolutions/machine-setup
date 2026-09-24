# End-to-end test of the `tars pull` bootstrap flow, fully isolated.
#
# Shape: a root-owned, READ-ONLY repo clone shared by two users (the bastion
# scenario), plus a third user with no clone at all (the brew-install scenario,
# served by the dotfiles embedded in the binary). Only the offline pull flow
# runs — NO apt installs, NO sudo, NO network at runtime (zsh is installed at
# build time for the shell smoke test). Every assertion is a `RUN` line, so a
# failure fails `docker build`. The container is discarded afterward.
#
# Run:  docker build -f test/e2e/Dockerfile -t tars-e2e .
FROM golang:1.24-bookworm

# zsh for the shell smoke test (build-time network is fine; runtime is not).
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y zsh && rm -rf /var/lib/apt/lists/*

# Root-owned repo — users can read it but never write into it.
WORKDIR /repo
COPY . /repo
RUN cd /repo/cli && go build -o /usr/local/bin/tars .

RUN useradd -m usera && useradd -m userb && useradd -m userc

# ── usera: pull from the shared read-only clone ──────────────────────────
USER usera
WORKDIR /home/usera
ENV HOME=/home/usera
ENV TARS_REPO=/repo

RUN tars pull

# Every file-producing component landed (terminal and profiles are no-ops here; pull exiting 0 covers them).
RUN test -f "$HOME/.vimrc" \
 && test -f "$HOME/.zshrc" \
 && test -f "$HOME/.zshrc_aliases" \
 && test -f "$HOME/.profile" \
 && test -f "$HOME/.byobu/.tmux.conf" \
 && test -x "$HOME/.byobu/bin/1_git" \
 && test -d "$HOME/.config/nvim" \
 && ls "$HOME/.local/share/fonts/"*.ttf >/dev/null \
 && test -x "$HOME/.claude/hooks/block-unreviewable-edits.sh" \
 && test -f "$HOME/.claude/CLAUDE.md" \
 && grep -q 'block-unreviewable-edits.sh' "$HOME/.claude/settings.json"

# The pulled shell configs must actually parse on Linux.
RUN zsh -ic 'exit'
RUN bash -lc 'exit'

# Backups are per-user under $HOME, never inside the repo clone. (v1 was made
# on the first pull: useradd -m seeds ~/.profile from /etc/skel, so replacing
# it archived the skel copy; the modified .zshrc becomes v2.)
RUN echo '# e2e local change' >> "$HOME/.zshrc" \
 && tars pull \
 && test -d "$HOME/.local/state/tars/backups/zsh/v1" \
 && test -d "$HOME/.local/state/tars/backups/zsh/v2" \
 && test ! -e /repo/backups

# Idempotency: a clean re-pull adds no new backup version — nvim and claude included.
RUN tars pull \
 && test ! -e "$HOME/.local/state/tars/backups/zsh/v3" \
 && test ! -e "$HOME/.local/state/tars/backups/nvim" \
 && test ! -e "$HOME/.local/state/tars/backups/claude"

# Machine-local settings survive the merge; the hook entry is added exactly once.
RUN echo '{"autoMode":{"allow":["x"]}}' > "$HOME/.claude/settings.json" \
 && tars pull \
 && grep -q '"autoMode"' "$HOME/.claude/settings.json" \
 && test "$(grep -c 'block-unreviewable-edits.sh' "$HOME/.claude/settings.json")" = "1" \
 && test -f "$HOME/.local/state/tars/backups/claude/v1/settings.json"

# The init wizard runs headless and honors its choices on the next pull.
RUN TARS_NO_FORM=1 tars claude init \
 && grep -q 'rules:' "$HOME/.config/tars/config.yaml" \
 && mkdir -p "$HOME/proj" && TARS_NO_FORM=1 tars claude project "$HOME/proj" \
 && grep -q '^# CLAUDE.md' "$HOME/proj/CLAUDE.md" \
 && if TARS_NO_FORM=1 tars claude project "$HOME/proj" >/dev/null 2>&1; then echo 'expected refusal without --force' && exit 1; fi

# pi: agent, prompts, edit guard, permissions and AGENTS.md land from pull alone;
# providers are never written without saved choices.
RUN test -f "$HOME/.pi/agent/agents/tars.md" \
 && test -f "$HOME/.pi/agent/prompts/tdd.md" \
 && test -f "$HOME/.pi/agent/extensions/block-unreviewable-edits.ts" \
 && test -f "$HOME/.pi/agent/extensions/pi-permission-system/config.json" \
 && grep -q '^# Working rules' "$HOME/.pi/agent/AGENTS.md" \
 && grep -q 'npm:pi-subagents' "$HOME/.pi/agent/settings.json" \
 && test ! -e "$HOME/.pi/agent/models.json" \
 && tars pull >/dev/null \
 && test ! -e "$HOME/.local/state/tars/backups/pi"

# Failures must surface as a non-zero exit (an unwritable HOME cannot succeed).
RUN mkdir -p /home/usera/ro && chmod 555 /home/usera/ro \
 && if HOME=/home/usera/ro tars pull >/dev/null 2>&1; then echo 'expected tars pull to fail' && exit 1; fi

# ── userb: a second user off the same read-only clone ────────────────────
USER userb
WORKDIR /home/userb
ENV HOME=/home/userb

RUN tars pull \
 && test -f "$HOME/.zshrc" \
 && test -x "$HOME/.byobu/bin/1_git" \
 && zsh -ic 'exit'

# ── userc: no clone at all — the embedded assets materialize on demand ───
USER userc
WORKDIR /home/userc
ENV HOME=/home/userc
ENV TARS_REPO=

RUN tars pull \
 && test -f "$HOME/.local/share/tars/repo/zsh/zshrc" \
 && test -f "$HOME/.zshrc" \
 && test -x "$HOME/.byobu/bin/1_git" \
 && ls "$HOME/.local/share/fonts/"*.ttf >/dev/null \
 && zsh -ic 'exit'

RUN echo "E2E OK"
