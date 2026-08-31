# Linux / multi-user notes

Findings from evaluating `tars` for an unattended install on **Ubuntu Server 24.04 LTS, aarch64
(Raspberry Pi 5)**, driven by Ansible, on a machine with **several human users sharing it**.

Sections are ordered by how much they matter independent of that use case. §1 is broken everywhere,
including on a laptop. §2 is Linux-specific. §3 only matters if the machine is shared — take or
leave it.

Line references are against `59aa3f8`.

---

## 1. Bugs regardless of platform

### 1.1 `alias add-gh` uses backticks, so it executes at every shell start

`zsh/zshrc_aliases:28`:

```zsh
alias add-gh=`claude mcp add-json github '{"type":"http",...,"Authorization":"Bearer '"$(echo $GITHUB_TOKEN | cut -d '=' -f2)"'"}}'`
```

Every neighbouring alias uses single quotes; this one uses backticks. Two consequences:

1. `claude mcp add-json ...` **runs on every interactive shell startup**, not when the alias is
   invoked. The alias body then becomes whatever that command printed.
2. The token is interpolated onto a command line, so it is visible in `ps` for the life of the call.

Almost certainly meant to be `alias add-gh='...'`. Worth checking the other `add-*` aliases for the
same slip.

### 1.2 `tars pull` exits 0 even when every component fails

`cli/cmd/setup.go:233-240`:

```go
func (p SequentialPuller) PullAll() {
	for _, c := range p.Components {
		if err := c.Pull(); err != nil {
			fmt.Fprintf(p.Stderr, "  %s: %v\n", c.Name(), err)
		}
	}
}
```

`PullAll` returns nothing and `cmd/pull.go` returns `nil` unconditionally, so per-component failures
are printed and discarded.

This matters most for **CI and automation**: `test/e2e/Dockerfile` lines 28, 39 and 43 are `RUN`
assertions that can never fail on their own. The only real coverage there is the explicit
`test -f`/`test -d` checks, which cover three of six components — byobu and vim could break
completely and the e2e would still print `E2E OK`.

Suggestion: aggregate errors and return non-zero, or at minimum a `--strict` flag for CI. Anything
consuming `tars pull` programmatically currently has no way to detect failure.

### 1.3 Dead files

Not bugs, but they cost reader time:

- `zsh/bash_completion.d` and `zsh/zsh_completion.d` are **single-line text files** (contents:
  `kubectl completion bash` / `kubectl completion zsh`), not directories, and nothing in the repo
  references them.
- `walker.bttpreset` (73 KB) and `terminal/CloudWalk.terminal` are not referenced from any Go code.
- `scripts/pg-mcp.sh` still `sudo cp`s a binary, though `CLAUDE.md` says the `scripts/` bash tooling
  was removed.

---

## 2. Linux / arm64

The README is upfront that Linux is partial (`README.md:10`), so treat this as a gap list rather
than a complaint. The theme: **`tars pull` works well on Linux; `tars setup` is the untested half.**

`tars pull` needed no root, no sudo and no network, and behaved correctly — Linux font paths, the
darwin-only terminal no-op, and the backup/idempotency logic all did the right thing.

### 2.1 `.zshrc` sources `~/.zprofile`, but Linux installs to `~/.profile`

`cli/internal/paths/paths.go:97-100` writes `zsh/profile` to `~/.profile` on non-darwin. But
`zsh/zshrc:3` unconditionally does:

```zsh
source ~/.zprofile > /dev/null 2>&1
```

zsh reads `~/.zprofile`/`~/.zshenv`, never `~/.profile`. So on Linux everything in `zsh/profile`
(RVM, `~/.lmstudio/bin`, `~/.zprofile_local`) silently never loads — silently because the redirect
swallows the error.

Either write to `~/.zprofile` on Linux too, or have `.zshrc` source whichever exists. Related:
`zsh/zshrc_aliases:77,85` (`sp`, `vimp`) also hardcode `~/.zprofile`.

### 2.2 Package list gaps on Ubuntu 24.04

`cli/internal/pkg/registry.go:117-129`:

- **`gh` is not in the Ubuntu archives.** `sudo apt install -y gh` fails on 24.04; it needs the
  GitHub apt repo. Non-fatal today only because `InstallAll` continues past failures.
- **`bat` installs its binary as `batcat`** on Debian/Ubuntu. Anything aliasing `bat` won't find it.
- **`nodejs` from Noble is 18.x**, which is old for most tooling.
- The Linux list is much thinner than darwin's — no `eza`, `lazygit`, `k9s`, `terraform`,
  `golangci-lint`, `ansible`, `rustup`, `ghcup`. But `zsh/zshrc_aliases` defines aliases for several
  of them unconditionally, so a Linux user gets aliases for tools that were never installed.

### 2.3 The Neovim AppImage is fragile on server installs

`cli/internal/pkg/apt/apt.go:64-105` downloads `nvim-linux-aarch64.appimage` to `~/.local/bin/nvim`.

- **AppImages need FUSE, and Ubuntu Server 24.04 does not ship `libfuse2`.** On a headless box this
  fails at first run, not at install. `--appimage-extract` or a tarball/PPA avoids it entirely.
- The asset name has varied across upstream releases (`nvim-linux-arm64` vs `nvim-linux-aarch64`);
  worth pinning defensively.
- The version is hardcoded with no checksum verification.

### 2.4 `sudo` and apt hygiene for unattended runs

