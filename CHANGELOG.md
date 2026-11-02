# Changelog

## [0.3.3](https://github.com/caelis-labs/tiny-magpie/compare/tiny-v0.3.2...tiny-v0.3.3) (2026-11-02)


### Bug Fixes

* **agent/codex:** keep the user's model_provider over a takeover or reset ([#4347](https://github.com/caelis-labs/tiny-magpie/issues/4347)) ([e02bc8c](https://github.com/caelis-labs/tiny-magpie/commit/e02bc8c49bd2b377aaacd5bc9b9c8fa605ff1f42))
* **cli:** visible &lt;agent&gt; all clears the models taken out and the only-picked switch ([3b7fe8e](https://github.com/caelis-labs/tiny-magpie/commit/3b7fe8e5f3f77a41b86e9a9fba0ca0cf6730f710))
* **copilot:** list an Experimental vendor's model the picker offers ([#4343](https://github.com/caelis-labs/tiny-magpie/issues/4343)) ([021a757](https://github.com/caelis-labs/tiny-magpie/commit/021a75762aa97c349ffc8cf3e7081e218b316f2d))
* **gateway:** a Codex backend's relay draws the gpt-image family it serves ([#4432](https://github.com/caelis-labs/tiny-magpie/issues/4432)) ([7e44e57](https://github.com/caelis-labs/tiny-magpie/commit/7e44e57b4ef465a992b0f9552db8b74c943a9227))
* **gateway:** a Grok key relay's compaction-blob refusal now strips the sealed reasoning and asks again ([#4412](https://github.com/caelis-labs/tiny-magpie/issues/4412)) ([0639fdd](https://github.com/caelis-labs/tiny-magpie/commit/0639fdd8f9c2e4a52dabf5994503755500d8cf28))
* **gateway:** a model the user said sees is picked though its ID carries a namespace colon ([#4359](https://github.com/caelis-labs/tiny-magpie/issues/4359)) ([4932a8b](https://github.com/caelis-labs/tiny-magpie/commit/4932a8bd1ccc7c9b5ea0b707db7a62b973589497))
* **gateway:** a relayed Responses or Gemini reply's text is read for the calls written into it ([#4743](https://github.com/caelis-labs/tiny-magpie/issues/4743)) ([20f27f5](https://github.com/caelis-labs/tiny-magpie/commit/20f27f514886c00a4bfb6d72245041c5766fe66b))
* **gateway:** Copilot's Gemini takes Codex's disjoint unions as oneOf, its Grok the custom tools as functions ([#4374](https://github.com/caelis-labs/tiny-magpie/issues/4374)) ([a4c74a0](https://github.com/caelis-labs/tiny-magpie/commit/a4c74a059ede0d343836fd4decee071a6bf15dec))
* **gateway:** skip a nested group's effort classifier when the group above picked ([#4358](https://github.com/caelis-labs/tiny-magpie/issues/4358)) ([339ff56](https://github.com/caelis-labs/tiny-magpie/commit/339ff564e632e855f7bd70c324b270af9b852af6))
* **library:** give Kimi Code its own skills folder ([#4379](https://github.com/caelis-labs/tiny-magpie/issues/4379)) ([ecdcbe4](https://github.com/caelis-labs/tiny-magpie/commit/ecdcbe48a898b994b7c675daeacfc1d1bbc13e75))
* **library:** Kiro CLI gets the Library's skills, MCP servers and instructions ([#4769](https://github.com/caelis-labs/tiny-magpie/issues/4769)) ([eb469c7](https://github.com/caelis-labs/tiny-magpie/commit/eb469c70a941b13673c776bf2b0a7a5f8fdc479b))
* **library:** MCP writes to Claude Code's .claude.json take its save lock ([#4747](https://github.com/caelis-labs/tiny-magpie/issues/4747)) ([c57ac41](https://github.com/caelis-labs/tiny-magpie/commit/c57ac416d3244db3c8efe17f54300f7c1cf11587))
* **library:** update skills the skills CLI installed from where its lock says ([#4645](https://github.com/caelis-labs/tiny-magpie/issues/4645)) ([6803d7d](https://github.com/caelis-labs/tiny-magpie/commit/6803d7d51c637838c417671c09dc7f4a8818a5a0))
* **netproxy:** a switch of the global proxy drops the old proxy's idle connections ([#4523](https://github.com/caelis-labs/tiny-magpie/issues/4523)) ([cb35d2f](https://github.com/caelis-labs/tiny-magpie/commit/cb35d2fa9449ae9d994295c146bd7631d6f29c1f))
* **redact:** an error's words for a missing module are read before the quote, not the whole body before it ([#4766](https://github.com/caelis-labs/tiny-magpie/issues/4766)) ([d3bf8f7](https://github.com/caelis-labs/tiny-magpie/commit/d3bf8f73943aa84ad5c2e135b4c0d1d0cf58e41d))
* sync upstream and repair automatic update discovery ([a851fce](https://github.com/caelis-labs/tiny-magpie/commit/a851fceb2b57057e05350b7157dd3544391a59a8))
* sync upstream and repair automatic update discovery ([ccd7fb8](https://github.com/caelis-labs/tiny-magpie/commit/ccd7fb863987edad21a23c621b5201749cab8c8e))

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
