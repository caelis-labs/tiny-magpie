#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
if [[ $# != 1 || ! "$1" =~ ^[0-9a-f]{40}$ ]]; then echo "usage: $0 <full-upstream-sha>" >&2; exit 2; fi
if [[ -n "$(git status --porcelain)" ]]; then echo 'Commit or stash changes before integrating upstream.' >&2; exit 1; fi
git fetch upstream main
git merge-base --is-ancestor "$1" upstream/main
git merge --no-commit --no-ff "$1"
python3 scripts/check-scope.py
printf '%s\n' 'Review the merge against docs/tiny-scope.md; run compatibility checks before updating UPSTREAM.json and committing.'
