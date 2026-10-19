# Upstream maintenance

Tiny is a reduction fork of [yetone/magpie](https://github.com/yetone/magpie), with its complete Git history and MIT notices. `UPSTREAM.json` records the exact integrated SHA and retained built-in assembly. The Go module and internal layout deliberately remain upstream's to keep merges small. Tiny's binary, OS registrations and data directories are independent. Build with the documented Make targets: the inherited `make install` target is removed because raw `go install .` uses the upstream module's command name.

Use tested main commits, normally weekly; bring protocol, authentication and security fixes forward sooner. Merge a complete upstream interval rather than accumulating unrelated cherry-picks. Record the source SHA of an urgent cherry-pick, then reconcile it in the next merge. Do not accept a whole conflicted file with `ours`, resurrect a removed service to clear a conflict, or weaken a retained regression.

1. Start with a clean tiny branch; fetch upstream main and choose a full SHA.
2. Run `scripts/sync-upstream.sh <sha>`. It fetches and merges without committing or pushing. Resolve conflicts against [the scope](docs/tiny-scope.md).
3. Check new routes, CLI dispatch, startup goroutines, built-in registration, plugin hooks, model metadata, GUI navigation, OS registrations and data-directory migrations. Deleted modules stay removed; a new upstream product feature is not automatically in scope.
4. Update `UPSTREAM.json` only once its source has actually been integrated. Run the scope check, [retained test gates](docs/test-scope.md), GUI checks and binary differential against that SHA. New tests for removed products require scope review.
5. Commit with the upstream SHA and actual validation results. Experimental releases follow [the release gates](docs/releasing.md); stable releases also require the remaining platform/live acceptance.

`upstream-check.yml` runs Mondays at 02:00 UTC (10:00 Asia/Shanghai), uploads a commit/file inventory, and maintains one `Upstream sync` issue when new commits exist. It leaves the issue untouched while upstream is unchanged and closes it once the recorded base catches up. It does not merge, push, install or publish. Application self-update from Magpie's feed is disabled; provider-plugin and Bun update mechanisms are retained.
