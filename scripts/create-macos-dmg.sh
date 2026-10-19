#!/usr/bin/env bash
# A two-icon Finder installer with one short bilingual hint and an arrow.
set -euo pipefail
output=${1:?usage: create-macos-dmg.sh OUTPUT.dmg}
test ! -e "$output"
work=$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/tiny-dmg.XXXXXXXX")
trap 'rm -rf -- "$work"' EXIT
source "$(dirname "${BASH_SOURCE[0]}")/dmg-tools.sh"
mkdir -m 755 "$work/docs"
for file in LICENSE UPSTREAM.json README.md; do
  cp "$file" "$work/docs/"
  chmod 644 "$work/docs/$file"
done
swift -module-cache-path "$tiny_dmg_root/.tiny/swift-cache" "$tiny_dmg_root/scripts/dmg-background.swift" "$work/background.tiff"
"$tiny_dmg_env/bin/dmgbuild" -s "$tiny_dmg_root/scripts/dmg-settings.py" -D "docs=$work/docs" -D "background=$work/background.tiff" tiny-magpie "$output"
