# Upstream sync, 2026-11-21

Tiny integrates the complete 2,620-commit interval from
`c875e66c24052c34622cd410e8f9e3bf259dc188` through
`1b360f08c208d75da393f083f667a57bac8f9772`. The second merge parent and
`UPSTREAM.json` identify the pinned source. Its
[upstream Test workflow](https://github.com/yetone/magpie/actions/runs/44325491375)
passed on Linux, macOS and Windows, including the Linux and Android CLI jobs.

The next-reviewed release at `9924a48bbf7a53521ddeafb439eff455e1aa23a0`,
[v0.1.4625](https://github.com/yetone/magpie/releases/tag/v0.1.4625), had
[two failed macOS jobs](https://github.com/yetone/magpie/actions/runs/44863597903).
[v0.1.4626](https://github.com/yetone/magpie/releases/tag/v0.1.4626) was
published during the sync. Its one additional commit fixes the gateway's
recovery from an empty reply at the vendor's default output limit. Its
[Test workflow](https://github.com/yetone/magpie/actions/runs/44866238868)
was still running when checked. Both releases remain ahead of the pin.

## Codex credits

Tiny had hard-coded the credits policy off. An account with credits visible
in Usage could be held at 100% before a request reached the vendor. Codex
requests can now proceed on available credits by default. The vendor
determines eligibility and debits the balance. **Use credits** in the
account card and `tiny-magpie quota credits <account> on|first|last|off`
use the existing per-account preferences. Account and window caps and
gateway key restrictions continue to apply.

The pin also includes upstream's
[credits header and event handling](https://github.com/yetone/magpie/commit/8bf9cb99d466543cb27f4730e1e815fcf0677a7a).
A vendor's newly reported balance can clear a rest caused by ordinary
quota exhaustion. Reset-card spending and automated Agent account
switching remain outside Tiny's scope.

## Feature selection

- Retain model transport and accounting fixes, including Responses
  WebSocket handling, streaming and retry behavior, account and key caps,
  and vendor-reported allowance. Keep direct Decisions and System One
  calls, plus embedding and rerank discovery. Existing Tiny catalog
  caches keep their retrieval kinds on read.
- Keep memory, Library, public MCP and Code Mode, session ingestion,
  hosted chat and join pages, media endpoints, and arbitrary vendor
  API passthrough outside the product. Remote Magpie still serves as a
  model provider. Remote Library, Agent configuration following, and
  chained application updates have no Tiny entry point.
- Defer account enumeration and pinning extensions, any new network
  admission flow, and deeper Agent management until there is a concrete
  client requirement. The gateway's existing account pool and access
  controls cover ordinary routing.

## Validation

- The old Codex policy reproduced a 429 before reaching the loopback
  vendor. The restored policy passed for an account at 100% with credits.
  An explicit opt-out and an explicit usage cap each blocked the request.
- Tiny assembly and lifecycle checks, Go vet and the full Tiny Go boundary
  suite passed. The 56 Chromium and WebKit cases cover Tiny’s existing
  Routing page and group editor, alongside the other retained pages.
- The macOS arm64 app build, Windows amd64 desktop cross-build and Linux
  amd64 CLI cross-build succeeded. Tiny's release-tool fixtures passed.
  Signing, notarization and installed native UI checks are outside these
  local build results. Upstream-only whitespace in imported files was
  preserved.
- The test homes, vendor services and browser responses were isolated.
  Real Codex account eligibility and debits remain unverified. This
  change prepares a local branch for review without publishing a build.
