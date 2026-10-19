#!/usr/bin/env python3
import importlib.util
import pathlib
import subprocess
import unittest

spec = importlib.util.spec_from_file_location('tickets', pathlib.Path(__file__).with_name('notarization-stapler.py'))
tickets = importlib.util.module_from_spec(spec)
spec.loader.exec_module(tickets)
LOST = 'Error Domain=NSURLErrorDomain Code=-1005 "The network connection was lost."\nThe validate action failed! Error 68.\n'


class Tickets(unittest.TestCase):
    def run_case(self, outcomes):
        calls, waits = [], []
        def execute(args, **options):
            calls.append((args, options))
            outcome = outcomes[len(calls)-1]
            if isinstance(outcome, Exception):
                raise outcome
            return subprocess.CompletedProcess(args, outcome[0], '', outcome[1])
        result = tickets.staple('validate', 'example.app', execute, waits.append)
        return result, calls, waits

    def test_real_cloudkit_disconnect_then_success(self):
        result, calls, waits = self.run_case([(68, LOST), (0, 'validated\n')])
        self.assertEqual((result, len(calls), waits), (0, 2, [5]))
        self.assertEqual(calls[0][0], ['xcrun', 'stapler', 'validate', 'example.app'])
        self.assertEqual(calls[0][1]['timeout'], 120)

    def test_retry_limit_preserves_failure(self):
        result, calls, waits = self.run_case([(68, LOST)]*3)
        self.assertEqual((result, len(calls), waits), (68, 3, [5, 15]))

    def test_unknown_authenticity_tls_and_mixed_errors_never_retry(self):
        for status, output in ((68, 'Error 68'), (1, LOST), (68, LOST+'not notarized'),
                               (68, LOST+'invalid signature'), (68, LOST+'HTTP 403'),
                               (68, LOST+'Domain=NSURLErrorDomain Code=-1200'), (68, LOST+'TLS failure'),
                               (68, LOST+'signature check failed'), (68, LOST+'certificate trust failure')):
            with self.subTest(output=output):
                result, calls, waits = self.run_case([(status, output)])
                self.assertEqual((result, len(calls), waits), (status, 1, []))

    def test_hung_or_missing_tool_fails_closed(self):
        for error in (OSError('missing tool'), subprocess.TimeoutExpired('xcrun', 120)):
            result, calls, waits = self.run_case([error])
            self.assertEqual((result, len(calls), waits), (1, 1, []))


if __name__ == '__main__':
    unittest.main()
