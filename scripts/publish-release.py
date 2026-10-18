#!/usr/bin/env python3
"""Reconcile immutable draft assets before making an experimental release public."""
import hashlib
import json
import pathlib
import re
import subprocess
import sys
import tempfile

REPO = 'caelis-labs/tiny-magpie'
ROOT = pathlib.Path(__file__).resolve().parent.parent
OUT = ROOT / 'output/release'


def gh(*args):
    return subprocess.check_output(['gh', *args], cwd=ROOT, text=True).strip()


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main(tag):
    if not re.fullmatch(r'tiny-v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', tag):
        raise SystemExit('Invalid tiny release tag')
    version = tag.removeprefix('tiny-v')
    sha = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()
    tagged = subprocess.check_output(['git', 'rev-parse', f'{tag}^{{commit}}'], cwd=ROOT, text=True).strip()
    if sha != tagged:
        raise SystemExit('Release checkout must equal tagged source')
    expected = {}
    for line in (OUT / 'SHA256SUMS').read_text().splitlines():
        checksum, name = line.split('  ', 1)
        if pathlib.Path(name).name != name or digest(OUT / name) != checksum:
            raise SystemExit(f'Invalid local checksum: {name}')
        expected[name] = checksum
    expected['SHA256SUMS'] = digest(OUT / 'SHA256SUMS')
    if set(expected) != {p.name for p in OUT.iterdir()}:
        raise SystemExit('Release files differ from checksum manifest')
    endpoint = f'repos/{REPO}/releases/tags/{tag}'
    result = subprocess.run(['gh', 'api', endpoint], capture_output=True, text=True)
    if result.returncode:
        if 'HTTP 404' not in result.stderr:
            raise SystemExit(result.stderr)
        gh('release', 'create', tag, '-R', REPO, '--verify-tag', '--target', sha,
           '--draft', '--prerelease', '--title', f'tiny-magpie {version} (Preview)',
           '--notes-file', str(ROOT / f'docs/releases/{version}.md'))
    release = json.loads(gh('api', endpoint))
    if release['target_commitish'] != sha or release['tag_name'] != tag:
        raise SystemExit('Existing release targets different source')
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
        gh('release', 'edit', tag, '-R', REPO, '--draft=false', '--prerelease', '--latest=false')
    final = json.loads(gh('api', endpoint))
    if final['draft'] or {asset['name'] for asset in final['assets']} != set(expected):
        raise SystemExit('Publication readback failed')
    print(final['html_url'])


if __name__ == '__main__':
    main(sys.argv[1])