`cli/internal/pkg/apt/apt.go:24` shells `sudo apt ...` with no `-n`, no `DEBIAN_FRONTEND=noninteractive`,
and uses `apt` rather than `apt-get`. Under automation that means: a silent hang if sudo wants a
password with no TTY, apt's "unstable CLI interface" warning, and possible `needrestart` prompts on
24.04.

Also `GCloudCLI` (`apt.go:107-138`) adds a **system-wide** apt source and keyring from inside what is
otherwise a per-user command, and registers the repo without an `arch=` qualifier.

### 2.5 Executable bits are not preserved

`cli/internal/fsutil/fsutil.go:143-158` uses `os.Create`, so copies land 0644. `byobu/bin/*` (e.g.
`byobu/bin/1_git`) arrive non-executable.

### 2.6 macOS-isms that leak into a Linux shell

Cosmetic, but they produce warnings or wrong behavior:

- `zsh/zshrc:33` — `plugins=(... iterm2 ... macos ...)`; both are macOS-only oh-my-zsh plugins.
- `zsh/zshrc_aliases:44` — `alias ls='ls -aG'`. `-G` is *colorize* in BSD ls and *no group column*
  in GNU ls, so the alias silently means something different on Linux.
- `zsh/zshrc:68` — `export LC_ALL=en_US.UTF-8` warns unless that locale is generated.

### 2.7 Environment knobs (documenting what already works well)

These made unattended use straightforward and are worth keeping stable:

- `MACHINE_SETUP_REPO` — repo root, bypassing the upward search
- `MACHINE_SETUP_NO_FORM` — skips both TUIs (note: checks `!= ""`, so any value works)
- `MACHINE_SETUP_CONFIG_PATH`

Also good: `oh-my-zsh` is installed with `CHSH=no KEEP_ZSHRC=yes`, and `.zshrc_secret` is seeded only
when absent so it's never clobbered.

---

## 3. Multi-user — only relevant on a shared machine

None of this matters on a personal laptop, and some of it is a deliberate convenience there. Take it
as "if someone ever runs this on a shared box" — except §3.1, which we'd argue is worth fixing
regardless.

### 3.1 CWD-relative `.env` sourcing

`zsh/zshrc:82-85`:

```zsh
source_if_exists .env
source_if_exists .env.sh
source_if_exists ../.env.sh
source_if_exists ../../.env.sh
```

Every interactive shell sources these **relative to whatever directory it starts in**, including two
levels up.

On a laptop that's a nice per-project convenience. On a shared machine it's a lateral privilege
escalation: if user A opens a shell in any directory user B can write to — `/tmp`, a shared project
tree, a world-writable scratch dir — B's `.env` executes as A. Because it walks upward, a writable
*parent* is enough; A doesn't have to be in B's directory at all.

Even single-user, `cd`ing into a freshly cloned untrusted repo and getting a shell is a sharp edge.
`direnv` exists for this and solves it with explicit per-directory allow-listing.

Suggestion: drop the `../` and `../../` entries at minimum; ideally gate the whole block behind an
opt-in (`MACHINE_SETUP_AUTOENV=1`) or delegate to `direnv`.

### 3.2 Backups are written inside the repo clone

`cmd/pull.go:34`, `cmd/push.go:50`, `cmd/setup.go:261` all set:

```go
BackupRoot: filepath.Join(root, "backups")
```

With one shared clone, every user's previous dotfiles land in a directory the others can read, and
`nextVersion` numbers them globally — `backups/zsh/v7/` gives no indication of whose `.zshrc` that
is. It also forces the clone to be writable by everyone who runs `pull`, which in turn lets any user
edit the configs everyone else pulls.

`test/e2e/Dockerfile:15` (`chown -R tester:tester /repo`) sidesteps this by giving one user the whole
repo — which is why the test doesn't surface it.

Suggestion: an env override (`MACHINE_SETUP_BACKUP_ROOT`), defaulting under `$HOME` on Linux. Our
workaround is a per-user clone, which works but duplicates the fonts for every user.

### 3.3 `tars push` mutates the shared working tree

`cmd/push.go:55-59`, and `components/nvim.go:57` does `os.RemoveAll` on `<repo>/nvim`. One user
running `push` rewrites what every other user subsequently pulls. Anyone deploying a shared clone
should keep it read-only and not expose `push`.

### 3.4 `nvim` pull is never idempotent

`fsutil.SameContent` returns false for directories by design (`fsutil.go:74`), and `components/nvim.go:30-41`
unconditionally `RemoveAll`s `~/.config/nvim`. So every `pull` produces a new `backups/nvim/vN`.
Under config management that reports "changed" forever.

### 3.5 Per-user first-run cost

`nvim/lua/config/mason.lua` auto-installs nine language servers on first launch, per user. On a
4-core Pi with several users that's the same `gopls`/`rust_analyzer` downloaded N times. Not wrong,
just worth knowing before rolling out to a shared box.

---

## Suggested test change

`test/e2e/Dockerfile` is a good harness — assertions as `RUN` layers, non-root, throwaway `$HOME`,
explicit `MACHINE_SETUP_REPO`. Two changes would make it catch most of the above:

1. **Two users instead of one**, with the repo root-owned and read-only to them. That surfaces §3.2
   immediately and reflects a realistic shared-machine shape.
2. **Actually launch the shell** — `zsh -ic 'exit'` after installing zsh, and
   `nvim --headless -c quit`. Today nothing proves the pulled `.zshrc` even parses, which is why
   §2.1 and §2.6 are invisible to CI.

Also worth noting: CI builds without `--platform`, so it only ever exercises amd64. `.goreleaser.yaml`
already targets `linux/arm64`, and GitHub now offers arm64 runners.
