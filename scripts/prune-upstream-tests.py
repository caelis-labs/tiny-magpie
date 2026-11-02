#!/usr/bin/env python3
"""Keep only Tiny boundary tests; upstream owns its unchanged test suites."""
import argparse
from pathlib import Path
import subprocess


def inherited_tests(root):
    paths = subprocess.check_output(
        ['git', 'ls-files', '--cached', '--others', '--exclude-standard', '-z'],
        cwd=root, text=True).split('\0')
    result = []
    for name in sorted(set(paths)):
        path = root / name
        if not path.is_file() or name.startswith(('.tiny/', 'node_modules/', 'output/')):
            continue
        if ((name.endswith('_test.go') and not path.name.startswith('tiny_')) or
                (name.endswith(('.test.cjs', '.test.mjs', '.test.js')) and not path.name.startswith('tiny-'))):
            result.append(path)
    return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--apply', action='store_true', help='delete inherited test files')
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[1]
    paths = inherited_tests(root)
    for path in paths:
        if args.apply:
            path.unlink()
        else:
            print(path.relative_to(root))
    print(f'{"Removed" if args.apply else "Found"} {len(paths)} upstream test files')
