# Compatibility and acceptance

This is an experimental reduction fork, not a new gateway implementation. The integrated upstream baseline is `ba319a7587923fc7a4ea3859658244d591261a59` (see `UPSTREAM.json`). Earlier dated gates below describe their original revision. Automated tests below use isolated directories, fake credentials and fixture vendors. Native observations, exceptions and their limits are recorded in each revision.

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

## Gateway-core Dev revision, 2026-10-19

This local revision remains on `codex/gateway-core-dev`; it is not a release.
The integrated upstream range has not changed. Direct System One routing and
its decision discovery header/snippets come from that baseline. Embeddings and
Rerank already had upstream server endpoints; Tiny adds only typed inventory
and GUI examples around the inherited forwarding/accounting implementation.
Intelligent/intent routing remains unavailable.

- Local protocol differential: discovery and 67 bidirectional fixture comparisons
  pass, including explicit System One, Embedding and Rerank requests/replies.
- Retained Go suite: 51 test packages pass with Bun 1.4.2, including the actual plugin-host tests.
  Focused race checks cover direct classification, raw request forwarding,
  no full-body retention, fixed assembly and unavailable retired APIs.
- Chromium/WebKit: 70 cases across five languages and wide/narrow viewports, including
  visible token/cost/speed/TTFT summaries and tray/header icon bounds at zoom.
  Native usage summaries reuse the upstream container breakpoints so speed
  values and their units remain readable in the six-metric overview.
- Removing the System One handlers through a Go source overlay makes
  `TestTinySystemOnePassesThrough` fail with HTTP 404; the restored source
  passes. Removing the icon bounds makes all four tray regression cases fail.
- Native macOS Dev uses separate bundle identity, fixed environment and profile,
  gateway 13426, Web UI 13427 and a local vendor fixture on 13430. Four calls
  through this isolated gateway returned 200 and generated fake usage records.
  Known demo prices and unknown prices remain distinguishable in the real UI.

An initial macOS launcher did not inherit the isolation environment, so its
read-only first window showed existing account state. It was stopped before
UI actions or model calls; background startup reads cannot be excluded. The
replacement Dev launcher uses a fixed `execve` environment. Its actual process
path/environment, API data directory/version and two fixture-only Provider IDs
were independently read back, including a normal LaunchServices launch. No live
profile was intentionally edited or used for acceptance. No real vendor account
or current third-party release was exercised; native Windows/Linux GUI and
installer acceptance remain outstanding.

The first release review exposed a Dev-only menu-bar hosting failure in that
exec launcher (macOS MenuBarAgent rejected its scene). A standard native bundle
entry displayed the icon and was confirmed by the maintainer. The final local
Dev uses a native Go entry with task-local isolation applied before CLI startup;
the overlay and fixed test paths are not release source. Formal packages already
use the native Go entry. No production tray implementation or system permission
was changed. Final browser checks cover 70 cases, including available allowance
cards and the corrected subscription-card/tray-value bounds. Container smoke now
requires retained System One validation/auth instead of the obsolete 404.

The [macOS permission investigation](macos-permissions.md) verifies client tool
handoff and removed management endpoints with fixtures; it does not reproduce
the original TCC dialog. Tiny's own administrator update prompt now names Tiny.

See [Dev acceptance](dev-gateway-core.md). Task-local evidence is under ignored
`.tiny/core-dev*`, including the full suite, cross-platform vet/build, protocol,
race, browser/mutation logs, isolation proof and native screenshots. Generated
Dev apps, launchers, fixture profiles and credentials are not source artifacts.

## Before a stable release

- Run the committed CI matrix on macOS, Linux and Windows. Cross-builds and WebKit fixtures do not establish native Windows/Linux window, tray, WebView, permissions or installer behavior.
- Verify real accounts for at least one API-key/custom provider, the retained built-in subscription families and representative community Provider plugins. Exercise auth refresh, model discovery, quota failures, tools, screenshots, cancellation and concurrency. Use dedicated test accounts with explicit authorization.
- Run Caelis TUI, Bot with Caelis runtime, and Bot with Codex AppServer end to end. Existing Caelis Magpie discovery/Responses contracts can be reused; no client source was changed in this task.
- Verify clean-machine installation and any future Tiny update mechanism. Tiny never self-updates from Magpie's application feed. [The release workflow](releasing.md) gates experimental GitHub releases on exact-source CI, macOS signatures/notarization and asset readback; Windows/Linux remain unsigned. No container registry image is published.

