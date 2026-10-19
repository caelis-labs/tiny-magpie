#!/usr/bin/env python3
"""Compare built Magpie/tiny binaries on an isolated 4x4 HTTP/SSE protocol matrix.
No credentials, external servers or running gateways are used.
"""
import argparse
import contextlib
import http.server
import json
import os
import pathlib
import socket
import subprocess
import tempfile
import threading
import time
import urllib.error
import urllib.request

PROTOCOLS = ('chat', 'responses', 'anthropic', 'gemini')

class Vendor(http.server.BaseHTTPRequestHandler):
    def log_message(self, *args): pass
    def do_GET(self):
        self.send_response(200); self.send_header('Content-Type', 'application/json'); self.end_headers()
        self.wfile.write(b'{"object":"list","data":[{"id":"fixture-model"}]}')
    def do_POST(self):
        body = json.loads(self.rfile.read(int(self.headers.get('Content-Length', 0))))
        self.server.received.append((self.path, body))
        if self.path.endswith('/systemone') or self.path.endswith('/embeddings') or self.path.endswith('/rerank'):
            self.send_response(200); self.send_header('Content-Type','application/json'); self.end_headers()
            if self.path.endswith('/systemone'):
                reply={'model':body['model'],'answers':{'kind':{'type':'choice','choice':'code','confidence':0.9}}}
            elif self.path.endswith('/embeddings'):
                reply={'object':'list','data':[{'object':'embedding','index':0,'embedding':[0.1,-0.2]}],'usage':{'prompt_tokens':5,'total_tokens':5}}
            else:
                reply={'results':[{'index':0,'relevance_score':0.9}],'usage':{'total_tokens':5}}
            self.wfile.write(json.dumps(reply).encode()); return
        if self.path.endswith('/chat/completions'):
            events = [('', {'id':'fixture-chat','object':'chat.completion.chunk','created':1,'model':'fixture-model','choices':[{'index':0,'delta':{'role':'assistant','content':'fixture hello'}}]}),
                      ('', {'id':'fixture-chat','choices':[{'index':0,'delta':{},'finish_reason':'stop'}],'usage':{'prompt_tokens':7,'completion_tokens':3,'total_tokens':10}})]
        elif self.path.endswith('/responses'):
            response = {'id':'fixture-response','object':'response','model':'fixture-model','status':'completed','output':[{'id':'fixture-message','type':'message','role':'assistant','status':'completed','content':[{'type':'output_text','text':'fixture hello','annotations':[]}]}],'usage':{'input_tokens':7,'output_tokens':3,'total_tokens':10}}
            events = [('response.created',{'type':'response.created','response':dict(response,status='in_progress',output=[])}),
                      ('response.output_text.delta',{'type':'response.output_text.delta','item_id':'fixture-message','output_index':0,'content_index':0,'delta':'fixture hello'}),
                      ('response.completed',{'type':'response.completed','response':response})]
        elif self.path.endswith('/messages'):
            events = [('message_start',{'type':'message_start','message':{'id':'fixture-message','type':'message','role':'assistant','model':'fixture-model','content':[],'usage':{'input_tokens':7,'output_tokens':0}}}),
                      ('content_block_start',{'type':'content_block_start','index':0,'content_block':{'type':'text','text':''}}),
                      ('content_block_delta',{'type':'content_block_delta','index':0,'delta':{'type':'text_delta','text':'fixture hello'}}),
                      ('content_block_stop',{'type':'content_block_stop','index':0}),
                      ('message_delta',{'type':'message_delta','delta':{'stop_reason':'end_turn'},'usage':{'output_tokens':3}}),('message_stop',{'type':'message_stop'})]
        elif ':streamGenerateContent' in self.path or ':generateContent' in self.path:
            events = [('',{'candidates':[{'index':0,'content':{'role':'model','parts':[{'text':'fixture hello'}]},'finishReason':'STOP'}],'modelVersion':'fixture-model','usageMetadata':{'promptTokenCount':7,'candidatesTokenCount':3,'totalTokenCount':10}})]
        else:
            self.send_error(404); return
        self.send_response(200); self.send_header('Content-Type','text/event-stream'); self.end_headers()
        for name, data in events:
            if name: self.wfile.write(('event: '+name+'\n').encode())
            self.wfile.write(('data: '+json.dumps(data)+'\n\n').encode())
        if self.path.endswith('/chat/completions'): self.wfile.write(b'data: [DONE]\n\n')

@contextlib.contextmanager
def vendor():
    server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Vendor)
    server.received=[]
    thread=threading.Thread(target=server.serve_forever,daemon=True); thread.start()
    try: yield server, f'http://127.0.0.1:{server.server_port}'
    finally: server.shutdown(); server.server_close(); thread.join()

def free_port():
    with socket.socket() as sock:
        sock.bind(('127.0.0.1',0)); return sock.getsockname()[1]

