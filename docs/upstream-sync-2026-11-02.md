# Upstream sync, 2026-11-02

Integrated Tiny `61e6d95464b95a9b32e497f979db2f08f14c0f05` (0.3.2)
with the complete 1,361-commit upstream interval
`c37bcdbfc8b17eb12ca3c2fd59c1388eaed4fb7d..c875e66c24052c34622cd410e8f9e3bf259dc188`.
The second merge parent and `UPSTREAM.json` identify the exact upstream source.
No release or installed-application update is part of this change.

## Product decisions

Approved: upstream Responses WebSocket transport, the optional global shortcut
for opening the tray panel (unbound by default), and desktop browser access
(off by default, with upstream key/network guards). Tiny authentication and
existing product cuts remain. Rejected: the key-holder self-service `/me` and
`/v1/magpie/me` pages; their backend, client path exemptions and UI are removed.
The original marketing/update Worker under `site/` is excluded as well.

Model/provider/plugin protocols and subscription bridges continue to use the
upstream implementations. Tiny still has no Agent installer/profile/account
management, session ingestion/recording, public tools/media, smart routing,
redaction product, warm-ups, automatic credit spending or original-app feed.
Agent catalog changes require ownership by this Tiny profile. A new upstream
Codex disconnect history-rewrite call was removed; disconnect restores config
without rewriting conversation history.

## Automatic update repair

The startup check already existed. Two weak links made discovery unreliable:
feed errors waited the normal six-hour interval, and available-version state
was not promptly reflected in the tray/open pages. The existing Tiny download
and confirmation policy remains.

- Start the cancellable clock after notification callbacks are attached.
  First check: five seconds; healthy interval: six hours. Feed failures retry
  after one, five, then fifteen minutes, capped at fifteen minutes.
- Notify the native tray about an available version and open Settings directly
  at Tiny updates. Clear stale ready/restart items when the selected state changes.
- Open pages refresh local update state every thirty seconds. Snapshot reads,
  including initial Settings rendering, do not force another release-feed check.
- Only explicit download/restart actions stage or install anything. Keep Tiny's
  stable/preview selection, repository/asset/digest validation, bundle identity,
  cancellation and request-idle confirmation. Automatic download/restart and
  Windows `.com` activation remain disabled.

The production feed is the public `caelis-labs/tiny-magpie` GitHub releases API;
original Magpie feed overrides are ignored. Public release metadata was read
independently: latest stable `tiny-v0.3.2`, 19 assets, with positive-size SHA-256
metadata and own-repository download URLs. Fixture tests exercise selection and rejection paths. This is
not an installed-app upgrade test.

The reduced UI also fixes a dangling reference left by removing the `/me`
widget and a duplicated brand/title block introduced during header merging.
Existing Tiny browser tests caught both composition problems. Upstream's
new default-folded Connect section and responsive header behavior are retained.

## Maintenance policy

Per the maintainer's instruction, unchanged retained modules rely on upstream
CI. Tiny does not port upstream tests or maintain retired-test inventories.
`scripts/prune-upstream-tests.py --apply` applies a single naming convention:
keep only `tiny_*_test.go` / `tiny-*.test.*` and their sandbox support. The sync
script applies it after merging; the scope gate enforces it. Three-platform
CI runs Tiny boundaries and builds, rather than an upstream suite with a
per-platform retired-feature/stress matrix. The duplicate protocol differential
harness is removed. Existing Tiny release-tool fixtures remain.

New tests are limited to demonstrated effects: failed-check backoff and
notification timing, UI polling without download/feed requests, and Codex
history remaining byte-for-byte untouched. Existing Tiny fork-boundary checks
remain, including its Windows replacement-read regression. No tests for the
unchanged new shortcut or browser-access internals were copied.

## Validation

- Scope and conflict-marker checks and workflow YAML parsing: pass. Tiny’s delta
  is whitespace-clean; existing upstream prompt/SSE fixture whitespace is preserved.
- `go vet -tags nogui ./...`: pass.
- `go test -tags nogui -count=1 -timeout 10m ./...`: pass; 11 packages contain
  Tiny boundary tests. This is not a full upstream-suite result.
- Focused updater/Agent-history race checks: pass.
- Chromium and WebKit Tiny UI: 56/56 pass, including automatic discovery while
  Providers/Settings stays open, without a manual Check action.
- Fault injection: removing feed backoff makes the virtual-clock regression
  fail; removing UI polling makes the four discovery cases fail; restoring
  history rewriting makes the byte-preservation regression fail as expected.
- macOS arm64 application bundle, Windows amd64 desktop cross-build and Linux
  amd64 CLI cross-build: pass. Only inherited macOS SDK/C enum/linker warnings.
- Tiny publication, artifact inventory, notarization retry and release-resume
  fixture scripts: pass. These fixtures do not publish anything.

Upstream CI was independently read for the exact pinned SHA:
[run 41375812712](https://github.com/yetone/magpie/actions/runs/41375812712).
Windows tests and Linux/Android CLI builds passed. Linux/macOS tests failed:
`TestEveryRouteHasACommandOrIsNamed` lists the check-in-account command gap;
`TestCodexPoolSaysTheClientsVersion` reports client-version header mismatch;
macOS also has a window-hidden-report timeout. These are upstream results,
not Tiny failures that were fixed or claimed green. Original full-suite
runs attempted before the maintenance-policy change are diagnostic only.

Not validated: real vendor/accounts, registered OS shortcut interaction,
native Windows/Linux GUI, signing/notarization of this local build, container
smoke, clean-machine install or an actual application update/restart. Tests
use isolated homes, failing stand-in credential/Agent tools and fake vendors;
no live Agent or original Magpie state is used. No push, publication or install.
