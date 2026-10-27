package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
"path/filepath"
"strings"
"github.com/yetone/magpie/internal/provider"
	"testing"
"sync"
"time"
"unicode/utf16"

	"github.com/yetone/magpie/internal/netproxy"
	"github.com/yetone/magpie/internal/testenv"
)

// TestMain gives the package a home of its own (testenv). A home an agent is
// found through can sit outside HOME — DSH_HOME does — and the package's
// tests sandbox HOME alone: left as the developer has it, a test that picks
// a model writes it into the real ~/.dsh/profiles, and dsh then refuses
// every turn with a model its provider does not list. internal/gateway,
// internal/provider and internal/usage isolate themselves the same way.
func TestMain(m *testing.M) {
	if len(os.Args) == 3 && os.Args[1] == DryRunArg {
		if err := DryRun(os.Args[2]); err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}
	// whether Codex's ChatGPT account is out of its allowance is asked of
	// OpenAI; never from here
	codexUsedUp = func() bool { return false }
	// Sandboxed config writes must never write into the real OS keychain.
	zedCredential = func(string) error { return nil }
	// Pi's and omp's model registries are read from their installs; never
	// this machine's
	nodeModulesOf = func(string, []string) []string { return nil }
	// Microsoft Store apps are read from the current user's AppX registry;
	// never this machine's (a Windows box with the Codex app installed)
	windowsStorePackages = func() []string { return nil }
	// nor what an installer registered there (DeepSeek Harness Desktop's
	// folder, #2279)
	windowsRegisteredApp = func(string) (string, string) { return "", "" }
	// nor where this Mac's Spotlight has an app (DeepSeek Harness.app
	// outside /Applications, #2325)
	spotlightApps = func(string) []string { return nil }
	// a desktop app's Restart never quits or opens one of this machine's
	// (the Codex app, Claude Desktop, WorkBuddy the developer runs); the
	// restart tests stand in for each step
	appQuit = func(context.Context, string) error { return errors.New("no app is quit in a test") }
	appOpen = func(context.Context, string) error { return errors.New("no app is opened in a test") }
	winProcs = func(string) []winProc { return nil }
	winClose = func([]int) error { return errors.New("no app is closed in a test") }
	winEnd = func(int) error { return errors.New("no app is ended in a test") }
	winLaunch = func(winProc) error { return errors.New("no app is opened in a test") }
	// Aside's default is changed through the aside binary, which answers for
	// the account the developer is actually running: a test that picks a
	// model would change the real Aside's default, and the running one would
	// keep the old one, so the test would read back a file nothing had
	// applied. The Aside tests stand in for it themselves.
	asideSet = func(string, string) error { return errors.New("aside: no Aside in a test") }
	asideRead = func(string) (map[string]json.RawMessage, error) { return nil, errors.New("aside: no Aside in a test") }
	// an address an agent is pointed at is tried over the network (Drift,
	// #1013); never from a test, whose reach tests stand in for it
	reachProbe = func(string) bool { return true }
	// nor is the way Codex goes there read from this machine's proxy
	// settings, or a proxy tried; the proxy tests say both
	codexProxyEnv = func(string) string { return "" }
	// nor Claude Code's provider switches read from the login shell; the
	// cloud tests say what one there does
	claudeShellEnv = func(string) string { return "" }
	codexSystemProxy = func() netproxy.Proxy { return netproxy.Proxy{} }
	proxyProbe = func(string, string) bool { return true }
	// nor an address of the agent's own asked whether magpie answers
	// there (a reverse proxy, #1985): known, and not magpie, as for a
	// relay; the proxy-kept tests answer it
	magpieAt = func(string) (bool, bool) { return false, true }
	os.Exit(testenv.RunIn(m, func(string) { testenv.Offline() }))
}

// Upstream shared fixtures for retained detection/model-wiring tests.
func dirNames(t *testing.T, dir string) string {
	t.Helper()
	es, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range es {
		names = append(names, e.Name())
	}
	return strings.Join(names, ",")
}

