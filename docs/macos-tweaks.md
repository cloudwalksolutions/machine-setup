# Plan: `macos` component — opt-in macOS system tweaks

## Why

Some of what makes a Mac usable as a terminal-first dev box lives in macOS
preference state, not in dotfiles: today the byobu Ctrl+Option+Up/Down bindings
only work because two Mission Control shortcuts were disabled by hand on this
machine. That kind of tweak is invisible to `tars pull`, so a fresh Mac silently
regresses. The `macos` component makes these tweaks declarative, opt-in,
idempotent, snapshot-backed and revertible, the same way tars treats dotfiles.

Everything here is **opt-in**: no config means the component is a no-op. On
Linux it is always a no-op.

## Shape

- Ninth component (`cli/internal/components/macos.go`, `NewMacOSForOS(opts, goos)`),
  registered last in `AllPullable`; not `Pushable`. Pull applies only the tweaks
  listed under `macos.tweaks` in the tars config.
- `tars macos` (alias `m`) is the wizard: a huh multi-select over the registry
  with **nothing pre-checked** on first run (previously chosen tweaks pre-checked
  on re-runs), saves the selection, then pulls the component. `tars init` shows
  the same form on darwin after the tool picker; `TARS_NO_FORM` selects nothing.
- `tars macos list` (alias `l`) prints every tweak with `applied` / `pending` /
  `not selected`, computed by reading current values.
- `tars macos revert [id…]` restores the snapshotted previous values for the
  given tweaks (all selected when no ids) and removes them from the config.

## Registry (`cli/internal/macos/registry.go`, mirrors `pkg/registry.go`)

```go
type Tweak struct {
    ID      string   // stable slug, stored in config
    Title   string   // one line, shown in the form
    Why     string   // one line, shown under the title
    Steps   []Step   // applied in order
    Refresh []string // processes to `killall` afterwards (Dock, Finder, SystemUIServer)
    Logout  bool     // takes effect only after logging out
}

// Step is one reversible write. Implementations: Default (defaults write/read/delete,
// optional -currentHost), SymbolicHotkey (com.apple.symbolichotkeys entry + activateSettings),
// ModifierMapping (per-keyboard Ctrl<->Caps mapping).
type Step interface {
    Current(r Runner) (string, error)      // "" when absent
    Apply(r Runner) error
    Restore(r Runner, previous string) error // previous == "" deletes the key
}
```

Every external effect goes through an injected `Runner` (`defaults`, `hidutil`,
`killall`, `activateSettings`), so unit specs drive a fake and assert the exact
argv; the real runner is exercised manually and on the `macos-latest` CI leg
(unit only, no writes).

## Tweaks in v1

| ID | What it writes | Refresh |
|---|---|---|
| `mission-control-arrows` | `com.apple.symbolichotkeys` entries 34 and 35 `enabled=false` (Ctrl+Option+Up/Down slow-motion variants) | `activateSettings -u` |
| `spaces-fixed-order` | `com.apple.dock mru-spaces -bool false` | Dock |
| `keyboard-fast-repeat` | `NSGlobalDomain KeyRepeat -int 2`, `InitialKeyRepeat -int 15`, `ApplePressAndHoldEnabled -bool false` | logout |
| `keyboard-fn-keys` | `NSGlobalDomain com.apple.keyboard.fnState -bool true` (F1–F12 are function keys; media via Fn) | logout |
| `keyboard-swap-ctrl-caps` | per-keyboard `com.apple.keyboard.modifiermapping.<vendor>-<product>-0` in the `-currentHost` global domain, mapping Caps Lock (0x700000039) → Left Control (0x7000000E0) and back, for every keyboard `hidutil list` reports | logout |
| `text-no-autocorrect` | `NSGlobalDomain NSAutomaticSpellingCorrectionEnabled`, `NSAutomaticCapitalizationEnabled`, `NSAutomaticQuoteSubstitutionEnabled`, `NSAutomaticDashSubstitutionEnabled`, `NSAutomaticPeriodSubstitutionEnabled` all `-bool false` | none |
| `finder-dev` | `com.apple.finder AppleShowAllFiles true`, `ShowPathbar true`; `NSGlobalDomain AppleShowAllExtensions true`; `com.apple.desktopservices DSDontWriteNetworkStores true`, `DSDontWriteUSBStores true` | Finder |
| `dock-minimal` | `com.apple.dock autohide -bool true`, `show-recents -bool false` | Dock |
| `screenshots-dir` | `com.apple.screencapture location ~/Screenshots` (dir created), `disable-shadow -bool true` | SystemUIServer |
| `trackpad-tap-drag` | `com.apple.AppleMultitouchTrackpad` and `com.apple.driver.AppleBluetoothMultitouch.trackpad`: `Clicking true`, `TrackpadThreeFingerDrag true`; `NSGlobalDomain com.apple.mouse.tapBehavior -int 1` (also `-currentHost`) | logout |

