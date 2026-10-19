"""Finder presentation; never change metadata inside the signed app."""
import os

application = os.path.abspath("tiny-magpie.app")
format = "UDZO"
filesystem = "HFS+"
files = [application, (defines["docs"], ".documentation")]
symlinks = {"Applications": "/Applications"}
icon = os.path.join(application, "Contents", "Resources", "magpie.icns")
background = defines["background"]
window_rect = ((240, 180), (660, 360))
default_view = "icon-view"
show_toolbar = show_status_bar = show_tab_view = show_pathbar = show_sidebar = False
include_icon_view_settings = True
include_list_view_settings = False
arrange_by = None
icon_size = 112
text_size = 14
icon_locations = {"tiny-magpie.app": (170, 150), "Applications": (490, 150)}