func installApp(t *testing.T, dir, name, plist string) {
	t.Helper()
	c := filepath.Join(dir, name, "Contents")
	if err := os.MkdirAll(c, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(c, "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatal(err)
	}
}

func dshRouteHome(t *testing.T) (home, dir, web string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("DSH_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	if err := provider.Save(provider.Provider{ID: "deepseek", Name: "DeepSeek", Chat: "https://api.deepseek.com/v1", Key: "k", Models: []string{"pro", "flash"}}); err != nil {
		t.Fatal(err)
	}
	dir = filepath.Join(home, ".dsh")
	web = filepath.Join(dir, "profiles", "web", "cordis.patch.yml")
	os.MkdirAll(filepath.Dir(web), 0o755)
	return home, dir, web
}

func writeT(t *testing.T, p, body string) {
	t.Helper()
	mkdirT(t, filepath.Dir(p))
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func judgeFor(t *testing.T, goos string) {
	t.Helper()
	was := goneOS
	goneOS = goos
	t.Cleanup(func() { goneOS = was })
}
func mkdirT(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func fakeWSL(t *testing.T, installed, running string, probes, roots map[string]string) (asked, opened *[]string) {
	t.Helper()
	run, root := wslRun, wslRoot
	reset := func() {
		wsl.seen, wsl.at, wsl.runAt, wsl.names, wsl.running, wsl.dirty = nil, time.Time{}, time.Time{}, nil, nil, false
		wsl.runOK = false
	}
	t.Cleanup(func() { wslRun, wslRoot = run, root; reset() })
	reset()
	asked, opened = &[]string{}, &[]string{}
	wslRun = func(_ time.Duration, args ...string) ([]byte, error) {
		switch strings.Join(args, " ") {
		case "-l -q":
			return utf16le(installed, true), nil
		case "-l --running -q":
			return utf16le(running, true), nil
		}
		if len(args) > 2 && args[0] == "-d" {
			*asked = append(*asked, args[1])
			return []byte(probes[args[1]]), nil
		}
		return nil, errors.New("unexpected wsl.exe " + strings.Join(args, " "))
	}
	wslRoot = func(name string) string {
		*opened = append(*opened, name)
		return roots[name]
	}
	return asked, opened
}

func fakeDistro(home string, mirrored bool) distro {
	return distro{Name: "Ubuntu-24.04", Root: filepath.Dir(home), Home: "/" + filepath.Base(home),
		Has: map[string]bool{"dir:.codex": true}, Gateway: "172.20.0.1", Mirrored: mirrored, Running: true}
}

func utf16le(s string, bom bool) []byte {
	var b []byte
	if bom {
		b = append(b, 0xff, 0xfe)
	}
	for _, u := range utf16.Encode([]rune(s)) {
		b = append(b, byte(u), byte(u>>8))
	}
	return b
}

func fakeReach(t *testing.T, answers map[string]bool) *sync.Mutex {
	t.Helper()
	var mu sync.Mutex
	was, age := reachProbe, reachAge
	reachProbe = func(base string) bool { mu.Lock(); defer mu.Unlock(); return answers[base] }
	reachAge = time.Hour
	reach.Lock()
	reach.m = nil
	reach.Unlock()
	t.Cleanup(func() {
		reachProbe, reachAge = was, age
		reach.Lock()
		reach.m = nil
		reach.Unlock()
	})
	return &mu
}

func reachForget() {
	reach.Lock()
	reach.m = nil
	reach.Unlock()
}

func driftSettled(t *testing.T, a *Agent) *Drift {
	t.Helper()
	a.Drift()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(5 * time.Millisecond) {
		reach.Lock()
		busy := false
		for _, r := range reach.m {
			busy = busy || r.busy
		}
		reach.Unlock()
		if !busy {
			return a.Drift()
		}
	}
	t.Fatal("the probes never answered")
	return nil
}
