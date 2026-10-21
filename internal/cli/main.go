// Package cli is magpie's command line: every command, and the app, tray
// and TUI it opens (the root's main.go only calls Main).
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/yetone/magpie/internal/agent"
	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/claudebridge"
	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/netproxy"
	"github.com/yetone/magpie/internal/pluginrt"
	"github.com/yetone/magpie/internal/proc"
	"github.com/yetone/magpie/internal/profile"
	"github.com/yetone/magpie/internal/provider"

	"github.com/yetone/magpie/internal/update"
)

// version is the root main package's, set at build time
// (-ldflags "-X main.version=…") and handed over by Main.
var version = "dev"

const usage = `tiny-magpie — multi-provider model gateway

  tiny-magpie [app|tray|panel]       desktop gateway and provider management
  tiny-magpie web [--no-open]       shared Web GUI (no desktop required)
  tiny-magpie serve [--web]         HTTP gateway, optionally with Web GUI
  tiny-magpie providers|presets|models
  tiny-magpie provider add|set|key|models|test|rm ...
  tiny-magpie groups|group add|set|rm ...
  tiny-magpie accounts [add|refresh|discovery|alias|import] ...
  tiny-magpie plugin [add|rm|update|on|off|login|logout] ...
  tiny-magpie quota [--json]
  tiny-magpie gateway-key list|add|rotate|remove|limit|models ...
  tiny-magpie claude-code [install|remove]
  tiny-magpie sync|import|healthcheck|heap|cpu|version

Model APIs: Chat Completions, Responses, Anthropic Messages, Gemini,
Embeddings and Rerank. Existing MAGPIE_* gateway options are supported.
Agent model switching is available in the GUI. Intent routing, media tools and cloud backup/sync are not included.
`

var (
	bold  = lipgloss.NewStyle().Bold(true)
	muted = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8B8F98", Dark: "#7C8290"})
	faint = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C4C7CE", Dark: "#4A4F5A"})
	green = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0F9D58", Dark: "#7EE2A8"})
)

// Main runs magpie with os.Args, at the version the build was stamped with.
func Main(v string) {
	version = v
	// before anything reads or writes a file: no home, or a relative one,
	// would put the agents' configs and magpie's keys under the working
	// folder. What needs no file still answers (magpie version in a
	// container or a script without HOME): run makes magpie's folders first.
	ignored, err := appdir.CheckEnv()
	if err != nil {
		if len(os.Args) > 1 {
			switch os.Args[1] {
			case "-v", "--version", "version":
				fmt.Println("magpie", version)
				return
			case "-h", "--help", "help":
				fmt.Print(usage)
				return
			}
		}
		fmt.Fprintln(os.Stderr, "magpie:", err)
		os.Exit(1)
	}
	slices.Sort(ignored)
	for _, v := range ignored {
		fmt.Fprintf(os.Stderr, "magpie: ignoring %s: not an absolute path (the programs magpie starts still get it)\n", v)
	}
	if provider.TookOpenedURL(os.Args[1:]) {
		// Claude Code, signing in for magpie, handed over the page to open
		return
	}
	endProbesOnSignal()
	gateway.Version = version
	netproxy.Install()
	update.GUI = hasGUI
	err = run(os.Args[1:])
	proc.EndProbes() // a CLI still being asked something isn't left to init
	if err != nil {
		// a command with exit codes of its own (quota wait) says which
		code := 1
		var e exitError
		if errors.As(err, &e) {
			code = e.code
		}
		if msg := err.Error(); msg != "" {
			fmt.Fprintln(os.Stderr, "magpie:", msg)
		}
		os.Exit(code)
	}
}

