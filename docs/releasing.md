# Releases

Only `caelis-labs/tiny-magpie` is a release destination. Upstream Magpie release channels remain disabled. The source and modifications remain MIT licensed; retain the original copyright notice and Git history.

Tiny tags use `tiny-vX.Y.Z` so they do not collide with inherited upstream tags. Do not push upstream tags to origin. A release version PR must pass all required checks before merging into main. The Release tiny-magpie workflow reruns the full platform, protocol/UI and container gates at the exact tagged source before packaging.

## Routine release with release-please

Merge Conventional Commit PRs (`fix:`, `feat:`) into main. The Release please
workflow uses the organization `RELEASE_PLEASE_TOKEN`, granted to this repository,
to open or update a version PR containing `CHANGELOG.md` and
`.release-please-manifest.json`. Its initial baseline is the published 0.3.0.
The token must be able to write repository contents and pull requests; the
default Actions token cannot trigger the version PR's CI or the tag workflow.
No secret value is copied into this repository.

Review the generated version and notes, wait for all six required checks, then
merge that PR. release-please creates a draft and its `tiny-v*` tag
(`component: tiny`, separator `-`, `force-tag-creation: true`). The tag starts
Release tiny-magpie once. release-please does not publish the empty draft:
the existing publisher reconciles it by ID, validates all 19 immutable assets
and only then publishes it as Latest. Signing, notarization, integrity and
checkpoint recovery stay in the existing workflow. If the main-triggered
release-please run is interrupted, rerun it or dispatch Release please on main;
never recreate an existing tag or rebuild existing candidate bytes.

The generated draft contains the reviewed changelog notes; routine releases
need no hand-written tag or second release dispatch. `docs/releases/` remains
available for curated notes and manual recovery when no draft exists.

## Packaging and recovery

1. Review the version PR's changelog (or add `docs/releases/X.Y.Z.md` for a manual release), verify UPSTREAM.json, and merge through the protected main branch.
2. Wait for the six required checks: three OS Go suites/native builds, protocols-and-ui, and two architecture container checks. Do not tag a cancelled or failing head.
3. Normally release-please creates the `tiny-vX.Y.Z` tag at the merged version PR commit. For manual recovery only, create and push an annotated tag at that exact checked main commit. The tag triggers release. Once candidate artifacts exist, rerun only failed jobs (`gh run rerun RUN_ID --failed`) to reuse their original bytes. A fresh workflow_dispatch rebuilds and signs again; it must not replace existing draft assets or replay an unknown notarization submission.
4. The workflow builds desktop and CLI binaries on five native runners: Apple Silicon macOS, amd64/arm64 Windows and amd64/arm64 Linux. macOS uses shared organization secrets `APPLE_DEVELOPER_ID_P12_BASE64`, `APPLE_DEVELOPER_ID_P12_PASSWORD`, `APPLE_NOTARY_APPLE_ID`, `APPLE_NOTARY_PASSWORD` and organization variable `APPLE_TEAM_ID`. No repository copy of private material is needed. Signing uses a temporary runner keychain, hardened runtime and secure timestamps; no additional entitlements are introduced. Windows/Linux are unsigned.
5. macOS submits the signed app and CLI together, persists the signed archive and receipt, waits for Accepted, staples and validates the app, checks Gatekeeper and verifies the CLI's online ticket. It then creates a UDZO DMG containing that stapled app, an Applications shortcut and license/source documents, signs and notarizes the DMG, and staples and validates its ticket. The workflow mounts the final image and verifies the app inside it. Signature, team, notarization or ticket failures stop publication. A timeout is unknown: inspect the uploaded notary-receipt artifact, retain its submitted bytes and query its submission ID before resubmitting. Neither signing nor unknown submissions are blindly retried.
6. Each runner executes both the desktop and CLI `version` command in an empty profile. Packaging checks embedded Go source SHA, clean source and target architecture; `-trimpath` omits linker flags from Go build metadata, so version is checked by execution. Per-platform BUILD evidence records hashes for both binaries and both download assets. The macOS receipt binds its accepted submissions to the final DMG and binary hashes. Publishing checks the exact 19-asset inventory, creates a draft, downloads and verifies every uploaded asset, then publishes it. From 0.2.0, plain versions use `publish-release.py --stable` and become Latest; SemVer suffix versions remain Preview. A retry reads existing draft assets and rejects mismatched bytes or channels; it never overwrites them. A published release is not modified on retry.
7. Independently download the public release, check SHA256SUMS and BUILD evidence, mount the DMG, and verify the app/DMG signatures, stapled tickets, Gatekeeper and CLI notarization. Cross-platform CI is distinct from real native interaction and real accounts.

Main protection requires a PR, current successful checks, resolved review conversations, and applies to administrators. Force pushes and branch deletion are disabled. Approvals follow the organization's single-maintainer setup (no mandatory second reviewer). Merge commits remain available for complete upstream intervals. Release-tag rules prevent moving or deleting `tiny-v*` tags.

The publisher resolves existing drafts with `gh release view` and reads them through their release ID. GitHub's REST tag endpoint returns published releases only; a draft's 404 there must not cause another draft to be created. `scripts/test-publish-release.py` exercises that contract, unknown creation/upload outcomes, published retries and mismatched assets.

