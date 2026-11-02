package agent

// Sandbox support for Tiny boundary tests. Vendor requests and real Agent state are never used.
import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

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

	codexUsedUp = func() bool { return false }

	zedCredential = func(string) error { return nil }

	nodeModulesOf = func(string, []string) []string { return nil }

	windowsStorePackages = func() []string { return nil }

	windowsRegisteredApp = func(string) (string, string) { return "", "" }

	spotlightApps = func(string) []string { return nil }

	appQuit = func(context.Context, string) error { return errors.New("no app is quit in a test") }
	appOpen = func(context.Context, string) error { return errors.New("no app is opened in a test") }
	winProcs = func(string) []winProc { return nil }
	winClose = func([]int) error { return errors.New("no app is closed in a test") }
	winEnd = func(int) error { return errors.New("no app is ended in a test") }
	winLaunch = func(winProc) error { return errors.New("no app is opened in a test") }

	asideSet = func(string, string) error { return errors.New("aside: no Aside in a test") }
	asideRead = func(string) (map[string]json.RawMessage, error) { return nil, errors.New("aside: no Aside in a test") }

	reachProbe = func(string) bool { return true }

	codexProxyEnv = func(string) string { return "" }

	claudeShellEnv = func(string) string { return "" }
	codexSystemProxy = func() netproxy.Proxy { return netproxy.Proxy{} }
	proxyProbe = func(string, string) bool { return true }

	magpieAt = func(string) (bool, bool) { return false, true }
	os.Exit(testenv.RunIn(m, func(string) { testenv.Offline() }))
}
