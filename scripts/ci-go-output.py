#!/usr/bin/env python3
"""Keep the complete Go log and stream bounded lines for live CI diagnosis."""
import json
import pathlib
import sys

path = pathlib.Path('output/go-test.log')
path.parent.mkdir(exist_ok=True)
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
        for text in output.splitlines():
            prefix = event.get('Package', event.get('ImportPath', 'go'))
            print(prefix + ': ' + text[:2000] + (' [long line; see complete log artifact]' if len(text) > 2000 else ''), flush=True)
