# Compatibility and acceptance

This is an experimental reduction fork, not a new gateway implementation. The integrated upstream baseline is `d783df46cbac9021e26a6bd71ca0b55544552eb2`. Tests below use isolated directories, fake credentials and fixture vendors. No real Magpie profile, Agent configuration or subscription was used.

## Reproducible gates

| Gate | Method | Local evidence, 2026-10-18 |
| --- | --- | --- |
| Upstream baseline | Full upstream `go test -tags nogui ./...`; retry the three local-network failures outside the filesystem/network sandbox | Remaining three tests pass with local-network access; the initial sandbox failures are not product regressions |
| Tiny source and unit suite | `GOWORK=off go vet -tags nogui ./...`; `go test -tags nogui -timeout 15m ./...` | PASS: full Go suite (60 test packages) and vet; only inherited macOS SDK/linker warnings |
| Protocol differential | `scripts/verify-gateway.py` against separately built upstream and tiny | PASS: discovery and 64 comparisons: 4 inbound x 4 upstream protocols x streaming/non-streaming x text/tools-and-screenshot corpus; compare both vendor request bodies and client replies |
| Removed operations | Gateway/API negative tests, copied recording consent, imported nested intent groups, Provider edits against a sentinel Agent configuration | PASS: removed APIs unavailable, copied consent cannot archive, intent rejected, Agent file unchanged |
| Plugin ecosystem | Upstream Provider/middleware, mixed Provider hooks, async hook state/cancellation tests; tiny market/type checks | PASS in local retained-package run; real third-party releases and accounts remain untested |
| Browser GUI | `npm ci`; Playwright Chromium and WebKit, en/zh/zh-TW/ja/de, 1100/440 pixels; Web/light-theme port-conflict scenarios | PASS: 22 cases, no page errors, Providers default, retained pages accessible, removed controls remain hidden after feature refresh |
| macOS native | Build app; launch with `.tiny/native-profile`, port 13425; inspect actual accessibility tree/window | PASS: tiny bundle/window, Providers/API/subscription/local presets, Gateway, ordinary group editor and Settings navigation. No sign-in or vendor call |
| Headless platforms | `make release-cli` for darwin/linux/windows x amd64/arm64 | PASS: six targets compile |
| Windows GUI | `make release-windows` | PASS: amd64 and arm64 cross-build; not Windows runtime acceptance |
| Linux GUI compilation | Final source, CGO + GTK3/WebKit 4.1 in an Alpine arm64 container | PASS: production build and tiny GUI isolation tests; not a native window or release-package acceptance |
| Linux OS isolation | Container tests of XDG autostart and service helpers | PASS: tiny has independent names; enabling/disabling it preserves the original Magpie autostart entry |
| Linux container | Existing Dockerfile build and isolated HTTP/Web smoke | PASS: fresh volume and random loopback ports, health, named-key auth, discovery, model errors, authenticated Web and removed operations |
| Upstream update rehearsal | Apply reduction on baseline's parent; integrate baseline using `scripts/sync-upstream.sh`; compare result and rerun gates | PASS: `d694d9bf` to `d783df46`, one plugin-extension metadata conflict resolved by scope, resulting tree byte-identical to reduction `2c378a52`; build and 64-case differential pass |

The differential fixture returns text. Its second corpus exercises client instructions, tool calls/results, screenshots, Responses reasoning/include/max-output/service-tier fields, and verifies the exact request sent to each vendor. Tool-output events, reasoning signatures, opaque state, cancellation, retry boundaries and subscription lifecycle additionally rely on the retained upstream Go regressions. Generated response IDs/timestamps are normalized; stable content, terminal events and usage remain compared. This is evidence for the fixture corpus, not universal protocol equivalence or every vendor's live behavior.

The complete local baseline had three sandbox-dependent failures: `TestCallbackGoesPastTheAgentsProxy`, `TestClaudeToolCallbackOnANonLoopbackAddress`, and `TestLANModelListNotThroughProxy`. All three passed with local-network access and the same isolated fixtures. Run fixture tests with local listening/network access; do not use live credentials to bypass a test environment failure.

## Scope of removed tests

Tests exclusively accepting removed Agent management, cloud sync/archive, media/System One, intent classification, automatic image description, search execution and session pages leave the product suite with those features. Mixed tests retain their model/counting/routing/usage portions. Fixture config/cache/portable and OS identity expectations use tiny's independent paths. Negative tests replace the old acceptance where an imported setting or plugin could resurrect a removed operation. Removing an obsolete positive test is not evidence of runtime removal; the assembly guard and reachable API tests provide that evidence.

Coupled accounting and credential helpers retain upstream tests even where their former product UI is absent. The source tree still contains shared Agent/session/library types and helpers; this release does not claim their entire packages or the Wails, SQLite or Bun dependencies have disappeared. The CLI `sync` refreshes model catalogs; it is not cloud backup/sync.

## Before a stable release

- Run the committed CI matrix on macOS, Linux and Windows. Cross-builds and WebKit fixtures do not establish native Windows/Linux window, tray, WebView, permissions or installer behavior.
- Verify real accounts for at least one API-key/custom provider, the retained built-in subscription families and representative community Provider plugins. Exercise auth refresh, model discovery, quota failures, tools, screenshots, cancellation and concurrency. Use dedicated test accounts with explicit authorization.
- Run Caelis TUI, Bot with Caelis runtime, and Bot with Codex AppServer end to end. Existing Caelis Magpie discovery/Responses contracts can be reused; no client source was changed in this task.
- Complete platform packaging, signatures/notarization and installation/update checks. Tiny never self-updates from Magpie's application feed. Current workflows build artifacts only and do not publish a release or container registry image.

Task-local diagnostic logs are under ignored `.tiny/`: `baseline-tests.log`, `baseline-network-recheck.log`, `tests-final-head.log`, `usage-final.log`, `gui-isolation-tests.log`, `retained-tests-final.log`, `vet-final.log`, `differential-final.log`, `ui-tests-final.log`, `cross-final.log`, `windows-final.log`, `native-run-final.log`, and `docker-build.log`, `container-smoke.log`, `linux-isolation-tests.log`, `linux-gui-final.log` and `rehearsal-differential.log`. Browser screenshots are in ignored `output/playwright/`. CI uploads browser evidence. Logs and generated binaries are not committed.
