#!/usr/bin/env bash
# Disposable Linux gateway/Web smoke. Uses only fixed fixture keys and a new volume.
set -euo pipefail
image="${1:-tiny-magpie:local-validation}"
container="tiny-magpie-check-$$-$RANDOM"
volume="$container"
container_id=""
cleanup() {
  result=$?
  [[ -n "$container_id" ]] || return 0
  if [[ "$result" != 0 ]]; then docker logs "$container" || true; fi
  docker rm -f "$container_id" >/dev/null 2>&1 || true
  docker volume rm "$volume" >/dev/null 2>&1 || true
}
trap cleanup EXIT
container_id="$(docker run -d --name "$container" -v "$volume:/config" \
  -p 127.0.0.1::3425 -p 127.0.0.1::3430 \
  -e MAGPIE_NO_ACCOUNTS=1 -e DO_NOT_TRACK=1 -e MAGPIE_PLUGIN_MARKET=off \
  -e HTTP_PROXY=http://127.0.0.1:9 -e HTTPS_PROXY=http://127.0.0.1:9 -e NO_PROXY=localhost,127.0.0.1 \
  -e MAGPIE_WEB_KEY=tiny-fixture-web-key-32-characters \
  "$image" web --addr 0.0.0.0:3430 --no-open)"
ready=0
for _ in $(seq 1 30); do
  if docker exec "$container" tiny-magpie healthcheck >/dev/null 2>&1; then ready=1; break; fi
  sleep 1
done
test "$ready" = 1
docker exec "$container" tiny-magpie gateway-key add Fixture --key tiny-fixture-gateway-secret >/dev/null
export TINY_SMOKE_GATEWAY="http://$(docker port "$container" 3425/tcp)"
export TINY_SMOKE_WEB="http://$(docker port "$container" 3430/tcp)"
python3 - <<'PY'
import http.cookiejar,json,os,urllib.error,urllib.request
base=os.environ['TINY_SMOKE_GATEWAY'];web=os.environ['TINY_SMOKE_WEB']
direct=urllib.request.build_opener(urllib.request.ProxyHandler({}))
def call(path,body=None,auth=True):
    req=urllib.request.Request(base+path,data=None if body is None else json.dumps(body).encode(),headers={'Content-Type':'application/json',**({'Authorization':'Bearer tiny-fixture-gateway-secret'} if auth else {})})
    try:
        with direct.open(req,timeout=10) as r:return r.status,r.read()
    except urllib.error.HTTPError as e:return e.code,e.read()
assert call('/v1/models')[0]==200
assert call('/v1/models',auth=False)[0]==401
assert call('/v1/systemone',{})[0]==400, 'retained System One must validate the request'
assert call('/v1/systemone',{},auth=False)[0]==401, 'System One must require the gateway key'
assert 400<=call('/v1/responses',{'model':'fixture/missing','input':'hi'})[0]<500
cookies=http.cookiejar.CookieJar()
browser=urllib.request.build_opener(urllib.request.ProxyHandler({}),urllib.request.HTTPCookieProcessor(cookies))
with browser.open(web+'/?k=tiny-fixture-web-key-32-characters',timeout=10) as r:assert b'Providers' in r.read()
with browser.open(web+'/api/providers',timeout=10) as r:
    state=json.load(r);assert state['gateway']['mine'] and state['gateway']['running']
try:browser.open(urllib.request.Request(web+'/api/set',data=b'{}',headers={'Content-Type':'application/json'}),timeout=10)
except urllib.error.HTTPError as e:assert e.code==404
else:raise AssertionError('removed management route exists')
print('PASS: Linux container health, named-key auth, discovery, model error, authenticated Web and removed routes')
PY
