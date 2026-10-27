# Tiny test scope

The product boundary is [tiny-scope.md](tiny-scope.md). Test deletion follows
that boundary; it does not replace retained gateway implementations or silence
failed assertions. The first Preview still ran substantial removed-product
coverage. Issue [#6](https://github.com/caelis-labs/tiny-magpie/issues/6) records
the original measurements and the reduction.

## Removed positive suites

- Agent installation, CLI updates, profiles, and orchestration/optimization UI.
  Agent model wiring, effort, catalog sync, stashes and disconnect/restore regressions
  are restored verbatim from the pinned upstream (including shared fixture suites).
  The sandbox `TestMain` and `TestFolderVarsAreListed` guard remain.
- Library instructions, skills and public MCP setup; cloud backup/sync and
  profile bundles; keep-awake, image/video MCP generation, telemetry and
  automatic context tuning. Their positive Go suites are deleted.
- GUI tests accepting Library, session pages/export/resume,
  cloud backup/sync, telemetry, intelligent routing and context tuning. Corresponding
  browser tests leave the tree too. Shared GUI sandbox fixtures are kept
  unchanged in `internal/gui/sandbox_test.go`.

Session management/delete/restore/relocation, search and Markdown-export positive
tests are removed as well; mixed files keep their parsing/accounting tests and
shared fixtures. Upstream application self-update download, feed, installer,
relaunch and release-note tests leave the suite. Tiny now adds its own small
fixture suite for its GitHub source/channel, trusted DMG staging, automatic
check lifecycle and user-confirmed UI; it does not restore upstream-feed tests. Version comparison and
executable/CLI identity helpers used by plugins, Bun and Tiny startup remain.

## Retained guarantees

Gateway, all Provider families, subscription/plugin migrations and host,
middleware, routing/fallback, budgets, log credential scrubbing, cancellation, streaming,
usage/account attribution, SQLite, concurrent settings/account writes,
profile isolation and GUI/network guards keep their regressions. Vendor-owned
search and tool calls, screenshots and the private Claude subscription bridge
remain model/protocol behavior. Session parsing/accounting helpers remain
coupled to usage and keep their shared corpus; removing a session page does
not justify removing accounting coverage.

Tiny fixture tests exercise Agent model switching and configuration restoration.
Tiny negative tests reject removed endpoints, imported intent/recording
settings and standalone tool/page/Agent extensions, and verify Provider edits
leave Agent configuration alone. `scripts/check-scope.py` also rejects the
return of removed product suites during upstream integration. A new upstream
test in a retired suite requires scope review, rather than automatic adoption.

## CI execution

| Event | Linux/macOS | Windows | Protocol/UI and containers |
| --- | --- | --- | --- |
| PR | vet, complete retained Go suite, native desktop build | vet, compile every test, native desktop build, focused paths/state/storage/plugin/removed-surface checks; five-count account/proxy stress | 67 upstream differential comparisons, 84 Chromium/WebKit cases; amd64/arm64 container smoke |
| Main push | no duplicate matrix | no duplicate matrix | no duplicate matrix |
| Release tag | complete retained suite on the exact release source | same focused native gate | protocol/UI/container checks before five native packages |
| Weekly, Monday 03:00 UTC | complete retained suite | additionally complete retained suite and twenty-count stress | protocol/UI fixture gates |

Manual runs use the normal PR-sized gate. Routine jobs have a 20-minute job
budget; weekly jobs allow 60 minutes, with 45 minutes per Windows package.
The full Windows corpus is outside PR and release's critical path. This
follows upstream's compile-plus-native-regressions strategy while preserving
Tiny's added concurrent-write, isolation and removed-product checks.

Separate raw JSON/log artifacts retain Windows compile, fixture and stress
results; a later phase cannot overwrite an earlier phase's evidence. Tests
use `-count=1` (stress has its explicit count), isolated homes and fixture
vendors. No real accounts, Magpie profile or Agent state is used. Failures
remain failures; timeout increases or retries are not a replacement for fixes.

Main protection keeps all six existing checks. Strict PR checks and reviewed
merge-tree equality establish the merged code; release validation runs again
on the actual tagged commit. No new CI proof-reuse service is introduced.

Native Windows/Linux GUI, clean-machine installation and real-account
acceptance remain distinct [stable-release gates](compatibility.md#before-a-stable-release).

## Reduction measurement

Compared with released 0.1.1, Go test source changes from 2,121 files,
339,744 lines and 6,450 `Test*` functions to 1,654 files, 254,915 lines and
4,934 functions (including `TestMain`). 558 obsolete test files are deleted:
468 Go files and 90 browser cases. Shared fixture code is preserved. These
are source-size measurements, not a coverage percentage or a live-vendor
compatibility claim.

## Gateway-core Dev iteration

Direct System One API tests and differential fixtures are retained; intelligent
route classification remains a negative gate. Embedding/Rerank fixtures compare
vendor bodies and replies as well. The differential corpus now has 67 calls.
Removed masking/font/builtin-switch/reset-spending positive tests leave the
suite; shared auth, request fixtures, quota accounting and log scrubbing remain.
New Tiny checks cover explicit classification calls, split retrieval inventory,
fixed core assembly, visible token/cost/speed/TTFT summaries, unavailable retired settings APIs, old settings remaining
inert, and real rendered icon bounds. Browser acceptance uses both engines,
five languages, narrow/wide layouts and isolated native Dev inspection.

The gateway-page matrix uses 1100px and the real 560px native minimum, plus
440px in the Web phone layout. It keeps the original assertions and action
timeouts. Failed cases save a screenshot and navigation bounds in the existing
browser-evidence artifact before closing their browser.

The PR #12 independent review removes `TestDriftBypassed` and
`TestCodexDriftReached`: both require the retired Agent-log readers. New Tiny
regressions assert that configuration drift never invokes history callbacks,
Codex model picks never import session-only providers, and mixed plugins never
execute their Agent facet. The other restored upstream Agent tests remain.

Automatic Agent updates now require Tiny ownership. The catalog/list fixtures
explicitly Apply a model first; the multi-account disconnect fixture now
expects an unmanaged first start to leave Codex alone, while retaining all
connect/disconnect/reconnect assertions. The unmanaged Agent-config API guard
is unchanged, and a new catalog regression verifies both foreign-config
preservation and synchronization after an explicit model pick.

## 2026-10-27 integration

The [sync decision record](upstream-sync-2026-10-27.md) records the full interval.
New manual model-wiring/visibility tests are admitted individually. Local HTTP
fixtures for download retry/Range/checksum behavior and relaunch fallback are
retained for Tiny's updater; upstream feed/channel tests remain excluded.

New positive Agent-log ingestion cases in `request_cells`, `request_pairs`,
`request_disk`, `request_disk_map_unix`, `cache_bounds` and `heatmap` are removed
because they require the history sources Tiny intentionally leaves empty.
Mixed files retain their pure chunk/parser/accounting tests. The two new
vision-backed wait cases, masking facts case and Antigravity drawing suite
likewise require removed execution surfaces; generic wait/cancellation and
protocol regressions remain. Responses `additional_tools` translation and its
provider flag are retained client-protocol behavior, not tool execution.

Tiny-specific regression coverage also checks that local/WSL manual agy
model picks do not write shell startup files. The shared cache regression
checks invalidation across independent synthetic clocks. PWA names, native
flap geometry and updater paths use Tiny identity; Bun proxy authentication
accepts the equivalent explicit default port without relaxing credential or
request-path checks.
