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

The initial local Go run used Bun 1.3.6, so tests requiring upstream's minimum Bun 1.4.2 were skipped. Release CI pins 1.4.2 and runs those host tests, including mixed-Provider tools and standalone tool/page rejection. An older-Bun local pass alone is not plugin-host acceptance.

CI pins compiler/module cache paths before test homes change and executes
uncached tests. The [test scope and CI matrix](test-scope.md) describe the
retained suites, removed positive tests, native Windows regression gate and
weekly full Windows run. PR, main and tag no longer repeat the same full
matrix three times. Native CI evidence belongs to its exact commit/tree;
the table above records the original local acceptance, not a current CI count.

## Scope of removed tests

Removed-product positive suites are physically deleted. Shared credential,
accounting, session data and storage helpers retain their coverage. Tiny
negative tests and assembly guards verify removed operations stay unreachable;
a deleted positive test alone does not prove runtime removal. See
[test-scope.md](test-scope.md) for the explicit boundary and CI execution.

The tree still contains coupled Agent/session/library helpers and Wails,
SQLite and Bun dependencies. CLI `sync` refreshes model catalogs; it is not
cloud backup/sync.

## Installer/update and identity revision, 2026-10-19

- Local retained Go suite: 52 test packages, 8,513 passing test/subtest events,
  no failures; Bun 1.3.6 does not satisfy the actual plugin-host minimum.
  Native and Linux/Windows vet pass. Exact-head CI repeats native gates with
  Bun 1.4.2; this local count does not replace it.
- Chromium/WebKit: 42 retained cases across all five languages and two widths,
  including own update preference persistence, explicit download, separate
  restart confirmation and navigation from the title-bar update badge.
- Protocol differential: discovery, 64 bidirectional fixture comparisons and
  removed-endpoint/config-write checks pass against the pinned upstream.
- Mounted DMG: two visible icons, real Applications symlink, minimal background,
  exact Finder positions and full app content comparison pass. The old public
  0.1.2 layout fails the new layout assertion; the old disabled updater fails
  `TestTinyUpdaterChecksWithoutDownloading` through a source overlay.
- Native signed staging: the public 0.1.2 DMG and installed-app fixture were
  already present locally. `TestTinySignedDMGFixture` served those bytes over
  loopback, checked their SHA-256, mounted/copied them, validated team, bundle
  ID, version and Gatekeeper, and replaced only a temporary copy. This does
  not prove a native Windows/Linux restart or an unattended real-user upgrade.
- Read-only native audit of the user's running Tiny instance informed the
  [feature checklist](feature-review-2026-10-19.md). It did not change accounts,
  settings or plugins, or issue model requests. Existing page/background
  quota reads could still occur. The user's live binary was not replaced.

## Before a stable release

- Run the committed CI matrix on macOS, Linux and Windows. Cross-builds and WebKit fixtures do not establish native Windows/Linux window, tray, WebView, permissions or installer behavior.
- Verify real accounts for at least one API-key/custom provider, the retained built-in subscription families and representative community Provider plugins. Exercise auth refresh, model discovery, quota failures, tools, screenshots, cancellation and concurrency. Use dedicated test accounts with explicit authorization.
- Run Caelis TUI, Bot with Caelis runtime, and Bot with Codex AppServer end to end. Existing Caelis Magpie discovery/Responses contracts can be reused; no client source was changed in this task.
- Verify clean-machine installation and any future Tiny update mechanism. Tiny never self-updates from Magpie's application feed. [The release workflow](releasing.md) gates experimental GitHub releases on exact-source CI, macOS signatures/notarization and asset readback; Windows/Linux remain unsigned. No container registry image is published.

Task-local diagnostic logs are under ignored `.tiny/`: `baseline-tests.log`, `baseline-network-recheck.log`, `tests-final-head.log`, `usage-final.log`, `gui-isolation-tests.log`, `retained-tests-final.log`, `vet-final.log`, `differential-final.log`, `ui-tests-final.log`, `cross-final.log`, `windows-final.log`, `native-run-final.log`, and `docker-build.log`, `container-smoke.log`, `linux-isolation-tests.log`, `linux-gui-final.log` and `rehearsal-differential.log`. Browser screenshots are in ignored `output/playwright/`. CI uploads browser evidence. Logs and generated binaries are not committed.
