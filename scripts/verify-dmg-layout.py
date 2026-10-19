#!/usr/bin/env python3
"""Verify the mounted Finder layout independently of its settings input."""
import pathlib
import sys
from ds_store import DSStore

root = pathlib.Path(sys.argv[1])
assert sorted(p.name for p in root.iterdir() if not p.name.startswith(".")) == ["Applications", "tiny-magpie.app"]
assert (root / "Applications").readlink() == pathlib.Path("/Applications")
assert (root / ".background.tiff").is_file()
with DSStore.open(str(root / ".DS_Store"), "r") as store:
    view = store["."]["icvp"]
    assert view["iconSize"] == 112 and view["textSize"] == 14
    assert view["backgroundType"] == 2 and view["backgroundImageAlias"]
    assert store["tiny-magpie.app"]["Iloc"] == (170, 150)
    assert store["Applications"]["Iloc"] == (490, 150)
    window = store["."]["bwsp"]
    assert "660, 360" in window["WindowBounds"]
    for key in ("ShowToolbar", "ShowStatusBar", "ShowTabView", "ShowPathbar", "ShowSidebar"):
        assert not window[key], key
print("PASS: mounted drag-to-Applications layout, background, icon positions and hidden documentation")
