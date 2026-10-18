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
import tarfile
import tempfile
import zipfile

ROOT = pathlib.Path(__file__).resolve().parent.parent
OUT = ROOT / 'output/release'
PLATFORMS = [f'{os}-{arch}' for os in ('darwin', 'linux', 'windows') for arch in ('amd64', 'arm64')]


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


def binary(path, sha):
    details = run('go', 'version', '-m', str(path))
    if f'vcs.revision={sha}' not in details or 'vcs.modified=false' not in details:
        raise SystemExit(f'Binary source does not match clean checkout: {path}')
    return digest(path)


def package(version, platform):
    sha = source()
    system = platform.split('-')[0]
    ext = '.exe' if system == 'windows' else ''
    cli = ROOT / f'dist/tiny-magpie-cli-{platform}{ext}'
    desktop = ROOT / ('tiny-magpie.app/Contents/MacOS/tiny-magpie' if system == 'darwin' else f'dist/tiny-magpie-{platform}{ext}')
    native = run('go', 'env', 'GOOS') + '-' + run('go', 'env', 'GOARCH')
    if platform != native:
        raise SystemExit(f'Package on a native runner: expected {platform}, got {native}')
    # -trimpath intentionally omits ldflags from Go build info (Go issue 52372).
    # Read the actual linked version by running the native CLI in a fresh profile.
    with tempfile.TemporaryDirectory(prefix='tiny-version-') as temp:
        env = dict(os.environ)
        env.update({key: temp for key in ('HOME', 'USERPROFILE', 'APPDATA', 'LOCALAPPDATA', 'XDG_CONFIG_HOME', 'XDG_CACHE_HOME')})
        env.update(MAGPIE_NO_ACCOUNTS='1', DO_NOT_TRACK='1')
        actual_version = subprocess.check_output([str(cli), 'version'], env=env, text=True, encoding='utf-8', timeout=30).strip()
        if actual_version != f'tiny-magpie {version}':
            raise SystemExit(f'Binary version does not match release: {actual_version!r}')
    metadata = {'version': version, 'source': sha, 'platform': platform,
                'upstream': json.loads((ROOT / 'UPSTREAM.json').read_text(encoding='utf-8')),
                'signing': 'Developer ID + notarization' if system == 'darwin' else 'unsigned',
                'binaries': {'desktop': binary(desktop, sha), 'cli': binary(cli, sha)},
                'archives': {}}
    OUT.mkdir(parents=True, exist_ok=True)
    if system == 'darwin':
        evidence = json.loads((OUT / f'NOTARIZATION-{platform}.json').read_text(encoding='utf-8'))
        if evidence.get('status') != 'Accepted':
            raise SystemExit('Missing accepted notarization')
    for kind in ('desktop', 'cli'):
        suffix = '.zip' if system == 'windows' or (system == 'darwin' and kind == 'desktop') else '.tar.gz'
        name = f'tiny-magpie{"-cli" if kind == "cli" else ""}-{version}-{platform}{suffix}'
        with tempfile.TemporaryDirectory(prefix='tiny-package-') as temp:
            stage = pathlib.Path(temp) / 'tiny-magpie'
            stage.mkdir()
            for file in ('LICENSE', 'UPSTREAM.json', 'README.md'):
                shutil.copy2(ROOT / file, stage / file)
            write_json(stage / 'BUILD.json', {k: v for k, v in metadata.items() if k != 'archives'})
            if system == 'darwin' and kind == 'desktop':
                subprocess.run(['ditto', str(ROOT / 'tiny-magpie.app'), str(stage / 'tiny-magpie.app')], check=True)
                subprocess.run(['ditto', '-c', '-k', '--sequesterRsrc', '--keepParent', str(stage), str(OUT / name)], check=True)
            else:
                shutil.copy2(cli if kind == 'cli' else desktop, stage / ('tiny-magpie' + ext))
                if suffix == '.zip':
                    with zipfile.ZipFile(OUT / name, 'w', zipfile.ZIP_DEFLATED) as archive:
                        for file in sorted(stage.iterdir()):
                            archive.write(file, 'tiny-magpie/' + file.name)
                else:
                    with tarfile.open(OUT / name, 'w:gz') as archive:
                        archive.add(stage, arcname='tiny-magpie')
        metadata['archives'][name] = digest(OUT / name)
    write_json(OUT / f'BUILD-{platform}.json', metadata)


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
        for kind in ('desktop', 'cli'):
            suffix = '.zip' if system == 'windows' or (system == 'darwin' and kind == 'desktop') else '.tar.gz'
            archive = f'tiny-magpie{"-cli" if kind == "cli" else ""}-{version}-{platform}{suffix}'
            if metadata['archives'].get(archive) != digest(OUT / archive):
                raise SystemExit(f'Archive hash mismatch: {archive}')
            expected.add(archive)
        if system == 'darwin':
            name = f'NOTARIZATION-{platform}.json'
            evidence = json.loads((OUT / name).read_text(encoding='utf-8'))
            if evidence.get('status') != 'Accepted' or not evidence.get('submission_id'):
                raise SystemExit(f'Notarization missing: {name}')
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
    parser.add_argument('command', choices=['package', 'finalize'])
    parser.add_argument('--version', required=True)
    parser.add_argument('--platform', choices=PLATFORMS)
    args = parser.parse_args()
    if not re.fullmatch(r'\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', args.version):
        parser.error('Invalid release version')
    if args.command == 'package':
        if not args.platform:
            parser.error('package requires --platform')
        package(args.version, args.platform)
    else:
        finalize(args.version)
