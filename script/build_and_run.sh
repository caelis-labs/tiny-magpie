#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
root="$PWD"
profile="$root/.tiny/native-profile"
mkdir -p "$profile"
pidfile="$root/.tiny/native.pid"
program="$root/tiny-magpie.app/Contents/MacOS/tiny-magpie"
# Only stop the exact process previously launched by this script.
if [[ -f "$pidfile" ]]; then
  read -r pid < "$pidfile"
  if [[ "$pid" =~ ^[0-9]+$ ]] && [[ "$(ps -p "$pid" -o comm=)" == "$program" ]]; then
    kill "$pid"
    for _ in {1..20}; do kill -0 "$pid" 2>/dev/null || break; sleep 0.1; done
    if kill -0 "$pid" 2>/dev/null; then echo "Previous test app is still stopping." >&2; exit 1; fi
  fi
  rm -f "$pidfile"
fi
GOWORK=off make app VERSION=0.1.0-dev
export XDG_CONFIG_HOME="$profile/config" XDG_CACHE_HOME="$profile/cache"
export MAGPIE_ADDR=127.0.0.1:13425 MAGPIE_NO_ACCOUNTS=1 DO_NOT_TRACK=1
# The app inherits only this launch's sandbox roots; launchd is untouched.
case "${1:-run}" in
  --debug) exec lldb -- "$root/tiny-magpie.app/Contents/MacOS/tiny-magpie" app ;;
  run|--verify|--logs|--telemetry)
    "$program" app > "$root/.tiny/native-app.log" 2>&1 &
    pid=$!
    printf '%s\n' "$pid" > "$pidfile"
    if [[ "${1:-}" == --verify ]]; then sleep 2; kill -0 "$pid"; fi
    if [[ "${1:-}" == --logs || "${1:-}" == --telemetry ]]; then /usr/bin/log stream --info --style compact --predicate 'process == "tiny-magpie"'; fi ;;
  *) echo "usage: $0 [--debug|--verify|--logs|--telemetry]" >&2; exit 2 ;;
esac
