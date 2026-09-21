package profiles

import (
	"fmt"
	"path/filepath"
	"strings"
)

const managedHeader = "# managed by tars profiles — edit ~/.config/tars/profiles.yaml\n"

// BeginMarker and EndMarker delimit the block tars owns inside ~/.gitconfig.
const (
	BeginMarker = "# BEGIN tars profiles"
	EndMarker   = "# END tars profiles"
)

// GitconfigPath is where a profile's rendered include lives: <dir>/<alias>.gitconfig.
func GitconfigPath(dir, alias string) string {
	return filepath.Join(dir, alias+".gitconfig")
}

// RenderEnv seeds the profile's shell env file; the user owns it afterwards.
func RenderEnv(p Profile) string {
	return fmt.Sprintf(`# tars profile %q: sourced by the shell while active. Personal, never synced.
export TARS_PROFILE=%q
export GITHUB_USER=%q
# export GITHUB_TOKEN=""
`, p.Alias, p.Alias, p.GitHub)
}

// Splice replaces the managed block inside an existing gitconfig, or appends it when absent.
func Splice(existing []byte, block string) ([]byte, error) {
	s := string(existing)
	begin := strings.Index(s, BeginMarker)
	if begin < 0 {
		return appendBlock(s, block), nil
	}
	end := strings.Index(s[begin:], EndMarker)
	if end < 0 {
		return nil, fmt.Errorf("gitconfig has %q without %q; remove the partial block and re-run", BeginMarker, EndMarker)
	}
	end += begin + len(EndMarker)
	if end < len(s) && s[end] == '\n' {
		end++
	}
	return []byte(s[:begin] + block + s[end:]), nil
}

func appendBlock(s, block string) []byte {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return []byte(block)
	}
	return []byte(s + "\n\n" + block)
}

// RenderBlock is the managed ~/.gitconfig section: one dir-scoped include per profile.
func RenderBlock(f File, active, dir string) string {
	var b strings.Builder
	b.WriteString(BeginMarker + "\n")
	if p, found := f.Find(active); found {
		fmt.Fprintf(&b, "[include]\n\tpath = %s\n", GitconfigPath(dir, p.Alias))
	}
	for _, p := range f.Profiles {
		fmt.Fprintf(&b, "[includeIf \"gitdir:%s/\"]\n\tpath = %s\n", p.Dir(f.ProjectsDir), GitconfigPath(dir, p.Alias))
	}
	b.WriteString(EndMarker + "\n")
	return b.String()
}

// RenderGitconfig is the per-profile include: identity plus the key git must use.
func RenderGitconfig(p Profile, home string) string {
	var b strings.Builder
	b.WriteString(managedHeader)
	fmt.Fprintf(&b, "[user]\n\tname = %s\n\temail = %s\n", p.FullName, p.Email)
	fmt.Fprintf(&b, "[core]\n\tsshCommand = ssh -i %s -o IdentitiesOnly=yes\n", p.KeyPath(home))
	fmt.Fprintf(&b, "[github]\n\tuser = %s\n", p.GitHub)
	return b.String()
}
