#!/usr/bin/env python3
"""Keep the complete Go log and stream bounded lines for live CI diagnosis."""
import json
import pathlib
import sys

path = pathlib.Path('output/go-test.log')
path.parent.mkdir(exist_ok=True)
pending = {}


def console(prefix, output):
    # test2json splits an oversized logical line across Output events.
    # Bound that whole line, rather than printing every small fragment.
    buffered, truncated = pending.get(prefix, ('', False))
    while output:
        fragment, newline, output = output.partition('\n')
        if not truncated:
            buffered += fragment[:2001 - len(buffered)]
            if len(buffered) > 2000:
                print(prefix + ': ' + buffered[:2000] + ' [long line; see complete log artifact]', flush=True)
                buffered, truncated = '', True
        if newline:
            if not truncated:
                print(prefix + ': ' + buffered.rstrip('\r'), flush=True)
            buffered, truncated = '', False
    pending[prefix] = buffered, truncated


with path.open('w', encoding='utf-8') as log, path.with_suffix('.jsonl').open('w', encoding='utf-8') as raw:
    for line in sys.stdin:
        raw.write(line)
        raw.flush()
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            event = {'Output': line}
        if not isinstance(event, dict) or 'Action' not in event:
            event = {'Output': line}
        output = event.get('Output', '')
        if not output:
            continue
        log.write(output)
        log.flush()
        console(event.get('Package', event.get('ImportPath', 'go')), output)
for prefix, (buffered, _) in pending.items():
    if buffered:
        print(prefix + ': ' + buffered, flush=True)
