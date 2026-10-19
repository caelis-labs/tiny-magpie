# Releases

Only `caelis-labs/tiny-magpie` is a release destination. Upstream Magpie release channels remain disabled. The source and modifications remain MIT licensed; retain the original copyright notice and Git history.

Tiny tags use `tiny-vX.Y.Z` so they do not collide with inherited upstream tags. Do not push upstream tags to origin. The tag must identify a clean commit on main, with a green Test tiny-magpie run before tagging. The Release tiny-magpie workflow reruns the full platform, protocol/UI and container gates at the exact tagged source before packaging.

1. Add reviewed notes at `docs/releases/X.Y.Z.md`, verify UPSTREAM.json, and merge through the protected main branch.
2. Wait for the six required checks: three OS Go suites/native builds, protocols-and-ui, and two architecture container checks. Do not tag a cancelled or failing head.
3. Create and push an annotated `tiny-vX.Y.Z` tag at that exact main commit. The tag triggers release. Once candidate artifacts exist, rerun only failed jobs (`gh run rerun RUN_ID --failed`) to reuse their original bytes. A fresh workflow_dispatch rebuilds and signs again; it must not replace existing draft assets or replay an unknown notarization submission.
4. The workflow builds six desktop and six CLI archives. macOS uses shared organization secrets `APPLE_DEVELOPER_ID_P12_BASE64`, `APPLE_DEVELOPER_ID_P12_PASSWORD`, `APPLE_NOTARY_APPLE_ID`, `APPLE_NOTARY_PASSWORD` and organization variable `APPLE_TEAM_ID`. No repository copy of private material is needed. Signing uses a temporary runner keychain, hardened runtime and secure timestamps; no additional entitlements are introduced. Windows/Linux are unsigned.
5. macOS submits both signed binaries to Apple, persists the signed submission archive and receipt, waits for Accepted, staples and validates the app, checks Gatekeeper and verifies the CLI's online ticket. Signature, team, notarization or ticket failures stop publication. A timeout is unknown: inspect the uploaded notary-receipt artifact, retain its archive and query its submission ID before resubmitting. An accepted original archive can be recovered without replacing or signing its binaries again.
6. Each of the six platform/architecture runners checks the native CLI's linked version in an empty profile. Packaging checks embedded Go source SHA and clean source, records binary/archive hashes, and includes LICENSE/UPSTREAM.json in each archive. `-trimpath` omits linker flags from Go build metadata, so version is checked by execution. Publishing checks the complete 23-asset inventory, creates a draft, downloads and verifies every uploaded asset, then publishes it as a Preview. A retry reads existing draft assets and rejects mismatched bytes; it never overwrites them. A published release is not modified on retry.
7. Independently download the public release, check SHA256SUMS and embedded BUILD evidence, and verify the macOS app signature, stapled ticket, Gatekeeper and CLI notarization. Cross-platform CI is distinct from real native interaction and real accounts.

Main protection requires a PR, current successful checks, resolved review conversations, and applies to administrators. Force pushes and branch deletion are disabled. Approvals follow the organization's single-maintainer setup (no mandatory second reviewer). Merge commits remain available for complete upstream intervals. Release-tag rules prevent moving or deleting `tiny-v*` tags.

The publisher resolves existing drafts with `gh release view` and reads them through their release ID. GitHub's REST tag endpoint returns published releases only; a draft's 404 there must not cause another draft to be created. `scripts/test-publish-release.py` exercises that contract, unknown creation/upload outcomes, published retries and mismatched assets.

The first releases remain experimental until the remaining [stable acceptance](compatibility.md#before-a-stable-release) is complete. There is no Tiny automatic application updater or registry/container publication in this workflow.

PR validation runs once per reviewed branch; release tags run the [retained matrix](test-scope.md#ci-execution) again on the actual source. Main pushes do not repeat it. Routine Windows validation compiles every test and executes the native regression gate; the complete Windows corpus runs weekly.

## Current artifact formats

The 0.1.1 Preview covers the same six desktop OS/architecture pairs as
upstream, but its artifact formats are not identical:

| Product | Tiny 0.1.1 | Upstream v0.1.1290 |
| --- | --- | --- |
| macOS amd64/arm64 desktop | ZIP containing signed, notarized `.app` | DMG and ZIP |
| Windows amd64/arm64 desktop | ZIP containing unsigned `.exe` | standalone `.exe` |
| Linux amd64/arm64 desktop | tar.gz containing unsigned executable | standalone executable |
| CLI, the same six pairs | tar.gz or ZIP with license/provenance | standalone executable |
| Android arm64 CLI | outside Tiny's current desktop matrix | standalone executable |

The first Tiny workflow used uniform archive packaging and omitted DMG
creation; this is not a signing/notarization limitation. A macOS ZIP contains
an actual application bundle: extract it and move `tiny-magpie.app` to
Applications. CI/test-only changes do not repack or replace an existing
release. See [the upstream release](https://github.com/yetone/magpie/releases/tag/v0.1.1290)
for the compared assets. Stable installer acceptance remains a separate gate.
