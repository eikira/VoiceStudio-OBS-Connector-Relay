#!/bin/bash
PLIST="$HOME/Library/LaunchAgents/com.voicestudio.obsrelay.plist"
launchctl unload "$PLIST" 2>/dev/null; rm -f "$PLIST"
echo "Done. The relay no longer starts at login."
