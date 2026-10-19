#!/usr/bin/env python3
"""Small assembly/lifecycle tripwire for future upstream merges; not a protocol test."""
import json
import pathlib
import re
root=pathlib.Path(__file__).resolve().parents[1]
spec=json.loads((root/'UPSTREAM.json').read_text(encoding='utf-8'))
actual=set(re.findall(r'magpie/plugins/([^"/]+)',(root/'plugins/all/all.go').read_text(encoding='utf-8')))
assert actual==set(spec['retainedBuiltins']), ('built-in assembly changed',actual)
for name in ('tui','library','sessions','search','warmups','otel','tools'):
    assert not (root/'plugins'/name).exists(),f'removed built-in resurrected: {name}'
assert not (root/'internal/tui').exists(),'TUI resurrected'
s=(root/'internal/gateway/gateway.go').read_text(encoding='utf-8')
for name in ('KeepOnAnAccountWithRoom','KeepCodexWindowsWarm','KeepClaudeWindowsWarm','KeepResetsFromRunningOut','KeepPluginsCheckedIn','awake.Keep'):
    assert name not in s, f'removed gateway background job resurrected: {name}'
for path in ('POST /v1/systemone','POST /v1/images/generations','POST /v1/videos','POST /v1/audio/transcriptions'):
    assert f'mux.HandleFunc("{path}"' not in s,path
for workflow in (root/'.github/workflows').glob('*'):
    assert 'yetone/magpie-releases' not in workflow.read_text(encoding='utf-8'),f'upstream release target in {workflow}'
assert 'go install ' not in (root/'Makefile').read_text(encoding='utf-8'), 'upstream-named installer resurrected'
for package in ('backup', 'davsync', 'library', 'profile', 'awake', 'imagemcp', 'otlp', 'tune'):
    assert not list((root/'internal'/package).glob('*_test.go')), f'removed product tests resurrected: {package}'
assert {p.name for p in (root/'internal/agent').glob('*_test.go')} <= {'main_test.go', 'agentenv_test.go'}, 'Agent management tests resurrected'
for path in (root/'internal/gui').glob('*_test.go'):
    if path.name == 'sessions_stats_cap_test.go':
        continue  # shared usage statistics cache, not the removed session page
    assert not path.name.startswith(('agent_model', 'archive', 'backup', 'davsync_', 'library_', 'sessions_')), f'removed GUI test resurrected: {path.name}'
for path in (root/'internal/gui/tests').glob('*.test.cjs'):
    assert not path.name.startswith(('agents-', 'agent-connect', 'agent-disconnect', 'agent-model', 'library', 'sessions', 'session-', 'backup', 'archive', 'sync-', 'intent', 'systemone', 'warmup', 'searcher-')), f'removed page test resurrected: {path.name}'
print('PASS: tiny assembly/lifecycle scope')
