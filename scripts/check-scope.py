#!/usr/bin/env python3
"""Small assembly/lifecycle tripwire for future upstream merges; not a protocol test."""
import json
import pathlib
import re
root=pathlib.Path(__file__).resolve().parents[1]
spec=json.loads((root/'UPSTREAM.json').read_text(encoding='utf-8'))
actual=set(re.findall(r'magpie/plugins/([^"/]+)',(root/'plugins/all/all.go').read_text(encoding='utf-8')))
assert actual==set(spec['retainedBuiltins']), ('built-in assembly changed',actual)
for name in ('tui','library','sessions','search','warmups','otel','tools','redact'):
    assert not (root/'plugins'/name).exists(),f'removed built-in resurrected: {name}'
assert not (root/'internal/tui').exists(),'TUI resurrected'
assert not (root/'site').exists(),'upstream marketing/update Worker resurrected'
s=(root/'internal/gateway/gateway.go').read_text(encoding='utf-8')
for name in ('KeepOnAnAccountWithRoom','KeepCodexWindowsWarm','KeepClaudeWindowsWarm','KeepResetsFromRunningOut','KeepPluginsCheckedIn','awake.Keep'):
    assert name not in s, f'removed gateway background job resurrected: {name}'
for path in ('POST /v1/images/generations','POST /v1/videos','POST /v1/audio/transcriptions'):
    assert f'mux.HandleFunc("{path}"' not in s,path
for workflow in (root/'.github/workflows').glob('*'):
    assert 'yetone/magpie-releases' not in workflow.read_text(encoding='utf-8'),f'upstream release target in {workflow}'
assert 'go install ' not in (root/'Makefile').read_text(encoding='utf-8'), 'upstream-named installer resurrected'
# One convention replaces per-product/per-test retirement inventories.
import runpy
policy = runpy.run_path(str(root/'scripts/prune-upstream-tests.py'))
assert not policy['inherited_tests'](root), 'run scripts/prune-upstream-tests.py --apply'
for path in ('GET /me', 'GET /v1/magpie/me'):
    assert f'mux.HandleFunc("{path}"' not in s, f'rejected key-holder page resurrected: {path}'
print('PASS: tiny assembly/lifecycle scope')