Dropped by decision: Globe/Fn key behaviour. Not in scope: iTerm2 preferences
(terminal component), anything requiring `sudo` or Full Disk Access.

## Safety

- **Snapshot before first write.** Applying a tweak first records each step's
  `Current()` value into `<BackupRoot>/macos/vN/snapshot.yaml` keyed by tweak and
  step. Revert reads the newest snapshot that contains the tweak. Absent keys are
  recorded as absent and restored with `defaults delete`.
- **Idempotent.** A step whose `Current()` already equals the target is skipped;
  a tweak with no changed steps triggers no refresh. Re-running `tars pull` is a
  no-op and writes no new snapshot.
- **Scoped.** Only selected tweaks are read or written. Refresh kills only the
  processes listed by tweaks that actually changed something, once each.
- **Honest output.** Pull prints which tweaks needed a logout; `tars macos list`
  shows the live state, not the config.
- `keyboard-swap-ctrl-caps` only covers keyboards attached at apply time; a new
  keyboard needs another `tars pull`. The plan notes this in the form's `Why`.

## Config and paths

```yaml
macos:
  tweaks: [mission-control-arrows, spaces-fixed-order, keyboard-swap-ctrl-caps]
```

`config.Config` gains `MacOS MacOSConfig{Tweaks []string}`; `Save` sets `macos`.
No new dotfile dir and nothing to embed.

## TDD order (one red spec at a time)

1. `internal/macos`: registry has the ten ids above, unique, each with a title,
   a why, at least one step.
2. `Default` step: `Current` parses `defaults read` (absent → `""`), `Apply`
   builds the right argv for bool/int/string and `-currentHost`, `Restore` writes
   back or deletes.
3. `SymbolicHotkey` step: `Apply` emits the `-dict-add` plist fragment for ids
   34/35 and runs `activateSettings -u`; `Current` reads `enabled`.
4. `ModifierMapping` step: keyboards enumerated from `hidutil list` output;
   one `-currentHost write` per keyboard; `Restore` deletes the key.
5. `components.MacOS.Pull`: no config → no runner calls; selected tweaks apply
   only changed steps; snapshot written under `macos/v1` before the first write;
   refresh once per process; linux → no-op; a runner error aborts before further
   writes.
6. `MacOS.Revert(ids)`: restores from the newest snapshot, removes ids from config.
7. `config`: `macos.tweaks` round-trips.
8. `cmd/macos.go`: `MacOS{Asker, Config, Registry, Apply, Stdout}` with spies:
   selection saved then applied; abort saves nothing; `list` output; `revert`
   flow. Composition roots + cobra wiring (`macos`, `m`; `list`/`l`; `revert`/`r`).
9. `setup`: on darwin the macOS asker runs after the tool picker with nothing
   pre-checked; `TARS_NO_FORM` selects nothing (spec with the OS injected).
10. Docs: README command table + "What it touches"; CLAUDE.md component list and
    "Adding a macOS tweak" note; this file trimmed to the user-facing parts.

## Verification

- `go test ./...`, lint, coverage ≥ 76 as usual; the `macos-latest` CI leg compiles
  and runs the darwin paths against the fake runner.
- Manual, on this Mac: `tars macos` → select `mission-control-arrows` only →
  `defaults read com.apple.symbolichotkeys` shows 34/35 disabled and a snapshot
  exists → `tars pull` twice adds no snapshot → `tars macos revert
  mission-control-arrows` re-enables them and the config no longer lists it.
- Manual, scratch user account: select everything, log out and in, check key
  repeat, Caps Lock acting as Control, F-keys, Finder, Dock, screenshots dir,
  trackpad. Then `tars macos revert` and confirm System Settings shows defaults.