func run(args []string) error {
	if len(args) == 2 && args[0] == agent.DryRunArg {
		// Upstream disconnect preview executes against a copied temporary home.
		return agent.DryRun(args[1])
	}

	if len(args) > 0 {
		switch args[0] {
		case "app", "gui", "tray", "panel", "web", "serve", "healthcheck", "heap", "cpu", "providers", "presets", "provider", "models", "model", "groups", "group", "gateway-key", "accounts", "account", "quota", "quotas", "plugin", "plugins", "sync", "import", "claude-code", "claude-mcp-helper", "-Embedding", "version", "--version", "-v", "help", "--help", "-h":
		default:
			return fmt.Errorf("unknown tiny-magpie command %q; run tiny-magpie --help", args[0])
		}
		if helpAsked(args) {
			fmt.Print(helpFor(args))
			return nil
		}
		if (args[0] == "accounts" || args[0] == "account") && len(args) > 1 {
			switch args[1] {
			case "switch", "switch-only", "checkin", "warmup":
				return fmt.Errorf("account maintenance %q is not available in tiny-magpie", args[1])
			}
		}
		if args[0] == "models" && len(args) > 1 {
			return fmt.Errorf("tiny-magpie models lists gateway models; agent configuration is not managed")
		}
	}

	if len(args) > 0 {
		switch args[0] {
		case "version", "--version", "-v":
			fmt.Println("tiny-magpie", version)
			return nil
		case "healthcheck":
			return healthcheck()
		case "heap":
			return heapCmd(args[1:])
		case "cpu":
			return cpuCmd(args[1:])
		}
	}
	makeDirs()
	if _, err := pluginrt.Settle(); err != nil {
		return err
	}
	if _, err := pluginrt.Refresh(); err != nil {
		return err
	}
	// Keep upstream model lists current for agents connected through the GUI.
	catalog.Changed = agent.SyncCatalog
	gateway.StandIn = agent.StandIn
	gateway.CodexOwnOnly = agent.CodexOwnOnly
	if len(args) == 0 {
		if hasGUI {
			return runGUI(true, "")
		}
		return serve()
	}
	// how this process was started, for the Routing page of another magpie on
	// this computer to name the one that serves the gateway (#2606)
	switch args[0] {
	case "web", "serve", "tray":
		gateway.Serves = args[0]
	case "app", "gui", "panel", "-Embedding":
		gateway.Serves = "app"
	}
	switch args[0] {
	case "app", "gui":
		return runWindow("providers")
	case "tray":
		return runGUI(false, "")
	case "panel":
		return runPanel()
	case "web":
		return webCmd(args[1:])
	case "serve":
		return serveCmd(args[1:])
	case "providers":
		return providers()
	case "presets":
		return presets()
	case "provider":
		return providerCmd(args)
	case "models":
		return models(nil)
	case "model":
		return modelCmd(args[1:])
	case "groups":
		return groups()
	case "group":
		return groupCmd(args)
	case "gateway-key":
		return gatewayKeys(args)
	case "accounts", "account":
		return accountsCmd(args)
	case "quota", "quotas":
		return quotaCmd(args)
	case "plugin", "plugins":
		return pluginCmd(args)
	case "claude-code":
		return claudeCodeCmd(args[1:])
	case "claude-mcp-helper":
		return claudebridge.RunMCP(args[1:])
	case "import":
		return importCmd(args[1:])
	case "sync":
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := catalog.Sync(ctx); err != nil {
			return err
		}
		refreshLive(ctx)
		return nil
	}
	return nil
}

func set(a *agent.Agent, key, value string) error {
	f := a.Field(key)
	if f == nil {
		var keys []string
		for _, f := range a.Fields {
			keys = append(keys, f.Key)
		}
		return fmt.Errorf("%s has no field %q (fields: %s)", a.Name, key, strings.Join(keys, ", "))
	}
	// a field's name is no value: magpie codex subagent effort ran Codex's
	// subagents on a model called effort. Claude Code's opus is a model's
	// name as well as a tier's
	if g := a.Field(value); value != "" && g != nil && !modelAlias(a, value) {
		return fieldAsValue(a, f, g, value)
	}
	value, err := a.Spell(f.Key, value)
	if err != nil {
		return err
	}
	before := f.Get()
	wired := a.Wired()
	// a magpie model picked on an agent not connected yet connects it
	// first, as the Agents page's picker does (#1900): the agent is wired
	// in, and Disconnect puts back what it had before
	if err := a.Pick(f.Key, value); err != nil {
		return err
	}
	// what the config reads now, not what was asked: an agent may name the
	// model under a provider of its own (OpenCode's magpie-relay/…), and a
	// value it had already is said to be so
	now := f.Get()
	shown := now
	if value == "" || now == "" {
		shown = muted.Render("default")
	}
	if now == before {
		shown += " " + muted.Render("(unchanged)")
	}
	fmt.Println(green.Render("✓"), bold.Render(a.Name), muted.Render(f.Label), shown)
	// a value that is the agent's own, not one of magpie's, takes magpie
	// out: said, as Disconnect is, with the way back (#2000: magpie claude
	// sonnet put Claude Code back on its DeepSeek endpoint, which refused
	// Sonnet, and only "✓ Claude Code model sonnet" was printed)
	if wired && now != "" && !a.Wired() {
		fmt.Println(muted.Render(fmt.Sprintf("  disconnected from magpie: %s is %s's own, sent to the endpoint it had before magpie; magpie %s <model> with one of magpie's models connects it again", now, a.Name, a.ID)))
	}
	if a.Notice != nil && (!f.Live || !wired && a.Wired()) {
		if n := a.Notice(); n != "" {
			fmt.Println(muted.Render("  ↻ " + n))
		}
	}
	return nil
}

