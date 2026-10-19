#!/usr/bin/env python3
"""Package only clean, source-stamped binaries; verify the complete release set."""
import argparse
import hashlib
import json
import os
import pathlib
import re
import shutil
import subprocess
import tempfile

ROOT = pathlib.Path(__file__).resolve().parent.parent
OUT = ROOT / 'output/release'
PLATFORMS = ['darwin-arm64', 'linux-amd64', 'linux-arm64', 'windows-amd64', 'windows-arm64']


def artifact_name(version, platform, kind):
    suffix = '.dmg' if platform.startswith('darwin-') and kind == 'desktop' else '.exe' if platform.startswith('windows-') else ''
    return f'tiny-magpie{"-cli" if kind == "cli" else ""}-{version}-{platform}{suffix}'


def run(*args):
    return subprocess.check_output(args, cwd=ROOT, text=True).strip()


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write_json(path, value):
    path.write_text(json.dumps(value, indent=2) + '\n', encoding='utf-8')


def source():
    if run('git', 'status', '--porcelain', '--untracked-files=no'):
        raise SystemExit('Release source has uncommitted changes')
    return run('git', 'rev-parse', 'HEAD')


def binary(path, sha, platform):
    details = run('go', 'version', '-m', str(path))
    system, arch = platform.split('-')
    settings = {line.strip().removeprefix('build\t') for line in details.splitlines()}
    if not {f'vcs.revision={sha}', 'vcs.modified=false', f'GOOS={system}', f'GOARCH={arch}'} <= settings:
        raise SystemExit(f'Binary source does not match clean checkout: {path}')
    return digest(path)


def binary_paths(platform):
    system = platform.split('-')[0]
    ext = '.exe' if system == 'windows' else ''
    return {
        'desktop': ROOT / ('tiny-magpie.app/Contents/MacOS/tiny-magpie' if system == 'darwin' else f'dist/tiny-magpie-{platform}{ext}'),
        'cli': ROOT / f'dist/tiny-magpie-cli-{platform}{ext}',
    }


def native_metadata(version, platform):
    sha = source()
    system = platform.split('-')[0]
    paths = binary_paths(platform)
    native = run('go', 'env', 'GOOS') + '-' + run('go', 'env', 'GOARCH')
    if platform != native:
        raise SystemExit(f'Package on a native runner: expected {platform}, got {native}')
    # -trimpath intentionally omits ldflags from Go build info (Go issue 52372).
    # Run both shipped binaries; a correct CLI version cannot prove the desktop's.
    with tempfile.TemporaryDirectory(prefix='tiny-version-') as temp:
        env = dict(os.environ)
        env.update({key: temp for key in ('HOME', 'USERPROFILE', 'APPDATA', 'LOCALAPPDATA', 'XDG_CONFIG_HOME', 'XDG_CACHE_HOME')})
        env.update(MAGPIE_NO_ACCOUNTS='1', DO_NOT_TRACK='1')
        for executable in paths.values():
            actual_version = subprocess.check_output([str(executable), 'version'], env=env, text=True, encoding='utf-8', timeout=30).strip()
            if actual_version != f'tiny-magpie {version}':
                raise SystemExit(f'Binary version does not match release: {executable}: {actual_version!r}')
    return {'version': version, 'source': sha, 'platform': platform,
                'upstream': json.loads((ROOT / 'UPSTREAM.json').read_text(encoding='utf-8')),
                'signing': 'Developer ID + notarization' if system == 'darwin' else 'unsigned',
                'binaries': {kind: binary(path, sha, platform) for kind, path in paths.items()},
                'artifacts': {}}


def package(version, platform):
    metadata = native_metadata(version, platform)
    paths = binary_paths(platform)
    system = platform.split('-')[0]
    OUT.mkdir(parents=True, exist_ok=True)
    if system == 'darwin':
        evidence = json.loads((OUT / f'NOTARIZATION-{platform}.json').read_text(encoding='utf-8'))
        verify_notarization(evidence, OUT / artifact_name(version, platform, 'desktop'), metadata)
    for kind in ('desktop', 'cli'):
        name = artifact_name(version, platform, kind)
        # The DMG already contains the stapled app and is signed/notarized itself.
        if system != 'darwin' or kind != 'desktop':
            shutil.copy2(paths[kind], OUT / name)
        metadata['artifacts'][name] = digest(OUT / name)
    write_json(OUT / f'BUILD-{platform}.json', metadata)


def verify_notarization(evidence, dmg, metadata):
    if (evidence.get('status') != 'Accepted' or not evidence.get('submission_id')
            or not evidence.get('dmg_submission_id') or not evidence.get('team_id')
            or evidence.get('app_ticket') != 'stapled' or evidence.get('dmg_ticket') != 'stapled'
            or evidence.get('binaries') != metadata['binaries']
            or evidence.get('dmg_sha256') != digest(dmg)):
        raise SystemExit('Missing or mismatched accepted app/CLI/DMG notarization')


def finalize(version):
    sha = source()
    expected = set()
    for platform in PLATFORMS:
        name = f'BUILD-{platform}.json'
        metadata = json.loads((OUT / name).read_text(encoding='utf-8'))
        if (metadata['version'], metadata['source'], metadata['platform']) != (version, sha, platform):
            raise SystemExit(f'Wrong build source: {name}')
        expected.add(name)
        system = platform.split('-')[0]
        artifacts = {artifact_name(version, platform, kind) for kind in ('desktop', 'cli')}
        if set(metadata['artifacts']) != artifacts or set(metadata['binaries']) != {'desktop', 'cli'}:
            raise SystemExit(f'Wrong artifact inventory: {name}')
        for kind in ('desktop', 'cli'):
            artifact = artifact_name(version, platform, kind)
            checksum = digest(OUT / artifact)
            if metadata['artifacts'].get(artifact) != checksum:
                raise SystemExit(f'Artifact hash mismatch: {artifact}')
            if not (system == 'darwin' and kind == 'desktop') and metadata['binaries'][kind] != checksum:
                raise SystemExit(f'Binary hash mismatch: {artifact}')
            expected.add(artifact)
        if system == 'darwin':
            name = f'NOTARIZATION-{platform}.json'
            evidence = json.loads((OUT / name).read_text(encoding='utf-8'))
            verify_notarization(evidence, OUT / artifact_name(version, platform, 'desktop'), metadata)
            expected.add(name)
    for name in ('LICENSE', 'UPSTREAM.json'):
        shutil.copy2(ROOT / name, OUT / name)
        expected.add(name)
    actual = {p.name for p in OUT.iterdir()}
    if actual != expected:
        raise SystemExit(f'Unexpected release inventory: missing={expected-actual}, extra={actual-expected}')
    (OUT / 'SHA256SUMS').write_text(''.join(f'{digest(OUT / name)}  {name}\n' for name in sorted(expected)), encoding='utf-8')
    print(f'Verified {len(expected)} assets, source {sha}; wrote SHA256SUMS')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', choices=['package', 'finalize', 'verify-native'])
    parser.add_argument('--version', required=True)
    parser.add_argument('--platform', choices=PLATFORMS)
    args = parser.parse_args()
    if not re.fullmatch(r'\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', args.version):
        parser.error('Invalid release version')
    if args.command in ('package', 'verify-native'):
        if not args.platform:
            parser.error('Native verification requires --platform')
        if args.command == 'package':
            package(args.version, args.platform)
        else:
            print(json.dumps(native_metadata(args.version, args.platform), indent=2))
    else:
        finalize(args.version)
