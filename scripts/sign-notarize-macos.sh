#!/usr/bin/env bash
# Release only: this bundle contains one Mach-O and no bundled frameworks.
set -euo pipefail
umask 077
platform=${1:?usage: sign-notarize-macos.sh darwin-arm64}
[[ "$platform" = darwin-arm64 ]]
[[ "$platform" = "darwin-$(go env GOARCH)" ]]
: "${VERSION:?missing release version}"
[[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]
: "${APPLE_DEVELOPER_ID_P12_BASE64:?missing Developer ID certificate}"
: "${APPLE_DEVELOPER_ID_P12_PASSWORD:?missing certificate password}"
: "${APPLE_NOTARY_APPLE_ID:?missing Notary Apple ID}"
: "${APPLE_NOTARY_PASSWORD:?missing Notary password}"
: "${APPLE_TEAM_ID:?missing Team ID}"
app=tiny-magpie.app
cli="dist/tiny-magpie-cli-$platform"
dmg="output/release/tiny-magpie-$VERSION-$platform.dmg"
mkdir -p output/notary output/release
work=$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/tiny-notary.XXXXXXXX")
keychain="$work/signing.keychain-db"
cleanup() {
  security delete-keychain "$keychain" >/dev/null 2>&1 || true
  rm -rf -- "$work"
}
trap cleanup EXIT
printf '%s' "$APPLE_DEVELOPER_ID_P12_BASE64" | base64 -D > "$work/identity.p12"
keychain_password=$(uuidgen)
security create-keychain -p "$keychain_password" "$keychain" >/dev/null
security unlock-keychain -p "$keychain_password" "$keychain" >/dev/null
security set-keychain-settings -lut 21600 "$keychain"
security import "$work/identity.p12" -k "$keychain" -P "$APPLE_DEVELOPER_ID_P12_PASSWORD" -T /usr/bin/codesign >/dev/null
security set-key-partition-list -S apple-tool:,apple: -s -k "$keychain_password" "$keychain" >/dev/null
security list-keychains -d user -s "$keychain"
identity=$(security find-identity -v -p codesigning "$keychain" | sed -n '/Developer ID Application:/s/.*\([A-F0-9]\{40\}\).*/\1/p' | head -1)
test -n "$identity"
codesign --force --timestamp --options runtime --keychain "$keychain" --sign "$identity" "$app/Contents/MacOS/tiny-magpie"
codesign --force --timestamp --options runtime --keychain "$keychain" --sign "$identity" "$cli"
codesign --force --timestamp --options runtime --keychain "$keychain" --sign "$identity" "$app"
for item in "$app" "$cli"; do
  codesign --verify --strict --verbose=2 "$item"
  details=$(codesign -dv --verbose=4 "$item" 2>&1)
  printf '%s\n' "$details" | grep -q 'flags=.*runtime'
  printf '%s\n' "$details" | grep -q '^Timestamp='
  test "$(printf '%s\n' "$details" | sed -n 's/^TeamIdentifier=//p')" = "$APPLE_TEAM_ID"
done
mkdir "$work/submission"
ditto "$app" "$work/submission/$app"
cp "$cli" "$work/submission/"
ditto -c -k --sequesterRsrc --keepParent "$work/submission" "$work/submission.zip"
cp "$work/submission.zip" "output/notary/$platform-submission.zip"
notary=(--apple-id "$APPLE_NOTARY_APPLE_ID" --password "$APPLE_NOTARY_PASSWORD" --team-id "$APPLE_TEAM_ID")
# Persist the upload receipt before waiting. A timeout is an unknown outcome;
# inspect this ID rather than automatically resubmitting the signed bytes.
notarize() {
  local file=$1 receipt=$2 submission_id
  xcrun notarytool submit "$file" "${notary[@]}" --output-format json > "output/notary/$receipt-submit.json"
  submission_id=$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["id"])' "output/notary/$receipt-submit.json")
  xcrun notarytool wait "$submission_id" "${notary[@]}" --timeout 45m --output-format json > "output/notary/$receipt-status.json"
  python3 - "$submission_id" "output/notary/$receipt-status.json" <<'PY'
import json, sys
status = json.load(open(sys.argv[2]))
if status.get('status') != 'Accepted' or status.get('id') != sys.argv[1]:
    raise SystemExit(f'Notarization failed: {status}')
PY
}
notarize "$work/submission.zip" "$platform"
xcrun stapler staple "$app"
xcrun stapler validate "$app"
codesign --verify --deep --strict --verbose=2 "$app"
spctl --assess --type execute --verbose=2 "$app"
codesign -vvvv -R=notarized --check-notarization "$cli"
# Build the disk image from the already stapled app so its installed copy
# also works offline. The outer image gets its own signature and ticket.
bash scripts/create-macos-dmg.sh "$dmg"
codesign --timestamp --keychain "$keychain" --sign "$identity" "$dmg"
codesign --verify --strict --verbose=2 "$dmg"
test "$(codesign -dv --verbose=4 "$dmg" 2>&1 | sed -n 's/^TeamIdentifier=//p')" = "$APPLE_TEAM_ID"
cp "$dmg" "output/notary/$platform-dmg-submission.dmg"
notarize "$dmg" "$platform-dmg"
xcrun stapler staple "$dmg"
xcrun stapler validate "$dmg"
codesign --verify --strict --verbose=2 "$dmg"
spctl --assess --type open --context context:primary-signature --verbose=2 "$dmg"
bash scripts/test-macos-dmg.sh "$dmg" --notarized
# Only expose the successful evidence after all signature/ticket checks pass.
python3 - "$platform" "$APPLE_TEAM_ID" "$dmg" "$app/Contents/MacOS/tiny-magpie" "$cli" <<'PY'
import hashlib, json, pathlib, sys
platform, team, dmg, app_binary, cli = sys.argv[1:]
def digest(path):
    return hashlib.sha256(pathlib.Path(path).read_bytes()).hexdigest()
app_status = json.loads(pathlib.Path(f'output/notary/{platform}-status.json').read_text())
dmg_status = json.loads(pathlib.Path(f'output/notary/{platform}-dmg-status.json').read_text())
pathlib.Path(f'output/release/NOTARIZATION-{platform}.json').write_text(json.dumps({
    'submission_id': app_status['id'], 'dmg_submission_id': dmg_status['id'],
    'status': 'Accepted', 'team_id': team, 'hardened_runtime': True,
    'submission_sha256': digest(f'output/notary/{platform}-submission.zip'),
    'dmg_submission_sha256': digest(f'output/notary/{platform}-dmg-submission.dmg'),
    'dmg_sha256': digest(dmg),
    'binaries': {'desktop': digest(app_binary), 'cli': digest(cli)},
    'app_ticket': 'stapled', 'dmg_ticket': 'stapled',
    'cli_ticket': 'online (Mach-O cannot be stapled)',
}, indent=2) + '\n')
PY
