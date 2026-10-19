# tiny-magpie scope

This file is the product contract. Subsystem references elsewhere in this tree retain upstream detail; this page takes precedence for the fork.

## Kept

- Model discovery and detail, Chat Completions, Responses HTTP/SSE, Anthropic Messages/count_tokens, Gemini generate/stream/count, Embeddings, Rerank, direct System One classification calls (TypeSafe/Jev and compatible Providers), and model-related Codex compatibility routes. Existing Magpie IDs, headers, authentication, aliases, error shapes and conversion rules remain upstream's. Upstream's 426 WebSocket-to-HTTP fallback is retained; this is not a new WebSocket implementation.
- All existing API-key/custom/local Providers, built-in subscriptions and community Provider plugins, including their authentication, refresh and quota logic. Claude's real CLI and its private MCP callback are necessary subscription internals.
- Conventional groups (nested groups included), original scheduling, affinity, keys/accounts, quota/cost accounting, budgets, fallback, cancellation and concurrency. No classifier or intent inference is performed. Imported intent groups fail explicitly until their classifier/intent settings are removed.
- Provider auth/model/fetch plugins, gateway middleware, asynchronous hooks, routing strategies and usage callbacks. Bun, moejs and upstream manifest/wire formats stay intact. Pure tool/page/agent extensions are unsupported; mixed packages retain their gateway facets.
- Core built-ins (budget, gateway, groups, middleware, rotation, routing, usage) are fixed on; upstream runtime, schemas and community-plugin controls are retained. There is no Built-in switchboard in the GUI.
- Three-platform desktop/tray and Web GUI. Providers is the default page; Gateway, Routing, Plugins, Usage (quotas, token/cost accounting, output tok/s, TTFT and errors) and Settings remain. Existing visual style, translations, accessibility and keyboard conventions are reused.

## Removed from the product

Agent installation/configuration/Profiles and automatic account switching in other apps; TUI; Library; session browsing/resume and Agent-log ingestion; full conversation recording; cloud backup/sync; public MCP and extra search/vision/tool execution; image/video/audio services; intent-based intelligent routing; request-content redaction and its custom rules; installed font selection; warm-up, rate-limit reset spending, automatic credit spending and check-in; keep-awake; telemetry/cooperation promotions; application updates from the upstream channel.

Images/tool screenshots and client-owned tool calls remain model input. A provider's native search/tool feature is still forwarded. Removing public MCP does not remove Claude's private subscription bridge. Usage storage and pricing needed by routing/key budgets remain; no replacement ledger is introduced.

## Runtime and compatibility

Config/cache paths are `~/.config/tiny-magpie` and `~/.cache/tiny-magpie`, respecting XDG roots; portable markers are `.tiny-portable` and `tiny-data`. Config file formats and `MAGPIE_*` gateway options are inherited. The wire still identifies as Magpie; OS bundle, single-instance, URI and shortcut identities are tiny's. Do not share a data folder with original Magpie or automatically migrate it.

The default model port is 3425. If occupied, stop its owner or set `MAGPIE_ADDR`; tiny never offers to terminate the original gateway. The Web management interface retains its own run-key authentication and network guards. No real credentials are copied for tests.

Some upstream implementation and test helpers remain in coupled packages (Agent identity, credential discovery, sessions data types and accounting); keeping them does not expose their old management APIs. Remove a coupled helper only with a proven replacement-free dependency cut. Do not claim removal of Wails, SQLite, Bun or all Agent-related source.

The automatic acceptance gate is model-protocol equivalence plus absence of removed operations. [Compatibility evidence](compatibility.md) distinguishes fixtures, native builds, platform GUI and real-account use.

## Confirmed gateway boundary, 2026-10-19

System One is a client-requested API through the gateway. Tiny never invokes
Jev or another classifier to choose the conversation model or reasoning level.
Imported classifier/intent/automatic-effort groups remain explicitly refused.
The direct handler and decision discovery header are restored from the
integrated upstream, including its auth, key/account limits and fallback.

Embedding/rerank model metadata is kept in the existing vendor catalog cache,
separate from chat picks. Providers and Gateway expose those IDs and HTTP
examples. The chat model list and upstream plugin wire stay unchanged; no new
public discovery header or automatic cross-model embedding group is invented.
Vector-space compatibility remains the caller's responsibility.

Old redaction/spending/font fields remain parseable for config compatibility,
but cannot reactivate the removed runtime products. Secret scrubbing in logs,
credential storage guards and all gateway network/auth guards remain. Coupled
upstream helpers may remain; deletion means no product entry point or background
action, not a claim that every shared source type was removed.

Whole request bodies are not kept for the removed context-text inspector. Normal
upstream prompt-size parsing and usage/route metadata remain; this is not a
replacement accounting implementation. Native authenticated Codex passthrough
keeps upstream request semantics, while explicit account caps still use routing;
Tiny-managed accounts do not gain automatic credit/reset spending.
