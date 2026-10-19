#!/usr/bin/env bash
# Mount the real image and check what a user will copy into Applications.
set -euo pipefail
dmg=${1:?usage: test-macos-dmg.sh FILE.dmg [--notarized]}
work=$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/tiny-dmg-check.XXXXXXXX")
mounted=0
cleanup() {
  if [ "$mounted" = 1 ]; then hdiutil detach "$work/mount"; fi
  rm -rf -- "$work"
}
trap cleanup EXIT
mkdir "$work/mount"
hdiutil verify "$dmg"
hdiutil attach -readonly -nobrowse -mountpoint "$work/mount" "$dmg"
mounted=1
test "$(readlink "$work/mount/Applications")" = /Applications
for file in LICENSE UPSTREAM.json README.md; do
  cmp "$file" "$work/mount/$file"
done
cmp tiny-magpie.app/Contents/MacOS/tiny-magpie "$work/mount/tiny-magpie.app/Contents/MacOS/tiny-magpie"
cmp tiny-magpie.app/Contents/Info.plist "$work/mount/tiny-magpie.app/Contents/Info.plist"
if [ "${2:-}" = --notarized ]; then
  codesign --verify --deep --strict --verbose=2 "$work/mount/tiny-magpie.app"
  xcrun stapler validate "$work/mount/tiny-magpie.app"
  spctl --assess --type execute --verbose=2 "$work/mount/tiny-magpie.app"
fi
hdiutil detach "$work/mount"
mounted=0
