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
s=(root/'internal/gateway/gateway.go').read_text(encoding='utf-8')
for name in ('KeepOnAnAccountWithRoom','KeepCodexWindowsWarm','KeepClaudeWindowsWarm','KeepResetsFromRunningOut','KeepPluginsCheckedIn','awake.Keep'):
    assert name not in s, f'removed gateway background job resurrected: {name}'
for path in ('POST /v1/images/generations','POST /v1/videos','POST /v1/audio/transcriptions'):
    assert f'mux.HandleFunc("{path}"' not in s,path
for workflow in (root/'.github/workflows').glob('*'):
    assert 'yetone/magpie-releases' not in workflow.read_text(encoding='utf-8'),f'upstream release target in {workflow}'
assert 'go install ' not in (root/'Makefile').read_text(encoding='utf-8'), 'upstream-named installer resurrected'
for package in ('backup', 'davsync', 'library', 'profile', 'awake', 'imagemcp', 'otlp', 'tune'):
    assert not list((root/'internal'/package).glob('*_test.go')), f'removed product tests resurrected: {package}'
assert {p.name for p in (root/'internal/agent').glob('*_test.go')} <= {'tiny_scope_test.go', 'disconnect_stays_test.go', 'sync_keep_test.go', 'connect_magpie_empty_test.go', 'applied_keep_test.go', 'main_test.go', 'codex_connected_test.go', 'connect_group_test.go', 'codex_route_test.go', 'disconnect_test.go', 'applied_test.go', 'claude_effort_test.go', 'codex_effort_test.go', 'sync_test.go', 'connect_all_hidden_test.go', 'stash_test.go', 'agentenv_test.go', 'alma_test.go', 'astrbot_test.go', 'connect_pick_test.go', 'aionui_test.go', 'codex_catalog_gone_test.go', 'codex_signin_unknown_test.go', 'dsh_max_effort_test.go', 'dsh_official_test.go', 'kirocrew_test.go', 'newmax_test.go', 'penguin_test.go', 'wsl_agy_test.go', 'wsl_nat_test.go', 'zoocode_test.go'}, 'out-of-scope Agent tests resurrected'
assert {p.name for p in (root/'internal/update').glob('*_test.go')} <= {'main_test.go', 'cli_path_test.go', 'cli_path_first_test.go', 'cli_shim_test.go', 'relink_test.go', 'replaced_test.go', 'aside_test.go', 'update_test.go', 'tiny_update_test.go', 'tiny_download_idle_test.go', 'relaunch_fallback_test.go', 'relaunch_retry_test.go', 'download_test.go'}, 'upstream app updater tests resurrected'
session_tests = re.findall(r'^func (Test\w+)\(', '\n'.join(p.read_text(encoding='utf-8') for p in (root/'internal/sessions').glob('*_test.go')), re.M)
for name in session_tests:
    assert not name.startswith(('TestDeleteRestore', 'TestDeleteActive', 'TestDeleteOnlyKnown', 'TestPurge', 'TestMarkdown', 'TestClaudeRelocation', 'TestSearchSession', 'TestSearchOpenCode', 'TestSearchCursor', 'TestWindowsSession', 'TestDshDelete', 'TestGatewayDelete', 'TestGatewayPurge', 'TestGatewayRestore')), f'removed session management test resurrected: {name}'
for path in (root/'internal/gui').glob('*_test.go'):
    if path.name in ('sessions_stats_cap_test.go', 'agent_models_codex_group_test.go'):
        continue  # shared usage statistics cache, not the removed session page
    assert not path.name.startswith(('agent_model', 'archive', 'backup', 'davsync_', 'library_', 'sessions_')), f'removed GUI test resurrected: {path.name}'
removed_browser_tests = {
    'codex-memories.test.cjs',
    'codex-memories-settings.test.cjs',
    'codex-subagent-model.test.cjs',
    'codex-subagents-lead.test.cjs',
    'subagent-effort.test.cjs',
    'profile-confirm.test.cjs',
    'profile-details.test.cjs',
    'profile-in-use.test.cjs',
    'panel-profiles.test.cjs',
    'skill-install-agents-all.test.cjs',
    'skill-install-skip.test.cjs',
    'font-cascade.test.cjs',
    'provider-unmasked.test.cjs',
    'warm-at-own.test.cjs',
    'warm-at-times.test.cjs',
    'checkin-account.test.cjs',
    'usage-checkin.test.cjs',
    'vision-mcp.test.cjs',
    'image-timeout-setting.test.cjs',
    'images-library-way.test.cjs',
    'privacy-model.test.cjs',
}
for path in (root/'internal/gui/tests').glob('*.test.cjs'):
    assert path.name not in removed_browser_tests, f'removed product browser test resurrected: {path.name}'
    if path.name in ('agent-models-group-on.test.cjs', 'agent-models-own-served.test.cjs'):
        continue  # retained manual model visibility; no Agent management
    assert not path.name.startswith(('agents-', 'agent-connect', 'agent-disconnect', 'agent-model', 'library', 'sessions', 'session-', 'backup', 'archive', 'sync-', 'intent', 'systemone', 'warmup', 'searcher-', 'redact-')), f'removed page test resurrected: {path.name}'
print('PASS: tiny assembly/lifecycle scope')