def request(base, path, body=None):
    data=None if body is None else json.dumps(body).encode()
    req=urllib.request.Request(base+path,data,{'Content-Type':'application/json','Authorization':'Bearer magpie'})
    opener=urllib.request.build_opener(urllib.request.ProxyHandler({}))
    try:
        with opener.open(req,timeout=20) as r: return r.status, r.read().decode()
    except urllib.error.HTTPError as e: return e.code,e.read().decode()

@contextlib.contextmanager
def gateway(binary, name, upstream):
    with tempfile.TemporaryDirectory(prefix='tiny-magpie-verify-') as root:
        root=pathlib.Path(root); cfg=root/'.config'/name; cfg.mkdir(parents=True)
        cache=root/'.cache'/name; cache.mkdir(parents=True)
        (cache/'models.json').write_text('{}')
        providers=[]
        for protocol in PROTOCOLS:
            address=upstream+('/v1beta' if protocol=='gemini' else '' if protocol=='anthropic' else '/v1')
            providers.append({'id':'fixture-'+protocol,'name':protocol,'key':'fixture-key','models':['fixture-model'],protocol:address})
        providers.append({'id':'fixture-decision','name':'Decision','key':'fixture-key','decide':upstream+'/v1','models':['jev-latest']})
        (cfg/'providers.json').write_text(json.dumps({'providers':providers,'noAutoGroups':True}))
        (cfg/'settings.json').write_text(json.dumps({'noAccountDiscovery':True,'vision':'off','lang':'en'}))
        env={k:v for k,v in os.environ.items() if not k.startswith(('MAGPIE_','CODEX_','CLAUDE_','XDG_')) and 'proxy' not in k.lower()}
        env.update(HOME=str(root), USERPROFILE=str(root),XDG_CONFIG_HOME=str(root/'.config'),XDG_CACHE_HOME=str(root/'.cache'),
                   MAGPIE_ADDR=f'127.0.0.1:{free_port()}',MAGPIE_NO_ACCOUNTS='1',DO_NOT_TRACK='1',
                   HTTP_PROXY='http://127.0.0.1:9',HTTPS_PROXY='http://127.0.0.1:9',NO_PROXY='127.0.0.1,localhost')
        with (root/'gateway.log').open('w+') as log:
            process=subprocess.Popen([str(pathlib.Path(binary).resolve()),'serve'],env=env,stdout=log,stderr=log)
            base='http://'+env['MAGPIE_ADDR']
            try:
                for _ in range(150):
                    if process.poll() is not None:
                        log.seek(0); raise RuntimeError(log.read())
                    try:
                        if request(base,'/')[0]==200: break
                    except OSError: pass
                    time.sleep(.1)
                else: raise RuntimeError('gateway did not start')
                yield base
                assert not (root/'.codex').exists(), 'gateway unexpectedly wrote Codex configuration'
                assert not (root/'.claude').exists(), 'gateway unexpectedly wrote Claude configuration'
            finally:
                process.terminate()
                try: process.wait(timeout=10)
                except subprocess.TimeoutExpired: process.kill(); process.wait()

def call(protocol, upstream, stream):
    model='fixture-'+upstream+'/fixture-model'
    if protocol=='chat': return '/v1/chat/completions',{'model':model,'stream':stream,'messages':[{'role':'user','content':'hello'}]}
    if protocol=='responses': return '/v1/responses',{'model':model,'stream':stream,'store':False,'input':[{'role':'user','content':'hello'}]}
    if protocol=='anthropic': return '/v1/messages',{'model':model,'stream':stream,'max_tokens':64,'messages':[{'role':'user','content':'hello'}]}
    action='streamGenerateContent?alt=sse' if stream else 'generateContent'
    return '/v1beta/models/'+model+':'+action, {'contents':[{'role':'user','parts':[{'text':'hello'}]}]}

