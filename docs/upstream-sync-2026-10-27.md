# Upstream sync — 2026-10-27

## Source and boundary

- Tiny starting point: `3e13a52e8e67742a0f053b7581f1af7b13ab20f4` (0.3.1).
- Previous upstream: `ba319a7587923fc7a4ea3859658244d591261a59`.
- Integrated interval target: `c37bcdbfc8b17eb12ca3c2fd59c1388eaed4fb7d`, 617 commits, 1,404 changed paths before Tiny scope resolution.
- Full interval merge; [tiny-scope.md](tiny-scope.md) remains the product contract. No expansion into removed products, upstream update channels, or real-account testing.
- Branch: `codex/upstream-sync-20261027`. This document records integration and acceptance, not a release or production validation.

## New feature decisions

| Feature / representative upstream change | Decision and reasoning |
| --- | --- |
| Speed-first scheduling (`d9a8148d`) | Include. Conventional routing by observed latency, without model-based classification. Existing affinity, budgets and cancellation remain. |
| Several ordered models in a rule (`f99d27ab`), drag ordering (`b5a78fd8`) | Include. Extends explicit routing rules. Classifier/intent/automatic-effort groups remain refused. |
| Time-zone-aware rule hours (`c888ef27`) | Include. Rules retain the zone in which the user set them. |
| Single-provider Patience (`d56c14e9`), optional no-rest after rate limits (`c6768f06`) | Include. Explicit provider policy; retries remain bounded by request cancellation and configured patience. No automatic spending/warm-up is enabled. |
| Per-key plans (`6e6aba51`), key consolidation (`5b4dec58`) | Include within provider account/model management. Preserve upstream credentials and storage protocol. No live-account migration is claimed. |
| Model specification overrides (`4a3e07f1`), usage aliases (`ea9454d2`) | Include. Explicit context/output/vision/reasoning metadata and accounting names are within the existing model/usage boundary. |
| Relay-supplied Codex model entries (`6eb510f7`) | Include. Preserve the relay's own context window, effort levels and instructions. Ordinary model-list deduplication remains; a distinct Codex `client_version` metadata read happens once per refresh. |
| Catalog freshness and pricing, OpenRouter metadata (`3133f634`), Labs models (`f51ab771`) | Include. Retained provider/model discovery, not intelligent routing. |
| Usage line charts (`1c617c15`), viewer time zone (`e00fbdbb`), probe records (`ffe49377`) | Include for gateway usage. Keep Agent-log sources and identities empty; session-list performance work does not authorize session ingestion. |
| Responses dynamic `additional_tools` projection and fallback | Include. Preserve client-supplied tool definitions and cached prefixes; no Tiny tool executor or Library is added. |
| Routing live-view/rendering performance, cache-unknown display | Include. Preserve Tiny token/cost/speed/TTFT summaries and compact desktop/Web layouts. |
| Repeated client-tool-loop guard | Include as an extension of the existing reply-loop guard. Stops a repeatedly identical tool exchange without executing a tool; honors the existing loop-guard setting. Client input/tool forwarding remains upstream semantics. |
| Local goroutine dump command | Include beside existing heap/CPU diagnostics. Uses the gateway's loopback guard and does not stop the process. |
| WSL NAT gateway reachability (`a1c1b705`) | Include with upstream WSL-network and key checks. Requires Windows/WSL acceptance before claiming real-platform validation. |
| PWA/home-screen metadata (`9022169f`), one-port Web serving (`1b32bdcf`) | Include. Within the retained Web GUI; PWA identity and generated home-screen icons are Tiny; native Dock animation uses Tiny geometry. Management authentication remains required. |
| New Agent detection/manual wiring: AionUi, KiroCrew, NewMax, Penguin Harness, ZooCode, WSL agy | Include only model/provider/effort wiring, explicit connection/restoration and visibility. Existing Tiny ownership gate still controls automatic catalog sync. |
| Download retries, release-history browsing, desktop handover | Include using Tiny's GitHub feed, channels and OS identity. No upstream stable-release selection, remote notes feed, or original Magpie launcher is enabled. |
| Plugin/provider compatibility: authentication, model schemas, keyless/single-account behavior, Bun host and middleware | Include. Preserve upstream provider and plugin contracts. Pure Agent/tool/page facets remain unsupported. |
| Windows console `.com` companion (`2afacd7d`) | Defer activation. Shared code uses Tiny names, but existing packaging does not enable `comlauncher`. Dedicated Windows terminal/console acceptance is needed before enabling it. |
| Codex upstream WebSocket transport (`a2e98cd7`) | Exclude. Current contract keeps HTTP/SSE and the existing 426 fallback; new transport implementation and its positive suite are removed. |
| Privacy models, masking monitor/rules/confidence controls | Exclude. No masking built-in, navigation, scripts, or APIs. Existing request-mask adapters remain inert; credential scrubbing remains. |
| Session conversion/migration (`15f0d4cd`), Codex thread relabeling and history ingestion | Exclude. Remove new automatic thread moves and session-database reads used to rewrite other providers. Coupled parser/accounting types remain. |
| Local/full request archive | Exclude. The existing Tiny `archiving` path still returns nil even with old consent/settings. |
| Agent auto-restart (`d08b6118`), shell-profile wrappers, auto-review/memories/subagent configuration | Exclude. No new restart UI or automatic local/WSL shell-profile writes. Native client tool calls remain forwarded. |
| Public MCP, Library/tool CLI (`402577c1`, `ea1f9492`), extra vision/search execution | Exclude. Keep client images/tool calls and private Claude subscription internals. |
| Check-in, credit/reset spending, warm-up, partner promotion fetch | Exclude. No new background job or product entry point. |
| Upstream stable release channel (`4a33cf0a`, `8f446995`) | Exclude. Tiny stable/preview continue to use only its own GitHub releases. |

