# Changelog

## [0.3.4](https://github.com/caelis-labs/tiny-magpie/compare/tiny-v0.3.3...tiny-v0.3.4) (2026-11-21)

### Codex credits

* Codex requests can continue on available credits after the included allowance reaches 100%. OpenAI determines eligibility and debits the balance. Tiny reads balance updates from Codex responses and rate-limit events.
* **Use credits** on the account card and `tiny-magpie quota credits <account> on|first|last|off` control the account's preference. Account and window caps and gateway key restrictions still apply.

### Gateway and desktop

* Merge 2,620 upstream commits through `1b360f08c208d75da393f083f667a57bac8f9772`, including Responses WebSocket, streaming, retry and usage-accounting improvements.
* Serve explicit Decisions and System One requests. Existing Tiny catalog caches retain their embedding and rerank model kinds.
* Keep the Tiny sidebar wordmark on one line. The existing width check hides the complete wordmark when space runs out.

### Scope and validation

* Tiny keeps model and provider protocols, account pools, routing, and its own update channel. Hosted chat and join pages, memory and Library, public MCP and Code Mode, session ingestion, media endpoints, remote configuration following and chained updates stay outside the product.
* Gateway and UI checks use isolated fixtures. Real Codex account eligibility and debits, installed upgrades, and native Windows/Linux interaction remain unverified. [Sync evidence](https://github.com/caelis-labs/tiny-magpie/blob/main/docs/upstream-sync-2026-11-21.md)

## [0.3.3](https://github.com/caelis-labs/tiny-magpie/compare/tiny-v0.3.2...tiny-v0.3.3) (2026-11-02)

### Fixes

* Automatic update discovery retries failed feed checks after 1/5/15 minutes and refreshes the tray and open pages. Checks use only Tiny releases; downloading and restarting remain explicit actions.
* Codex disconnect restores configuration without rewriting conversation history.
* Merge 1,361 upstream commits through `c875e66c24052c34622cd410e8f9e3bf259dc188`, preserving Tiny's model/provider/plugin protocols, authentication and product cuts.

### Added

* Upstream Responses WebSocket transport and HTTP fallback.
* Optional global shortcut to open the tray panel; unbound by default.
* Desktop browser management access with upstream key/network guards; off by default.

### Maintenance and upgrade notes

* Tiny keeps focused fork-boundary tests and platform builds; unchanged retained implementations rely on pinned upstream CI. Retired-product tests are not ported. Key-holder `/me` pages remain excluded.
* If an older version does not show this update automatically, use Settings → General → Check once. The discovery repair takes effect after upgrading.
* macOS uses the signed/notarized Apple Silicon DMG. Windows/Linux are unsigned. Fixtures and CI do not establish real-account, installed-app upgrade or native Windows/Linux interaction acceptance; see [sync evidence](https://github.com/caelis-labs/tiny-magpie/blob/main/docs/upstream-sync-2026-11-02.md).

## [0.3.2](https://github.com/caelis-labs/tiny-magpie/compare/tiny-v0.3.1...tiny-v0.3.2) (2026-10-27)


### Features

* sync upstream through c37bcdbf with Tiny scope preserved ([a8747b8](https://github.com/caelis-labs/tiny-magpie/commit/a8747b844bfacf98a67871218acb15e9cc797d39))


### Bug Fixes

* create cache notifications only for subscribers ([e963c9a](https://github.com/caelis-labs/tiny-magpie/commit/e963c9a055ade429eaa2cb23061278dade1a9f19))
* **gui:** capture only geometrically settled scroll anchors ([671bea6](https://github.com/caelis-labs/tiny-magpie/commit/671bea6c9d9d98a1f7e4899001dabe624318bdad))
* **gui:** exclude pressed transforms from scroll anchors ([5eca61d](https://github.com/caelis-labs/tiny-magpie/commit/5eca61da62bae4e7bc9cf715ea7d17d6f65c60c4))
* **gui:** keep existing anchors clear of pressed boxes ([209faf5](https://github.com/caelis-labs/tiny-magpie/commit/209faf54f2ad4b95eca9b7c121124f338ac588e7))
* **gui:** preserve Tiny editors and cover synced features ([01bd36b](https://github.com/caelis-labs/tiny-magpie/commit/01bd36b703f4be61db6c317cfedd7f3a005447b1))

## [0.3.1](https://github.com/caelis-labs/tiny-magpie/compare/tiny-v0.3.0...tiny-v0.3.1) (2026-10-22)


### Bug Fixes

* restore tray Agent selection and automate Tiny releases ([#14](https://github.com/caelis-labs/tiny-magpie/issues/14)) ([0a92e61](https://github.com/caelis-labs/tiny-magpie/commit/0a92e61d119b17cf14934b54827ca89bdd8fbacb))
* **update:** keep slow Windows downloads alive while bytes arrive ([eb80aeb](https://github.com/caelis-labs/tiny-magpie/commit/eb80aeb11e9124eb1f603db72deca0afbdc68757))
