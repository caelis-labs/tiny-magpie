package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/gateway"
)

// sameGroup compares two values of a field, one of them read back from the
// agent's own settings where Claude Code's [1m] mark rides on the group's
// id: the marked one is the group magpie set, not another group — the
// gateway serves either (#750's comparison, with the mark on one side).
func TestSameGroupTakesTheOneMMarkOff(t *testing.T) {
	for _, pair := range [][2]string{
		{"group/auto-gpt-5", "group/auto-gpt-5[1m]"},
		{"group/auto-gpt-5[1m]", "group/auto-gpt-5"},
		{"group/auto-gpt-5[1m]", "magpie/group/auto-gpt-5"},
	} {
		if !sameGroup(pair[0], pair[1]) {
			t.Errorf("sameGroup(%q, %q) = false", pair[0], pair[1])
		}
	}
	if sameGroup("group/auto-gpt-5", "group/auto-other") {
		t.Error("two different groups are the same")
	}
	if sameGroup("group/auto-gpt-5", "a/gpt-5") {
		t.Error("a group and a model are the same")
	}
}

// Something else rewrote Codex's config: the base URL gone while the model
// is still magpie's is unwired, and setting it again wires it back.
func TestCodexDriftUnwired(t *testing.T) {
	home, read := codexHome(t, `{"tokens":{"access_token":"x","id_token":"x.e30.x"}}`, "")
	cx := codex(home)
	if err := cx.Apply("model", "fake/m1"); err != nil {
		t.Fatal(err)
	}
	if d := cx.Drift(); d != nil {
		t.Fatalf("drift right after a set: %+v", d)
	}
	os.WriteFile(filepath.Join(home, ".codex", "config.toml"), []byte("model = \"fake/m1\"\n"), 0o644)
	d := cx.Drift()
	if d == nil || d.Kind != "unwired" || d.Want != "fake/m1" {
		t.Fatalf("unwired: %+v", d)
	}
	if err := cx.Reapply(); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); !strings.Contains(cfg, "openai_base_url") || cx.Drift() != nil {
		t.Fatalf("reapplied:\n%s", cfg)
	}
}

// A magpie model replaced by one of the agent's own is drift; one magpie
// model for another (the agent's own picker) isn't. Keep forgets it.
func TestDriftReplaced(t *testing.T) {
	home, read := codexHome(t, "", "")
	cx := codex(home)
	if err := cx.Apply("model", "fake/m1"); err != nil {
		t.Fatal(err)
	}
	// a switcher put Codex back on OpenAI, wholesale
	os.WriteFile(filepath.Join(home, ".codex", "config.toml"), []byte("model = \"gpt-5.5\"\nmodel_provider = \"openai\"\n"), 0o644)
	d := cx.Drift()
	if d == nil || d.Kind != "replaced" || d.Want != "fake/m1" || d.Now != "gpt-5.5" {
		t.Fatalf("replaced: %+v", d)
	}
	if err := cx.Reapply(); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); !strings.Contains(cfg, `model = "fake/m1"`) || !strings.Contains(cfg, `model_provider = "magpie"`) || cx.Drift() != nil {
		t.Fatalf("reapplied:\n%s", cfg)
	}
	os.WriteFile(filepath.Join(home, ".codex", "config.toml"), []byte("model = \"gpt-5.5\"\n"), 0o644)
	cx.Keep()
	if d := cx.Drift(); d != nil {
		t.Fatalf("kept: %+v", d)
	}
}

// A profile that sets its own provider wins over magpie's top level.
func TestCodexDriftProfile(t *testing.T) {
	home, read := codexHome(t, "", "")
	cx := codex(home)
	if err := cx.Apply("model", "fake/m1"); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(home, ".codex", "config.toml"), []byte("profile = \"work\"\n"+read()+"\n[profiles.work]\nmodel_provider = \"openai\"\n"), 0o644)
	if d := cx.Drift(); d == nil || d.Kind != "unwired" || !strings.Contains(d.Detail, "profile work") {
		t.Fatalf("profile: %+v\n%s", d, read())
	}
}