Task-local diagnostic logs are under ignored `.tiny/`: `baseline-tests.log`, `baseline-network-recheck.log`, `tests-final-head.log`, `usage-final.log`, `gui-isolation-tests.log`, `retained-tests-final.log`, `vet-final.log`, `differential-final.log`, `ui-tests-final.log`, `cross-final.log`, `windows-final.log`, `native-run-final.log`, and `docker-build.log`, `container-smoke.log`, `linux-isolation-tests.log`, `linux-gui-final.log` and `rehearsal-differential.log`. Browser screenshots are in ignored `output/playwright/`. CI uploads browser evidence. Logs and generated binaries are not committed.

## Upstream sync and reduced Agent candidate, 2026-10-21

Integrated upstream `ba319a7587923fc7a4ea3859658244d591261a59`, the complete
203-commit interval after `d783df46cbac9021e26a6bd71ca0b55544552eb2` (#1).
The merge retains Tiny identity, independent gateway ownership and retired
products, including upstream's newly added privacy-model product.

Agent wiring uses upstream detection, Pick/ConnectHow, drift, model visibility
and disconnect/restore. Tiny filters model-write fields on both API and UI;
installation/updates, Profiles, sign-in, memories and subagent orchestration are
not exposed. The private disconnect-preview subprocess entry is restored,
and its fixture verifies the original configuration is unchanged by preview.
The restored Agent suites use upstream fixtures. The two out-of-scope
log-based drift tests are removed; catalog and disconnect fixtures establish
Tiny ownership explicitly and retain their original preservation assertions.
Model writers remain upstream-derived.

The pinned upstream/Tiny binary differential passes discovery, 67 protocol
comparisons (HTTP/SSE, tools/screenshots, System One and retrieval APIs), and
removed endpoint/config-write checks. Chromium and WebKit pass 84 Tiny UI
cases across five languages and narrow/wide layouts. CUA additionally inspected
and switched a model in a synthetic local Agent fixture, and inspected the
folded group editor. These are fixture results, not live-account acceptance.

macOS native compilation is checked without replacing or launching either
installed app. Windows/Linux native GUI, signing/notarization, clean-machine
installation and real Agent/account use remain unverified in this candidate.

All 55 retained Go test packages pass across the complete package run plus
its corrected gateway rerun (282s). The first run stopped gateway compilation
on a stale test-only import; the corrected package was rerun in full. The CLI
package is rerun after restoring the preview subprocess dispatch. Focused
Agent/GUI/CLI race checks and vet also pass. Local native builds retain the
upstream macOS SDK/deployment-target warnings; no signing or packaging claim
is made.

Independent PR #12 review found and cut two residual Agent session readers:
drift no longer asks LastUsed/Reached history callbacks, and Codex model writes
no longer reconstruct old providers from session databases. It also disabled
Agent extension discovery for mixed plugins, while preserving their gateway
entry metadata and original hosts. Three new regression tests failed on the
reviewed pre-fix source with actual history callbacks, a reconstructed legacy
provider, and an executed Agent module respectively. Configuration drift,
manual model switching and restoration remain upstream-derived.

The same review restored upstream `Reseat` and `RenameProvider` at GUI/CLI
Provider-edit seams: a renamed or disabled Provider must not leave an Agent on
an unavailable reference. Both regression tests failed before restoration
(`rename left Agent on "fixture/m"`) and exercise rename, same-model fallback
and disconnect restoration. This is a response to explicit configuration edits,
not the removed automatic account-switching service.

The final ownership pass reproduced an unmanaged Codex catalog/config being
rewritten by SyncCatalog. A Tiny ownership guard now limits automatic sync,
Follow, rename/reseat and port migration to this Tiny profile's actively wired
Agents. The existing TestTinyProviderEditLeavesAgentConfig regression is kept
unchanged; its failure after restoring Reseat exposed the same missing guard.
Connected-Agent list fixtures explicitly establish Tiny ownership before their
original catalog-preservation assertions. No live Agent state was inspected
or modified for these tests.