## Integration corrections

The merge initially had 118 unmerged paths, including 82 modify/delete conflicts. Deletions were retained; text conflicts were resolved by behavior, not by accepting entire files from one side. Automatically merged code was also audited for product entry points.

- Preserve the plugin-store write lock and add upstream file-cache invalidation.
- Adapt changed request-mask signatures while keeping Tiny's no-op behavior.
- Keep the CLI's retained command boundary and adopt lazy stdin acquisition for Windows console correctness.
- Remove duplicate update-channel declaration and route new release-history reads through Tiny's feed.
- Admit new model-wiring tests individually. Keep shared fixture helpers without restoring removed installation, masking, session or tool-product suites; mixed tests retain their in-scope assertions.
- Render configuration diffs safely when an optional previous value is absent.
- Remove the stale `addressMore` reference that prevented Provider editors from opening. Keep Tiny's folded Advanced layout, extend narrow routing layout into it and wrap time-zone labels. The newly admitted feature tests caught these integration regressions.
- Preserve Tiny identity for PWA metadata/icons, native Dock geometry and the optional Windows companion launcher.
- Restore preview-channel validation, failed-download cleanup/declared-size verification, the System One single-port alias and gateway-only Usage heatmap/read-memory handling. Resumed bytes exist only within one bounded download operation and are removed on final failure or cancellation.
- Keep multi-protocol list deduplication assertions at one ordinary fetch and separately require one Codex metadata fetch; the selected source added the latter without updating three older request-count fixtures.
- Adapt Bun proxy fixtures for its equivalent explicit default HTTP port; preserve exact credentials and target path.

## Upstream failure found independently

