#!/bin/bash
# Makes the relay start by itself when you log in.
cd "$(dirname "$0")" || exit 1
DIR="$(pwd)"
case "$(uname -m)" in arm64) BIN="$DIR/VoiceStudio-OBS-Connector-Relay-mac-apple-silicon" ;; *) BIN="$DIR/VoiceStudio-OBS-Connector-Relay-mac-intel" ;; esac
PLIST="$HOME/Library/LaunchAgents/com.voicestudio.obsrelay.plist"
mkdir -p "$HOME/Library/LaunchAgents"
cat > "$PLIST" <<PL
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>Label</key><string>com.voicestudio.obsrelay</string>
<key>ProgramArguments</key><array><string>$BIN</string></array>
<key>RunAtLoad</key><true/>
</dict></plist>
PL
launchctl unload "$PLIST" 2>/dev/null; launchctl load "$PLIST"
echo "Done. The relay will start at login."
