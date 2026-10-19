#!/usr/bin/env python3
"""Retry only explicit transient Apple ticket transport failures, never trust failures."""
import re
import subprocess
import sys
import time


def transient(status, output):
    if status != 68 or re.search(r'invalid|revok|reject|mismatch|signature|certificate|trust|authentic|unauthor|unsigned|not (?:code-)?signed|not (?:previously )?notarized|\b(?:TLS|SSL)\b|\bHTTP(?:/[\d.]+)?\s+[45]\d\d\b', output, re.I):
        return False
    codes = re.findall(r'\bDomain=NSURLErrorDomain\s+Code=(-?\d+)\b', output)
    return bool(codes) and set(codes) <= {'-1001', '-1004', '-1005'}


def staple(command, path, execute=subprocess.run, wait=time.sleep):
    if command not in ('staple', 'validate') or not path:
        raise ValueError('Expected staple|validate and one artifact path')
    for attempt in range(3):
        try:
            result = execute(['xcrun', 'stapler', command, path], capture_output=True,
                             text=True, encoding='utf-8', errors='replace', timeout=120)
        except (OSError, subprocess.TimeoutExpired) as error:
            print(f'Stapler did not complete: {error}', file=sys.stderr)
            return 1
        output = result.stdout + result.stderr
        print(output, end='', file=sys.stderr)
        if result.returncode == 0:
            return 0
        if attempt == 2 or not transient(result.returncode, output):
            return result.returncode if result.returncode > 0 else 1
        seconds = (5, 15)[attempt]
        print(f'Temporary Apple ticket connection failure; retry in {seconds}s.', file=sys.stderr)
        wait(seconds)


if __name__ == '__main__':
    if len(sys.argv) != 3:
        raise SystemExit('Expected staple|validate and one artifact path')
    raise SystemExit(staple(*sys.argv[1:]))