// modelAlias says whether a word names the agent's model, in any case,
// though a field goes by it too: Claude Code's opus, a tier's name as well
func modelAlias(a *agent.Agent, w string) bool {
	return slices.ContainsFunc(a.ModelAliases, func(m string) bool { return strings.EqualFold(m, w) })
}

// fieldAsValue refuses field g's name given as field f's value, for what
// was meant: two words that are one field's label (magpie codex subagent
// effort), a switch, which is turned on (Claude Code's ultracode), or the
// field, which its name alone shows.
func fieldAsValue(a *agent.Agent, f, g *agent.Field, value string) error {
	for _, name := range []string{f.Key + " " + value, f.Key + "_" + value} {
		if h := a.Field(name); h != nil {
			return fmt.Errorf("%q is one field of %s's: magpie %s %s shows it", h.Label, a.Name, a.ID, h.Key)
		}
	}
	if g.Options != nil {
		if o := g.Options(a.Values()); len(o) == 1 && o[0].Value == "on" {
			return fmt.Errorf("%q is a switch of %s's, not a value for its %s: magpie %s %s on", value, a.Name, f.Label, a.ID, g.Key)
		}
	}
	return fmt.Errorf("%q is a field of %s's, not a value for its %s: magpie %s %s shows it", value, a.Name, f.Label, a.ID, g.Key)
}

// showField is `magpie <agent> <field>`: what the field is set to, as set
// prints it
func showField(a *agent.Agent, f *agent.Field) error {
	v := f.Get()
	if v == "" {
		v = muted.Render("default")
	}
	fmt.Println(bold.Render(a.Name), muted.Render(f.Label), v)
	return nil
}

// disconnect is `magpie <agent> default` on an agent magpie is wired into:
// the Agents page's Disconnect, which puts back what the user had before
// magpie — Claude Code's own model, its endpoint — where a field's default
// leaves the agent as installed (__jingling on X: magpie claude default
// took the model they had set away with magpie's)
func disconnect(a *agent.Agent) error {
	before := a.Values()
	if err := a.Disconnect(); err != nil {
		return err
	}
	now := a.Values()
	for _, f := range a.Fields {
		if before[f.Key] == now[f.Key] {
			continue
		}
		shown := now[f.Key]
		if shown == "" {
			shown = muted.Render("default")
		}
		fmt.Println(green.Render("✓"), bold.Render(a.Name), muted.Render(f.Label), shown)
	}
	fmt.Println(muted.Render("  disconnected from magpie, back to what it had before"))
	if a.Notice != nil {
		if n := a.Notice(); n != "" {
			fmt.Println(muted.Render("  ↻ " + n))
		}
	}
	return nil
}

func fieldForValue(a *agent.Agent, v string) *agent.Field {
	vals := a.Values()
	// the agent's suffix after a model (omp's ":max") aside: a role on the
	// same model at that level offers it as typed, and would take it. A list
	// of models no picker offers: it is for the model
	if a.SplitSuffix != nil {
		m, _, one := a.SplitSuffix(v)
		if !one {
			return nil
		}
		v = m
	}
	// a model stays with the model, even where other fields offer it too
	// (Claude Code's opus/sonnet/haiku/fable)
	for _, o := range a.Fields[0].Options(vals) {
		if o.Value == v {
			return nil
		}
	}
	for i := 1; i < len(a.Fields); i++ {
		for _, o := range a.Fields[i].Options(vals) {
			if o.Value == v {
				return &a.Fields[i]
			}
		}
	}
	return nil
}

