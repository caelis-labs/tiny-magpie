#!/usr/bin/env python3
"""Reconcile immutable draft assets before publishing a Tiny release."""
import argparse
import hashlib
import json
import pathlib
import re
import subprocess
import tempfile

REPO = 'caelis-labs/tiny-magpie'
ROOT = pathlib.Path(__file__).resolve().parent.parent
OUT = ROOT / 'output/release'


def gh(*args):
    return subprocess.check_output(['gh', *args], cwd=ROOT, text=True).strip()


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main(tag, stable=False):
    if not re.fullmatch(r'tiny-v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', tag):
        raise SystemExit('Invalid tiny release tag')
    version = tag.removeprefix('tiny-v')
    if stable and '-' in version:
        raise SystemExit('A prerelease version cannot be published as stable')
    sha = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()
    tagged = subprocess.check_output(['git', 'rev-parse', f'{tag}^{{commit}}'], cwd=ROOT, text=True).strip()
    if sha != tagged:
        raise SystemExit('Release checkout must equal tagged source')
    expected = {}
    for line in (OUT / 'SHA256SUMS').read_text(encoding='utf-8').splitlines():
        checksum, name = line.split('  ', 1)
        if pathlib.Path(name).name != name or digest(OUT / name) != checksum:
            raise SystemExit(f'Invalid local checksum: {name}')
        expected[name] = checksum
    expected['SHA256SUMS'] = digest(OUT / 'SHA256SUMS')
    if set(expected) != {p.name for p in OUT.iterdir()}:
        raise SystemExit('Release files differ from checksum manifest')
    # The REST tag endpoint excludes drafts. The CLI resolves either state;
    # use its stable ID for every subsequent REST read.
    lookup = ('release', 'view', tag, '-R', REPO, '--json', 'databaseId')
    result = subprocess.run(['gh', *lookup], capture_output=True, text=True)
    if result.returncode:
        if 'HTTP 404' not in result.stderr and result.stderr.strip() != 'release not found':
            raise SystemExit(result.stderr)
        gh('release', 'create', tag, '-R', REPO, '--verify-tag', '--target', sha,
           '--draft', f'--prerelease={str(not stable).lower()}', '--title',
           f'tiny-magpie {version}' + ('' if stable else ' (Preview)'),
           '--notes-file', str(ROOT / f'docs/releases/{version}.md'))
    release_id = json.loads(gh(*lookup))['databaseId']
    endpoint = f'repos/{REPO}/releases/{release_id}'
    release = json.loads(gh('api', endpoint))
    if release['target_commitish'] != sha or release['tag_name'] != tag:
        raise SystemExit('Existing release targets different source')
    if release['prerelease'] != (not stable):
        raise SystemExit('Existing release has a different channel')
    present = {asset['name'] for asset in release['assets']}
    if present - expected.keys():
        raise SystemExit(f'Unexpected remote assets: {present - expected.keys()}')
    for name, checksum in expected.items():
        if name not in present:
            if not release['draft']:
                raise SystemExit(f'Published release is incomplete: {name}')
            # An unknown upload outcome is reconciled on the next run by reading
            # this same name; assets are never overwritten or silently replaced.
            gh('release', 'upload', tag, str(OUT / name), '-R', REPO)
        with tempfile.TemporaryDirectory(prefix='tiny-readback-') as temp:
            gh('release', 'download', tag, '-R', REPO, '--pattern', name, '--dir', temp)
            if digest(pathlib.Path(temp) / name) != checksum:
                raise SystemExit(f'Remote asset differs: {name}')
    if release['draft']:
        gh('release', 'edit', tag, '-R', REPO, '--draft=false',
           f'--prerelease={str(not stable).lower()}', f'--latest={str(stable).lower()}')
    final = json.loads(gh('api', endpoint))
    if final['draft'] or final['prerelease'] != (not stable) or {asset['name'] for asset in final['assets']} != set(expected):
        raise SystemExit('Publication readback failed')
    print(final['html_url'])


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag')
    parser.add_argument('--stable', action='store_true')
    args = parser.parse_args()
    main(args.tag, args.stable)
