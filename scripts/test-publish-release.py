#!/usr/bin/env python3
"""Exercise the publisher against GitHub's draft and immutable-asset contracts."""
import contextlib
import importlib.util
import io
import json
import pathlib
import subprocess
import tempfile
import unittest
from types import SimpleNamespace
from unittest.mock import patch


class PublishRecovery(unittest.TestCase):
    def setUp(self):
        spec = importlib.util.spec_from_file_location(
            'publisher', pathlib.Path(__file__).with_name('publish-release.py'))
        self.publisher = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(self.publisher)
        temp = self.enterContext(tempfile.TemporaryDirectory())
        self.publisher.OUT = pathlib.Path(temp)
        asset = self.publisher.OUT / 'test.zip'
        asset.write_bytes(b'fixture signed bytes')
        (self.publisher.OUT / 'SHA256SUMS').write_text(
            self.publisher.digest(asset) + '  test.zip\n', encoding='utf-8')
        self.sha, self.tag = '1' * 40, 'tiny-v0.1.1'
        self.exists, self.draft = False, True
        self.prerelease = True
        self.latest = False
        self.assets, self.uploads = {}, {}
        self.creates, self.edits = 0, 0
        self.interrupt_create = self.interrupt_upload = False
        self.lookup_error = ''
        self.enterContext(patch.object(self.publisher, 'gh', self.gh))
        self.enterContext(patch.object(self.publisher.subprocess, 'run', self.lookup))
        self.enterContext(patch.object(self.publisher.subprocess, 'check_output',
                                      lambda *args, **kwargs: self.sha))
        self.enterContext(contextlib.redirect_stdout(io.StringIO()))

    def snapshot(self):
        return {'id': 4242, 'target_commitish': self.sha, 'tag_name': self.tag,
                'draft': self.draft, 'prerelease': self.prerelease,
                'assets': [{'name': n} for n in self.assets],
                'html_url': 'fixture://published'}

    def lookup(self, args, **kwargs):
        if self.lookup_error:
            return SimpleNamespace(returncode=1, stdout='', stderr=self.lookup_error)
        by_cli = args[1:3] == ['release', 'view']
        if by_cli:
            found = {'databaseId': 4242} if self.exists else None
        else:
            # GitHub's REST tag endpoint returns published releases only.
            found = self.snapshot() if self.exists and not self.draft else None
        return SimpleNamespace(returncode=0 if found else 1,
                               stdout=json.dumps(found) if found else '',
                               stderr='' if found else (
                                   'release not found' if by_cli else 'gh: Not Found (HTTP 404)'))

    def gh(self, *args):
        if args[0] == 'api':
            by_id = args[1].endswith('/releases/4242')
            if self.exists and (by_id or not self.draft):
                return json.dumps(self.snapshot())
            raise subprocess.CalledProcessError(1, args, stderr='HTTP 404')
        action = args[1]
        if action == 'view':
            return json.dumps({'databaseId': 4242})
        if action == 'create':
            self.assertFalse(self.exists, 'an existing draft must not be recreated')
            self.exists = True
            self.prerelease = '--prerelease=true' in args
            self.creates += 1
            if self.interrupt_create:
                self.interrupt_create = False
                raise RuntimeError('unknown create response after remote accepted draft')
        elif action == 'upload':
            path = pathlib.Path(args[3])
            self.assertNotIn(path.name, self.assets, 'existing bytes must not be overwritten')
            self.assets[path.name] = path.read_bytes()
            self.uploads[path.name] = self.uploads.get(path.name, 0) + 1
            if self.interrupt_upload:
                self.interrupt_upload = False
                raise RuntimeError('unknown upload response after remote accepted bytes')
        elif action == 'download':
            name = args[args.index('--pattern') + 1]
            out = pathlib.Path(args[args.index('--dir') + 1])
            (out / name).write_bytes(self.assets[name])
        elif action == 'edit':
            self.draft = False
            self.prerelease = '--prerelease=true' in args
            self.latest = '--latest=true' in args
            self.edits += 1
        else:
            raise AssertionError(args)
        return ''

    def test_draft_is_resolved_without_published_tag_endpoint(self):
        self.publisher.main(self.tag)
        self.assertFalse(self.draft)
        self.assertEqual((self.creates, self.edits), (1, 1))

    def test_unknown_creation_reuses_original_draft(self):
        self.interrupt_create = True
        with self.assertRaisesRegex(RuntimeError, 'unknown create'):
            self.publisher.main(self.tag)
        self.assertTrue(self.draft)
        self.publisher.main(self.tag)
        self.assertEqual(self.creates, 1)
        self.assertFalse(self.draft)

    def test_unknown_upload_readback_and_published_retry(self):
        self.interrupt_upload = True
        with self.assertRaisesRegex(RuntimeError, 'unknown upload'):
            self.publisher.main(self.tag)
        self.assertTrue(self.draft)
        self.publisher.main(self.tag)
        edits = self.edits
        self.publisher.main(self.tag)
        self.assertEqual(self.edits, edits)
        self.assets['test.zip'] = b'wrong bytes'
        with self.assertRaisesRegex(SystemExit, 'differs'):
            self.publisher.main(self.tag)
        self.assertTrue(all(n == 1 for n in self.uploads.values()))

    def test_lookup_failure_does_not_create_release(self):
        self.lookup_error = 'HTTP 503: upstream unavailable'
        with self.assertRaises(SystemExit):
            self.publisher.main(self.tag)
        self.assertEqual(self.creates, 0)

    def test_stable_is_latest_and_reconciles_without_republishing(self):
        self.publisher.main(self.tag, stable=True)
        self.assertFalse(self.prerelease)
        self.assertTrue(self.latest)
        self.publisher.main(self.tag, stable=True)
        self.assertEqual((self.creates, self.edits), (1, 1))
        with self.assertRaisesRegex(SystemExit, 'different channel'):
            self.publisher.main(self.tag)

    def test_preview_version_cannot_be_stable(self):
        with self.assertRaisesRegex(SystemExit, 'prerelease version'):
            self.publisher.main('tiny-v0.2.0-rc.1', stable=True)
        self.assertEqual(self.creates, 0)


if __name__ == '__main__':
    unittest.main()
