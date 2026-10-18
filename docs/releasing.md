# Releases

Only `caelis-labs/tiny-magpie` is a release destination. Upstream Magpie release channels remain disabled. The source and modifications remain MIT licensed; retain the original copyright notice and Git history.

Tiny tags use `tiny-vX.Y.Z` so they do not collide with inherited upstream tags. Do not push upstream tags to origin. The tag must identify a clean commit on main, with a green Test tiny-magpie run before tagging. The Release tiny-magpie workflow reruns the full platform, protocol/UI and container gates at the exact tagged source before packaging.

1. Add reviewed notes at `docs/releases/X.Y.Z.md`, verify UPSTREAM.json, and merge through the protected main branch.
2. Wait for the six required checks: three OS Go suites/native builds, protocols-and-ui, and two architecture container checks. Do not tag a cancelled or failing head.
3. Create and push an annotated `tiny-vX.Y.Z` tag at that exact main commit. The tag triggers release; workflow_dispatch can reconcile the same tag after a failure.
4. The workflow builds six desktop and six CLI archives. macOS uses shared organization secrets `APPLE_DEVELOPER_ID_P12_BASE64`, `APPLE_DEVELOPER_ID_P12_PASSWORD`, `APPLE_NOTARY_APPLE_ID`, `APPLE_NOTARY_PASSWORD` and organization variable `APPLE_TEAM_ID`. No repository copy of private material is needed. Signing uses a temporary runner keychain, hardened runtime and secure timestamps; no additional entitlements are introduced. Windows/Linux are unsigned.
5. macOS submits both signed binaries to Apple, persists the submission receipt, waits for Accepted, staples and validates the app, checks Gatekeeper and verifies the CLI's online ticket. Signature, team, notarization or ticket failures stop publication. A timeout is unknown: inspect the uploaded notary-receipt artifact and query its submission ID before resubmitting.
6. Packaging checks embedded Go source SHA, clean source and version, records binary/archive hashes, and includes LICENSE/UPSTREAM.json in each archive. Publishing checks the complete 23-asset inventory, creates a draft, downloads and verifies every uploaded asset, then publishes it as a Preview. A retry reads existing draft assets and rejects mismatched bytes; it never overwrites them. A published release is not modified on retry.
7. Independently download the public release, check SHA256SUMS and embedded BUILD evidence, and verify the macOS app signature, stapled ticket, Gatekeeper and CLI notarization. Cross-platform CI is distinct from real native interaction and real accounts.

Main protection requires a PR, current successful checks, resolved review conversations, and applies to administrators. Force pushes and branch deletion are disabled. Approvals follow the organization's single-maintainer setup (no mandatory second reviewer). Merge commits remain available for complete upstream intervals. Release-tag rules prevent moving or deleting `tiny-v*` tags.

The first releases remain experimental until the remaining [stable acceptance](compatibility.md#before-a-stable-release) is complete. There is no Tiny automatic application updater or registry/container publication in this workflow.
