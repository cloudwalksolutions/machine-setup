# End-to-end test of the `tars pull` bootstrap flow, fully isolated.
#
# Safety: runs as a non-root user with HOME forced to a throwaway dir, and only
# exercises the offline config/font pull — NO apt, NO sudo, NO network at runtime.
# Every assertion is a `RUN` line, so a failure fails `docker build`. The container
# is discarded afterward, so it structurally cannot mutate a real machine.
#
# Run:  docker build -f test/e2e/Dockerfile -t tars-e2e .
FROM golang:1.24-bookworm

# Non-root throwaway user.
RUN useradd -m tester
WORKDIR /repo
COPY . /repo
RUN chown -R tester:tester /repo
USER tester

# Build the CLI from the cli/ module (build-time network is fine; runtime is not).
WORKDIR /repo/cli
RUN go build -o /home/tester/tars .

# Throwaway HOME + explicit repo root so `tars` never guesses a real path.
ENV HOME=/home/tester/fakehome
ENV MACHINE_SETUP_REPO=/repo
RUN mkdir -p "$HOME"

# 1. Apply configs (offline: dotfiles + Linux fonts; terminal is a darwin no-op).
RUN /home/tester/tars pull

# 2. Assert the expected files landed.
RUN test -f "$HOME/.zshrc" \
 && test -d "$HOME/.config/nvim" \
 && test -d "$HOME/.local/share/fonts" \
 && ls "$HOME/.local/share/fonts/"*.ttf >/dev/null

# 3. Backup path: modify a pulled dotfile, re-pull, and assert a versioned backup
#    was created (SafeCopy skips byte-identical files, so the change is required).
RUN echo '# e2e local change' >> "$HOME/.zshrc" \
 && /home/tester/tars pull \
 && ls -d /repo/backups/zsh/v*/ >/dev/null

# 4. Idempotency: a clean re-pull must still succeed.
RUN /home/tester/tars pull

RUN echo "E2E OK"
