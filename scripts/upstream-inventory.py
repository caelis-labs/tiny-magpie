#!/usr/bin/env python3
"""Keep one actionable upstream issue; unchanged upstream produces no notification."""
import json
import os
import pathlib
import subprocess
import tempfile

ROOT = pathlib.Path(__file__).resolve().parent.parent
REPO = 'caelis-labs/tiny-magpie'


def run(*args):
    return subprocess.check_output(args, cwd=ROOT, text=True).strip()


def main():
    base = json.loads((ROOT / 'UPSTREAM.json').read_text(encoding='utf-8'))['base']
    head = run('git', 'rev-parse', 'upstream/main')
    subprocess.run(['git', 'merge-base', '--is-ancestor', base, head], cwd=ROOT, check=True)
    interval = f'{base}..{head}'
    commits = run('git', 'log', '--oneline', interval)
    files = run('git', 'diff', '--stat', interval)
    output = ROOT / 'output'
    output.mkdir(exist_ok=True)
    for name, content in [('commits', commits), ('files', files)]:
        (output / f'upstream-{name}.txt').write_text(content + '\n', encoding='utf-8')
    body = (f'Integrated: `{base}`\n\nUpstream main: `{head}`\n\n'
            f'[Compare upstream changes](https://github.com/yetone/magpie/compare/{base}...{head})\n\n'
            'Follow [UPSTREAM.md](https://github.com/caelis-labs/tiny-magpie/blob/main/UPSTREAM.md). '
            'Retain model and plugin protocols and all Provider families. Review removed APIs, '
            'background services, UI pages and update channels before merging. '
            'Run scope, full platform CI, binary differential and GUI checks. '
            'Update UPSTREAM.json only after integration; this inventory does not merge code.\n\n'
            f'```text\n{commits[:18000] or "No new commits."}\n```\n\n'
            f'```text\n{files[:18000]}\n```\n\n<!-- tiny-upstream:{head} -->\n')
    if summary := os.environ.get('GITHUB_STEP_SUMMARY'):
        pathlib.Path(summary).write_text(body, encoding='utf-8')
    issues = json.loads(run('gh', 'issue', 'list', '-R', REPO, '--state', 'open',
                            '--limit', '1000', '--json', 'number,title,body'))
    issues = [issue for issue in issues if issue['title'] == 'Upstream sync']
    if len(issues) > 1:
        raise SystemExit('Multiple upstream issues require reconciliation')
    if base == head:
        if issues:
            run('gh', 'issue', 'close', str(issues[0]['number']), '-R', REPO, '--reason', 'completed')
        print('Integrated upstream is current')
        return
    if issues and f'<!-- tiny-upstream:{head} -->' in issues[0]['body'] and f'`{base}`' in issues[0]['body']:
        print('Upstream unchanged since last inventory; issue left untouched')
        return
    with tempfile.TemporaryDirectory(prefix='tiny-inventory-') as temp:
        path = pathlib.Path(temp) / 'body.md'
        path.write_text(body, encoding='utf-8')
        if issues:
            print(run('gh', 'issue', 'edit', str(issues[0]['number']), '-R', REPO, '--body-file', str(path)))
        else:
            print(run('gh', 'issue', 'create', '-R', REPO, '--title', 'Upstream sync', '--body-file', str(path)))


if __name__ == '__main__':
    main()
