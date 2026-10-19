#!/usr/bin/env python3
"""Accept only a completed, same-repository tag run with all original gates green."""
import json
import sys

REPO = 'caelis-labs/tiny-magpie'
REQUIRED = {
    'source', 'checks / protocols-and-ui',
    'checks / go (ubuntu-latest)', 'checks / go (macos-latest)', 'checks / go (windows-latest)',
    'containers / build (ubuntu-24.04)', 'containers / build (ubuntu-24.04-arm)',
    'package (ubuntu-24.04, linux-amd64)', 'package (ubuntu-24.04-arm, linux-arm64)',
    'package (windows-latest, windows-amd64)', 'package (windows-11-arm, windows-arm64)',
}


def verify(run, jobs, run_id, tag, sha):
    if (str(run.get('id')) != run_id or run.get('repository', {}).get('full_name') != REPO
            or run.get('head_repository', {}).get('full_name') != REPO
            or run.get('status') != 'completed' or run.get('event') != 'push'
            or run.get('path') != '.github/workflows/release.yml'
            or run.get('head_branch') != tag or run.get('head_sha') != sha):
        raise SystemExit('Recovery run does not match the trusted tagged source')
    passed = {job['name'] for job in jobs['jobs'] if job.get('conclusion') == 'success'}
    if not REQUIRED <= passed:
        raise SystemExit(f'Recovery requires successful original gates/packages: {REQUIRED-passed}')


if __name__ == '__main__':
    run_file, jobs_file, run_id, tag, sha = sys.argv[1:]
    verify(json.load(open(run_file)), json.load(open(jobs_file)), run_id, tag, sha)
    print('Trusted original gates and four platform packages verified')
