#!/usr/bin/env python3
import copy
import importlib.util
import pathlib
import unittest

spec = importlib.util.spec_from_file_location('resume', pathlib.Path(__file__).with_name('verify-release-resume.py'))
resume = importlib.util.module_from_spec(spec)
spec.loader.exec_module(resume)


class Resume(unittest.TestCase):
    def setUp(self):
        self.run = dict(id=42, repository={'full_name': resume.REPO}, head_repository={'full_name': resume.REPO},
                        status='completed', event='push', path='.github/workflows/release.yml',
                        head_branch='tiny-v0.1.2', head_sha='a'*40)
        self.jobs = {'jobs': [dict(name=name, conclusion='success') for name in resume.REQUIRED]}

    def verify(self, run=None, jobs=None):
        resume.verify(self.run if run is None else run, self.jobs if jobs is None else jobs,
                      '42', 'tiny-v0.1.2', 'a'*40)

    def test_accepted_original_run(self):
        self.verify()

    def test_different_run_source_tag_or_workflow_is_rejected(self):
        for key, value in (('id', 43), ('head_sha', 'b'*40), ('head_branch', 'main'),
                           ('event', 'pull_request'), ('path', '.github/workflows/test.yml'),
                           ('status', 'in_progress'), ('repository', {'full_name': 'other/repo'}),
                           ('head_repository', {'full_name': 'fork/repo'})):
            with self.subTest(key=key), self.assertRaises(SystemExit):
                self.verify(dict(self.run, **{key: value}))

    def test_any_missing_or_failed_gate_blocks_recovery(self):
        for i in range(len(self.jobs['jobs'])):
            jobs = copy.deepcopy(self.jobs)
            jobs['jobs'][i]['conclusion'] = 'failure'
            with self.assertRaises(SystemExit):
                self.verify(jobs=jobs)


if __name__ == '__main__':
    unittest.main()
