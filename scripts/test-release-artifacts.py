#!/usr/bin/env python3
"""Offline release inventory/provenance failures; no Apple or GitHub access."""
import importlib.util
import json
import pathlib
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('artifacts', pathlib.Path(__file__).with_name('release-artifacts.py'))
artifacts = importlib.util.module_from_spec(spec)
spec.loader.exec_module(artifacts)
VERSION = '0.1.2'
SHA = 'a' * 40
# Expected public names are independent of the packager's name generator.
DOWNLOADS = {
    'darwin-arm64': ('tiny-magpie-0.1.2-darwin-arm64.dmg', 'tiny-magpie-cli-0.1.2-darwin-arm64'),
    'linux-amd64': ('tiny-magpie-0.1.2-linux-amd64', 'tiny-magpie-cli-0.1.2-linux-amd64'),
    'linux-arm64': ('tiny-magpie-0.1.2-linux-arm64', 'tiny-magpie-cli-0.1.2-linux-arm64'),
    'windows-amd64': ('tiny-magpie-0.1.2-windows-amd64.exe', 'tiny-magpie-cli-0.1.2-windows-amd64.exe'),
    'windows-arm64': ('tiny-magpie-0.1.2-windows-arm64.exe', 'tiny-magpie-cli-0.1.2-windows-arm64.exe'),
}


class Inventory(unittest.TestCase):
    def setUp(self):
        temp = tempfile.TemporaryDirectory(prefix='tiny-artifact-test-')
        self.addCleanup(temp.cleanup)
        self.root = pathlib.Path(temp.name)
        self.out = self.root / 'output'
        self.out.mkdir()
        for name in ('LICENSE', 'UPSTREAM.json'):
            (self.root / name).write_text('fixture ' + name)
        for target, names in DOWNLOADS.items():
            hashes = {}
            for kind, name in zip(('desktop', 'cli'), names):
                (self.out / name).write_bytes(name.encode())
                hashes[kind] = artifacts.digest(self.out / name)
            artifacts.write_json(self.out / f'BUILD-{target}.json', {
                'version': VERSION, 'source': SHA, 'platform': target,
                'binaries': hashes,
                'artifacts': {name: artifacts.digest(self.out / name) for name in names},
            })
        self.receipt = self.out / 'NOTARIZATION-darwin-arm64.json'
        artifacts.write_json(self.receipt, {
            'status': 'Accepted', 'submission_id': 'app-id', 'dmg_submission_id': 'dmg-id',
            'team_id': 'team', 'app_ticket': 'stapled', 'dmg_ticket': 'stapled',
            'binaries': json.loads((self.out / 'BUILD-darwin-arm64.json').read_text())['binaries'],
            'dmg_sha256': artifacts.digest(self.out / DOWNLOADS['darwin-arm64'][0]),
        })
        for name, value in (('ROOT', self.root), ('OUT', self.out), ('source', lambda: SHA)):
            context = patch.object(artifacts, name, value)
            context.start()
            self.addCleanup(context.stop)

    def test_five_native_pairs_and_standalone_cli(self):
        artifacts.finalize(VERSION)
        expected = {name for names in DOWNLOADS.values() for name in names}
        expected |= {f'BUILD-{target}.json' for target in DOWNLOADS}
        expected |= {'LICENSE', 'UPSTREAM.json', self.receipt.name}
        entries = (self.out / 'SHA256SUMS').read_text().splitlines()
        self.assertEqual({line.split('  ')[1] for line in entries}, expected)
        self.assertEqual(len(entries), 18)  # Plus the manifest itself: 19 assets.
        for line in entries:
            checksum, name = line.split('  ')
            self.assertEqual(checksum, artifacts.digest(self.out / name))

    def test_old_intel_and_archive_outputs_are_rejected(self):
        for name in ('tiny-magpie-0.1.2-darwin-amd64.dmg', 'tiny-magpie-0.1.2-windows-amd64.zip'):
            with self.subTest(name=name):
                (self.out / name).write_bytes(b'old release output')
                with self.assertRaisesRegex(SystemExit, 'Unexpected release inventory'):
                    artifacts.finalize(VERSION)
                (self.out / name).unlink()

    def test_changed_or_missing_download_fails(self):
        path = self.out / DOWNLOADS['linux-arm64'][0]
        path.write_bytes(b'different bytes')
        with self.assertRaisesRegex(SystemExit, 'Artifact hash mismatch'):
            artifacts.finalize(VERSION)
        path.unlink()
        with self.assertRaises(FileNotFoundError):
            artifacts.finalize(VERSION)

    def test_notary_receipt_must_cover_exact_dmg_and_binaries(self):
        original = json.loads(self.receipt.read_text())
        for field, value in (('dmg_sha256', 'wrong'), ('binaries', {}), ('dmg_submission_id', ''), ('app_ticket', ''), ('status', 'Invalid')):
            with self.subTest(field=field):
                artifacts.write_json(self.receipt, dict(original, **{field: value}))
                with self.assertRaisesRegex(SystemExit, 'notarization'):
                    artifacts.finalize(VERSION)

    def test_wrong_source_and_binary_architecture_fail(self):
        with self.assertRaisesRegex(SystemExit, 'Wrong build source'):
            artifacts.finalize('0.1.3')
        path = self.out / DOWNLOADS['linux-amd64'][0]
        for settings in (f'vcs.revision={SHA}\nvcs.modified=true\nGOOS=linux\nGOARCH=amd64',
                         f'vcs.revision={SHA}\nvcs.modified=false\nGOOS=linux\nGOARCH=arm64'):
            with patch.object(artifacts, 'run', return_value=settings):
                with self.assertRaisesRegex(SystemExit, 'Binary source'):
                    artifacts.binary(path, SHA, 'linux-amd64')


if __name__ == '__main__':
    unittest.main()