def tool_call(protocol, upstream, stream):
    path, body = call(protocol, upstream, stream)
    image = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII='
    schema = {'type':'object','properties':{'path':{'type':'string'}},'required':['path']}
    if protocol == 'chat':
        body.update(messages=[{'role':'user','content':'read the file'},
            {'role':'assistant','tool_calls':[{'id':'call_fixture','type':'function','function':{'name':'read','arguments':'{"path":"fixture.txt"}'}}]},
            {'role':'tool','tool_call_id':'call_fixture','content':'fixture file body'},
            {'role':'user','content':[{'type':'text','text':'compare with this screenshot'},{'type':'image_url','image_url':{'url':'data:image/png;base64,'+image}}]}],
            tools=[{'type':'function','function':{'name':'read','description':'Read a file','parameters':schema}}], reasoning_effort='medium')
    elif protocol == 'responses':
        body.update(instructions='Inspect the file and screenshot.', input=[{'role':'user','content':'read the file'},
            {'type':'function_call','call_id':'call_fixture','name':'read','arguments':'{"path":"fixture.txt"}'},
            {'type':'function_call_output','call_id':'call_fixture','output':'fixture file body'},
            {'role':'user','content':[{'type':'input_text','text':'compare with this screenshot'},{'type':'input_image','image_url':'data:image/png;base64,'+image}]}],
            tools=[{'type':'function','name':'read','description':'Read a file','parameters':schema}],
            reasoning={'effort':'medium','summary':'auto'}, include=['reasoning.encrypted_content'], max_output_tokens=128, service_tier='auto')
    elif protocol == 'anthropic':
        body.update(messages=[{'role':'user','content':'read the file'},
            {'role':'assistant','content':[{'type':'tool_use','id':'call_fixture','name':'read','input':{'path':'fixture.txt'}}]},
            {'role':'user','content':[{'type':'tool_result','tool_use_id':'call_fixture','content':'fixture file body'},
                {'type':'text','text':'compare with this screenshot'},{'type':'image','source':{'type':'base64','media_type':'image/png','data':image}}]}],
            tools=[{'name':'read','description':'Read a file','input_schema':schema}])
    else:
        body.update(contents=[{'role':'user','parts':[{'text':'read the file'}]},
            {'role':'model','parts':[{'functionCall':{'id':'call_fixture','name':'read','args':{'path':'fixture.txt'}}}]},
            {'role':'user','parts':[{'functionResponse':{'id':'call_fixture','name':'read','response':{'output':'fixture file body'}}},
                {'text':'compare with this screenshot'},{'inlineData':{'mimeType':'image/png','data':image}}]}],
            tools=[{'functionDeclarations':[{'name':'read','description':'Read a file','parameters':schema}]}])
    return path, body

def normalize(value):
    if isinstance(value,dict): return {k:normalize(v) for k,v in value.items() if k not in ('id','item_id','created','created_at','timestamp','responseId')}
    if isinstance(value,list): return [normalize(v) for v in value]
    return value

def semantic(status,body):
    if body.lstrip().startswith(('{','[')): return status,normalize(json.loads(body))
    return status,[normalize(json.loads(line[6:])) if line.startswith('data: ') and line[6:]!='[DONE]' else line for line in body.splitlines() if line]

def main():
    parser=argparse.ArgumentParser(); parser.add_argument('--baseline',required=True); parser.add_argument('--tiny',required=True); args=parser.parse_args()
    checked=0
    with vendor() as (fixture, upstream), gateway(args.baseline,'magpie',upstream) as original, gateway(args.tiny,'tiny-magpie',upstream) as tiny:
        a=request(original,'/v1/models'); b=request(tiny,'/v1/models')
        assert a==b,'model discovery differs'
        for corpus in (call, tool_call):
            for source in PROTOCOLS:
                for target in PROTOCOLS:
                    for stream in (False,True):
                        path,body=corpus(source,target,stream)
                        fixture.received.clear()
                        a=request(original,path,body)
                        sent_a=list(fixture.received)
                        fixture.received.clear()
                        b=request(tiny,path,body)
                        sent_b=list(fixture.received)
                        label=(corpus.__name__,source,target,stream)
                        assert a[0]==200 and b[0]==200,(label,a,b)
                        assert sent_a and sent_a==sent_b,(label,'upstream request differs',sent_a,sent_b)
                        assert 'fixture hello' in a[1] and 'fixture hello' in b[1],(label,'empty reply')
                        assert semantic(*a)==semantic(*b),(label,'response differs',a,b)
                        checked+=1
        for path,body in [('/v1/systemone',{'model':'fixture-decision/jev-latest','state':{'text':'hello'},'questions':{'kind':{'type':'choice','criteria':{'code':None,'other':None}}}}),
                          ('/v1/embeddings',{'model':'fixture-chat/fixture-model','input':['hello'],'dimensions':2}),
                          ('/v1/rerank',{'model':'fixture-chat/fixture-model','query':'hello','documents':['hello world']})]:
            fixture.received.clear(); a=request(original,path,body); sent_a=list(fixture.received)
            fixture.received.clear(); b=request(tiny,path,body); sent_b=list(fixture.received)
            assert a[0]==b[0]==200 and semantic(*a)==semantic(*b),(path,a,b)
            assert sent_a and sent_a==sent_b,(path,sent_a,sent_b)
            checked+=1
        for path in ('/v1/images/generations' ,'/v1/audio/transcriptions','/v1/videos','/mcp/magpie/tools','/alpha/search'):
            assert request(tiny,path,{})[0]==404,path
    print(f'PASS: discovery + {checked} bidirectional protocol/stream comparisons (text + tools/screenshots) + removed endpoint and config-write checks')
if __name__=='__main__': main()