// list prints the agents; those from dimFrom on (when not -1) are the ones
// hidden in the app, and are dimmed.
func list(agents []*agent.Agent, detectedOnly bool, dimFrom int) error {
	if len(agents) == 0 {
		return fmt.Errorf("no supported agents found on this machine")
	}
	type row struct{ name, vals, path string }
	var rows []row
	nameW, valW := 0, 0
	for i, a := range agents {
		r := row{name: a.Name, path: tilde(a.Path)}
		if !detectedOnly && !a.Detected() {
			r.name = faint.Render(a.Name)
			r.vals = faint.Render("not detected")
			r.path = ""
		} else if a.CLIMissing() {
			// its folder left by an uninstall (#2173): what it holds is
			// read by nothing
			r.name = faint.Render(a.Name)
			r.vals = faint.Render("CLI not found")
		} else {
			vals := a.Values()
			dim := dimFrom >= 0 && i >= dimFrom
			label, value := muted, lipgloss.NewStyle()
			if dim {
				label, value = faint, faint
			}
			var parts []string
			for _, f := range a.Fields {
				v := vals[f.Key]
				if v == "" && f.Quiet {
					continue
				}
				if v == "" {
					v = faint.Render("—")
				} else {
					v = value.Render(v)
				}
				if f.Label == "model" {
					parts = append(parts, v)
				} else {
					parts = append(parts, label.Render(f.Label)+" "+v)
				}
			}
			r.name = bold.Render(a.Name)
			if dim {
				r.name = faint.Render(a.Name) + " " + faint.Render("hidden")
			}
			r.vals = strings.Join(parts, label.Render("  ·  "))
			// through magpie for account failover while not connected (#1385)
			if said := a.FailoverSaid(); said != "" {
				r.vals += label.Render("  ·  " + said)
			}
			if a.Import != nil {
				if a.Added != nil && a.Added() {
					r.vals = value.Render("magpie added")
				} else {
					r.vals = label.Render("magpie "+a.ID+" add") + faint.Render("  to add magpie")
				}
			}
		}
		nameW = max(nameW, lipgloss.Width(r.name))
		valW = max(valW, lipgloss.Width(r.vals))
		rows = append(rows, r)
	}
	for _, r := range rows {
		fmt.Printf("  %s  %s  %s\n", pad(r.name, nameW), pad(r.vals, valW), faint.Render(r.path))
	}
	return nil
}

func tilde(p string) string {
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	return p
}

func profiles(args []string) error {
	switch args[0] {
	case "profiles":
		ps, err := profile.Load()
		if err != nil {
			return err
		}
		if len(ps) == 0 {
			fmt.Println(muted.Render("no profiles yet · magpie save <name>"))
			return nil
		}
		detected := agent.Detected()
		for _, n := range profile.Names(ps) {
			// ✓ the one the setup is now (superhtoooth on X)
			mark := " "
			if profile.InUse(ps[n], detected, nil) {
				mark = green.Render("✓")
			}
			fmt.Printf("%s %s  %s\n", mark, bold.Render(n), muted.Render(profile.LongSummary(ps[n])))
		}
		return nil
	case "save":
		if len(args) < 2 {
			return fmt.Errorf("usage: magpie save <name>")
		}
		p, err := profile.Snapshot()
		if err != nil {
			return err
		}
		if err := profile.Save(args[1], p); err != nil {
			return err
		}
		fmt.Println(green.Render("✓"), "saved profile", bold.Render(args[1]))
		if p.Library != nil {
			fmt.Println(" ", muted.Render("with the "+p.Library.Summary()))
		}
		return nil
	case "use":
		if len(args) < 2 {
			return fmt.Errorf("usage: magpie use <name>")
		}
		ps, err := profile.Load()
		if err != nil {
			return err
		}
		p, ok := ps[args[1]]
		if !ok {
			return fmt.Errorf("no profile named %q", args[1])
		}
		a, err := profile.Apply(p)
		if err != nil {
			return err
		}
		fmt.Println(green.Render("✓"), "applied", bold.Render(args[1]), muted.Render(fmt.Sprintf("(%d changed)", a.Changed)))
		for _, line := range profile.Report(a) {
			fmt.Println(" ", muted.Render(line))
		}
		return nil
	case "rm":
		if len(args) < 2 {
			return fmt.Errorf("usage: magpie rm <name>")
		}
		if err := profile.Delete(args[1]); err != nil {
			return err
		}
		fmt.Println(green.Render("✓"), "deleted profile", bold.Render(args[1]))
		return nil
	}
	return nil
}

func pad(s string, w int) string {
	if n := w - lipgloss.Width(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}