The selected source has passing Windows tests but failing Ubuntu/macOS test checks in [run 40464188023](https://github.com/yetone/magpie/actions/runs/40464188023). Ubuntu's provider suite aborts with `close of synctest channel from outside bubble` in `filememo.Forget`, through `TestUnreadLoginsWriteErrorsKeepCredentials`. This is an upstream failure, not evidence that Tiny passed.

The integration makes notification channels lazy: invalidation closes and clears a subscribed channel; the next subscriber creates its generation. Invalidating without subscribers no longer leaves a channel owned by a completed synthetic clock. A focused regression exercises independent clocks and notification delivery. The source SHA remains unchanged; this correction is a local merge adaptation.

## Validation

Local validation completed. Executed tests use temporary HOME/config/cache roots or repository fixtures; no real Agent/Magpie configuration or provider account was used. Cross-compilation does not execute the target binaries.

- Scope gate: passed after removing reintroduced endpoints/assets/tests.
- Complete retained test compilation: passed.
- Binary differential against the exact source SHA: passed discovery, 67 bidirectional HTTP/SSE comparisons, removed endpoints and no-Agent-config-write checks.
- Browser gate: 84 Tiny core cases plus 182 retained feature cases (266 total) now run by default. Local core/feature checks and focused reruns cover editor opening, model visibility, key plans, rate-limit policy, ordered/dragged rules, speed-first routing and time zones. See the PR's current-head protocol/UI check for the complete combined gate.
- `go vet -tags nogui ./...`, native macOS arm64 production build (Makefile-equivalent macOS 12 deployment target) and nogui build: passed. Native enum/linker warnings remain; this is not minimum-macOS runtime acceptance.
- Final binary differential repeated after integration corrections: all 67 comparisons and boundary checks passed.
- Complete retained Go suite: `go test -tags nogui -timeout 20m ./...` passed; gateway 508.392 seconds, provider 157.422 seconds. Earlier runs exposed the documented corrections; a default 10-minute run exhausted the gateway package budget while still advancing through tests.
- All Windows amd64 nogui tests cross-compiled with `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go test -exec=/usr/bin/true -tags nogui -run '^$' ./...`; target execution intentionally skipped. Windows amd64 production executable also cross-built.
- The first expanded remote UI run at `01bd36b7` timed out clicking an animated model-list entry and then stopped advancing; it was canceled after 20 minutes. Screenshot cleanup could skip browser closure on error. The follow-up registers cleanup immediately, closes browsers in `finally`, uses reduced motion for model-state assertions and bounds the UI job at 20 minutes. The original run remains available as failure evidence.
- The next expanded CI run completed all 182 feature cases and exposed four WebKit rule-editor fixture failures: automatic scrolling fought Routing's scroll-position hold, leaving targets outside the viewport or the fold closed. The fixture now wheels controls fully into view before normal click/hover actions and uses reduced motion for order/state assertions; model, visibility and saved-rule assertions remain.
- The following CI run passed all 20 rule-model combinations and 265 of 266 total browser cases. Its remaining German WebKit drag case measured a two-pixel scroll difference. The fixture now waits for fonts and a stable scroll position beyond the reader-scroll window before taking its baseline, with exact position checks after drag, keyboard reordering and menu reordering; no pixel tolerance was added.
- A full local Node 22 feature run then passed the drag checks but exposed two time-zone fixtures whose Advanced fold stayed closed with motion enabled. The parallel CI also completed 265 of 266 browser cases, with its one remaining failure in the same disclosure path. Their state/layout/save checks now use reduced motion and wait for fonts, consistently with the other routing state fixtures. The complete 182-case local feature run then passed under Node 22 with two workers and screenshots. The original failures remain in the PR evidence.
- Scope and staged diff checks passed. Upstream unified-diff patch context whitespace is preserved with an exact path attribute exemption.
- First integration head `fda9761f` passed all platform and protocol/UI checks in [run 40471220967](https://github.com/caelis-labs/tiny-magpie/actions/runs/40471220967) and both container builds in [run 40471221097](https://github.com/caelis-labs/tiny-magpie/actions/runs/40471221097). Subsequent UI fixes and the expanded gate require the PR's current-head checks; previous-head success is not reused as that evidence. Native Windows/WSL GUI, real account and distribution/install acceptance remain unclaimed.

No merge to the default branch, release, live provider-account acceptance, Windows/WSL GUI acceptance, or production deployment is implied.
