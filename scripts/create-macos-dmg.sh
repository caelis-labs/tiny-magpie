#!/usr/bin/env bash
# Same hdiutil UDZO layout as upstream: an app and an Applications shortcut.
set -euo pipefail
output=${1:?usage: create-macos-dmg.sh OUTPUT.dmg}
test ! -e "$output"
work=$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/tiny-dmg.XXXXXXXX")
trap 'rm -rf -- "$work"' EXIT
stage="$work/image"
mkdir -m 755 "$stage"
ditto tiny-magpie.app "$stage/tiny-magpie.app"
ln -s /Applications "$stage/Applications"
for file in LICENSE UPSTREAM.json README.md; do
  cp "$file" "$stage/"
  chmod 644 "$stage/$file"
done
hdiutil create -volname tiny-magpie -srcfolder "$stage" -format UDZO "$output"
