# Tiny test scope

Tiny follows complete upstream commits and keeps excluded products deleted.
Unchanged retained implementations rely on the pinned upstream's CI. Tiny does
not port upstream test suites or maintain retired-product test inventories.
This policy takes precedence over inherited instructions to run all upstream
regressions or duplicate their protocol coverage.

`python3 scripts/prune-upstream-tests.py --apply` removes inherited Go/browser
tests after each merge, including newly added files. It keeps the `tiny_` Go
and `tiny-` browser naming convention. Supporting sandbox fixtures use the same
prefix. `scripts/check-scope.py` enforces the convention; no product-by-product
allowlist is maintained.

Keep existing Tiny checks that protect a real fork boundary: own update source
and identity, independent data paths, excluded entry points remaining inert,
strict ownership of Agent configuration, gateway-only usage, plugin facet cuts,
and the reduced GUI. Add a small regression only when a Tiny change or cut has
a demonstrated effect. Do not copy tests for unchanged translators, provider
implementations, shortcuts or browser-access internals just because upstream
added them. Shared fixtures provide isolation, not positive tests for removed
products.

For this sync, demonstrated regressions are the updater's six-hour wait after
feed errors, available-update state not reaching open windows/tray, and Codex
disconnect rewriting conversation history. Their focused tests and fault
injections are recorded in the sync report.

CI runs scope checks, Go vet, Tiny boundary tests and native builds on three
platforms; Chromium/WebKit check the reduced UI. Release identity, packaging
and publication scripts retain their own fixture checks because Tiny owns
those implementations. All tests use isolated homes, failing stand-in Agent
and credential tools, and loopback fake vendors. Build, fixture, native GUI,
installation and real-account evidence are reported separately.

Record the integrated upstream SHA, its available CI evidence, Tiny changes,
new product decisions and the checks actually run. Missing upstream evidence
must be stated; it is not a reason to invent a claim of full upstream parity.
