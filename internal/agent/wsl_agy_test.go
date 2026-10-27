package agent

import (
	"path/filepath"
	"strings"
	"testing"
)

// mimi001234 on Discord: magpie's Agents page didn't find Antigravity CLI
// installed in WSL. The probe asks after its folder and its command, and
// finds it by either; a distro with neither has no agy row.
func TestWSLAgyFound(t *testing.T) {
	k := wslKindOf("agy")
	if !strings.Contains(wslProbeScript, `[ -d "$HOME/.gemini/antigravity-cli" ] && echo dir:.gemini/antigravity-cli; `) ||
		!strings.Contains(wslProbeScript, `command -v agy `) {
		t.Fatal("the probe doesn't ask after agy")
	}
	for _, p := range []string{"dir:.gemini/antigravity-cli\n", "bin:agy /home/me/.local/bin/agy\n"} {
		if d := parseProbe("U", "home:/home/me\n"+p); d == nil || !k.found(*d) {
			t.Errorf("not found by %q", p)
		}
	}
	// Gemini CLI's ~/.gemini, or Windows' agy on the PATH WSL inherits, is not it
	for _, p := range []string{"dir:.gemini\n", "bin:gemini /usr/bin/gemini\n", "bin:agy /mnt/c/Users/me/AppData/Local/Programs/agy/agy\n"} {
		if d := parseProbe("U", "home:/home/me\n"+p); d != nil && k.found(*d) {
			t.Errorf("found by %q", p)
		}
	}

	syncHome(t)
	with, without := t.TempDir(), t.TempDir()
	fakeWSL(t, "Ubuntu\r\nDebian\r\n", "Ubuntu\r\nDebian\r\n", map[string]string{
		"Ubuntu": "home:/home/me\nbin:agy /home/me/.local/bin/agy\ndir:.gemini/antigravity-cli\nroute:default via 172.20.0.1 dev eth0\n",
		"Debian": "home:/home/me\ndir:.codex\nroute:default via 172.20.0.1 dev eth0\n",
	}, map[string]string{"Ubuntu": with, "Debian": without})
	var ids []string
	for _, a := range wslAgentsOf(wslDistros()) {
		ids = append(ids, a.ID)
		if a.ID == "agy@wsl:Ubuntu" && (a.Name != "Antigravity CLI · WSL Ubuntu" || a.Bin != "" ||
			a.Path != filepath.Join(with, "home", "me", ".gemini", "antigravity-cli", "settings.json")) {
			t.Errorf("%+v", a)
		}
	}
	if got := strings.Join(ids, ","); got != "agy@wsl:Ubuntu,codex@wsl:Debian" {
		t.Fatalf("agents %s", got)
	}
}
