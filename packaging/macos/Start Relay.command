#!/bin/bash
# Starts the VoiceStudio OBS Connector Relay in the background. You can close this window afterwards.
cd "$(dirname "$0")" || exit 1
xattr -dr com.apple.quarantine . 2>/dev/null
case "$(uname -m)" in arm64) BIN=./VoiceStudio-OBS-Connector-Relay-mac-apple-silicon ;; *) BIN=./VoiceStudio-OBS-Connector-Relay-mac-intel ;; esac
chmod +x "$BIN"
DATA="$HOME/Library/Application Support/VoiceStudioOBSRelay"; mkdir -p "$DATA"
[ -f "$DATA/relay.log" ] && [ "$(stat -f%z "$DATA/relay.log")" -gt 1048576 ] && rm -f "$DATA/relay.log"
nohup "$BIN" >> "$DATA/relay.log" 2>&1 &
PID=$!
sleep 1
if kill -0 $PID 2>/dev/null; then
  echo "The relay is running in the background."
  echo "To stop it, open \"Stop Relay.command\". You can close this window."
else
  echo "The relay did not start. It may already be running, or the port is in use."
  tail -n 3 "$DATA/relay.log"
  exit 1
fi