// Every agent magpie wires through the gateway says when its wiring was
// taken out behind magpie's back — the gateway's URL swapped for another —
// and setting it again puts it back.
func TestDriftUnwiredEveryAgent(t *testing.T) {
	home, _ := codexHome(t, "", "")
	managed := claudeManaged
	claudeManaged = func() string { return filepath.Join(home, "managed-settings.json") }
	t.Cleanup(func() { claudeManaged = managed })
	almaApp := startAlma(t) // Alma keeps its providers in the app
	// OpenHanako gives the model to an agent of its own
	os.MkdirAll(filepath.Join(home, ".hanako", "agents", "hana"), 0o755)
	os.WriteFile(filepath.Join(home, ".hanako", "agents", "hana", "config.yaml"), []byte("agent:\n  name: Hana\n"), 0o644)
	// AstrBot is connected once it has written its settings
	os.MkdirAll(filepath.Join(home, ".astrbot", "data"), 0o755)
	os.WriteFile(filepath.Join(home, ".astrbot", "data", "cmd_config.json"), []byte(astrbotConfig), 0o644)
	for _, a := range All() {
		// agy's gateway is in the command that starts it, in no file (its
		// Check is TestAgy's)
		if a.Check == nil || a.Launch != nil || a.Native != nil {
			continue
		}
		// a WSL agent's files live in the distro, beyond the sandbox home:
		// picking a model here writes the user's real configs
		if a.WSL != "" {
			continue
		}
		t.Run(a.ID, func(t *testing.T) {
			// the field that takes one of magpie's models
			var f Field
			want := ""
			for _, g := range a.Fields {
				for _, o := range g.Options(a.Values()) {
					if o.Ref == "fake/m1" && want == "" {
						f, want = g, o.Value
					}
				}
			}
			// or the one that puts it on magpie as a whole (Claude Desktop)
			for _, g := range a.Fields {
				for _, o := range g.Options(a.Values()) {
					if o.Value == magpieID && want == "" {
						f, want = g, o.Value
					}
				}
			}
			if want == "" {
				t.Fatal("no field takes magpie's models")
			}
			if err := a.Apply(f.Key, want); err != nil {
				t.Fatal(err)
			}
			if d := a.Drift(); d != nil {
				t.Fatalf("drift right after a set: %+v", d)
			}
			// something else points the agent at another server
			n := 0
			filepath.WalkDir(home, func(p string, e os.DirEntry, err error) error {
				if err != nil || e.IsDir() {
					return nil
				}
				b, _ := os.ReadFile(p)
				if s := strings.ReplaceAll(string(b), gateway.URL(), "http://127.0.0.1:9"); s != string(b) {
					os.WriteFile(p, []byte(s), 0o644)
					n++
				}
				return nil
			})
			n += almaApp.repoint(gateway.URL(), "http://127.0.0.1:9")
			if n == 0 {
				t.Fatal("no file holds the gateway's URL")
			}
			d := a.Drift()
			if d == nil || d.Kind != "unwired" {
				t.Fatalf("unwired: %+v", d)
			}
			if err := a.Reapply(); err != nil {
				t.Fatal(err)
			}
			if d := a.Drift(); d != nil {
				t.Fatalf("still drifted after Apply again: %+v", d)
			}
		})
	}
}

// Claude Code's managed settings win over the user's: a base URL there is
// drift magpie can only point out.
func TestClaudeDriftManaged(t *testing.T) {
	home, _ := codexHome(t, "", "")
	managed := claudeManaged
	claudeManaged = func() string { return filepath.Join(home, "managed-settings.json") }
	t.Cleanup(func() { claudeManaged = managed })
	cl := claude(home)
	if err := cl.Apply("model", "fake/m1"); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(claudeManaged(), []byte(`{"env":{"ANTHROPIC_BASE_URL":"https://corp.example"}}`), 0o644)
	if d := cl.Drift(); d == nil || d.Kind != "unwired" || !strings.Contains(d.Detail, "managed") {
		t.Fatalf("managed: %+v", d)
	}
}

// A magpie on a port of its own (magpie-dev) runs beside the installed one:
// an agent that one wired to its gateway is not drift here, while one wired
// to neither still is.
func TestDriftInstalledMagpie(t *testing.T) {
	t.Setenv("MAGPIE_ADDR", "127.0.0.1:3426")
	home, _ := codexHome(t, "", "")
	cx := codex(home)
	if err := cx.Apply("model", "fake/m1"); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(home, ".codex", "config.toml")
	os.WriteFile(cfg, []byte("openai_base_url = \"http://"+gateway.DefaultAddr+gateway.CodexPath+"\"\nmodel = \"fake/m1\"\n"), 0o644)
	if d := cx.Drift(); d != nil {
		t.Fatalf("the installed magpie's wiring taken for drift: %+v", d)
	}
	os.WriteFile(cfg, []byte("openai_base_url = \"http://127.0.0.1:9999"+gateway.CodexPath+"\"\nmodel = \"fake/m1\"\n"), 0o644)
	if d := cx.Drift(); d == nil || d.Kind != "unwired" {
		t.Fatalf("wired to neither: %+v", d)
	}
}