Apple ticket staple/validate calls use `scripts/notarization-stapler.py`: at most three attempts, delays of 5/15 seconds and a 120-second attempt timeout. Only exit 68 with explicit `NSURLErrorDomain` -1001/-1004/-1005 transport diagnostics qualifies. Unknown errors, TLS, signature/trust failures and mixed diagnostics fail closed. The helper never submits notarization or signs code.

If the app/CLI submission is Accepted but ticket validation fails before DMG creation, dispatch Release on main with the existing tag and `resume_run_id`. Recovery verifies the original same-repository tag run, all six successful gates and four successful Windows/Linux package jobs, then reuses those original artifacts. It restores the signed app/CLI archive without building or signing either binary again, checks native source/version/signatures, queries the same Apple submission ID and resumes ticket validation. Only the new DMG needs signing and its first submission. A checkpoint already containing a DMG submission is rejected for manual inspection; it must not be blindly resubmitted. The original tag remains immutable. A branch dispatch may prepare recovery artifacts but cannot publish; publication resumes on main or is reconciled from the verified artifacts against the original tagged checkout.

The maintainer accepted the gateway-core Dev for the first stable release, 0.2.0. Release notes must retain the [acceptance limits](compatibility.md#before-a-stable-release): fixtures and cross-platform CI do not prove every vendor account or native Windows/Linux interaction. The application checks Tiny releases through its own GitHub API channel; see [the updater contract](subsystems/gui-shell.md#tiny-installation-identity-and-self-update). Registry/container publication is outside this workflow.

PR validation runs once per reviewed branch; release tags run the [retained matrix](test-scope.md#ci-execution) again on the actual source. Main pushes do not repeat it. Routine Windows validation compiles every test and executes the native regression gate; the complete Windows corpus runs weekly.

## Artifact formats from 0.1.2

The matrix is a subset of [upstream's release matrix](https://github.com/yetone/magpie/releases/tag/v0.1.1290), with the same download formats. Intel macOS, macOS ZIP and Android CLI outputs are omitted.

| Product | Architectures | Download |
| --- | --- | --- |
| macOS desktop | arm64 (Apple Silicon) | `tiny-magpie-VERSION-darwin-arm64.dmg` |
| Windows desktop | amd64, arm64 | `tiny-magpie-VERSION-windows-ARCH.exe` |
| Linux desktop | amd64, arm64 | `tiny-magpie-VERSION-linux-ARCH` |
| Headless CLI | the same five platform pairs | `tiny-magpie-cli-VERSION-OS-ARCH` (`.exe` on Windows) |

The release has ten downloads, five BUILD JSON files, one NOTARIZATION JSON file, LICENSE, UPSTREAM.json and SHA256SUMS: 19 assets. The DMG includes LICENSE, UPSTREAM.json and README.md in its hidden `.documentation` folder, leaving only the app and Applications visible. Keep the release's MIT LICENSE with redistributed standalone binaries.

On macOS, open the DMG, drag `tiny-magpie.app` along the arrow to Applications, then launch it from Applications and eject the disk. Both the disk image and its app have stapled Apple tickets. Windows `.exe` files are portable applications, not setup installers; run the matching architecture directly (the system WebView2 runtime is required). On Linux, run `chmod +x tiny-magpie-VERSION-linux-ARCH` before launching it; the desktop requires GTK 3 and WebKitGTK 4.1 on an Ubuntu 24.04-compatible system (`libgtk-3-0t64`, `libwebkit2gtk-4.1-0`). Standalone macOS/Linux CLI downloads also need `chmod +x` after downloading through a browser.

CLI remains a supported headless gateway and Web GUI entry point (`serve`, `web --no-open`), plus scriptable Provider/account, routing group, gateway key, quota and plugin management. It has no desktop WebView dependency and is useful on servers and in containers. The desktop binary shares those command handlers; removing `internal/cli` would also remove the desktop's entry point. `make cli` and the container image retain the upstream `nogui` build. Claude's private subscription CLI/bridge is a separate retained Provider dependency.

The native PR build executes both linked versions on all three operating systems and creates/mounts an unsigned macOS DMG to check its contents without release secrets. Offline inventory tests reject extra Intel/archive files, changed or missing downloads, wrong-source builds and receipts for different bytes. Public signing/notarization runs only for release tags. The historical 0.1.1 archive release is preserved; new formats use a new version. Stable installer acceptance remains a separate gate.


## Tiny self-update

From the installer/update revision onward, released desktop builds check only
`caelis-labs/tiny-magpie` at startup and every six hours. General settings offer
Check, automatic checking and preview/stable. Stable is the default from 0.2.0;
an explicitly saved Preview choice is preserved and includes stable releases too. Downloads and restart
require separate clicks; active gateway requests are waited for. Stable does
not accept previews or downgrade an installed version. Developer builds do not
poll automatically. A 0.1.2 installation has no active updater and needs one
manual installation of the new DMG; subsequent updates use the Tiny channel.

The public artifact matrix remains unchanged: the updater consumes the same
Apple Silicon DMG or Windows/Linux executable users can download. GitHub asset
digests are required, and macOS additionally checks team, bundle identity,
version and Gatekeeper. No updater-only ZIP, external update service or
upstream app feed is introduced. Source-channel errors remain visible and
retryable; they never permit unchecked replacement. Native installer and
update acceptance are distinct from cross-platform fixture/CI evidence.
