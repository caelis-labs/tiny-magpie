#!/usr/bin/env bash
# Build-time only, isolated from the user's Python installation.
set -euo pipefail
tiny_dmg_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
tiny_dmg_env="$tiny_dmg_root/.tiny/dmg-tools"
if [ ! -x "$tiny_dmg_env/bin/python" ]; then python3 -m venv "$tiny_dmg_env"; fi
if ! "$tiny_dmg_env/bin/python" -c 'import importlib.metadata as m; assert all(m.version(k)==v for k,v in {"dmgbuild":"1.6.7","ds-store":"1.3.3","mac-alias":"2.2.3"}.items())' 2>/dev/null; then
  "$tiny_dmg_env/bin/python" -m pip install --disable-pip-version-check -r "$tiny_dmg_root/scripts/dmg-requirements.txt"
fi
