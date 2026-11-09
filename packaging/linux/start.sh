#!/bin/bash
# Starts the VoiceStudio OBS Connector Relay in the background.
cd "$(dirname "$0")" || exit 1
case "$(uname -m)" in aarch64|arm64) BIN=./VoiceStudio-OBS-Connector-Relay-linux-arm64 ;; *) BIN=./VoiceStudio-OBS-Connector-Relay-linux-x64 ;; esac
chmod +x "$BIN"
DATA="${XDG_STATE_HOME:-$HOME/.local/state}/voicestudio-obs-relay"; mkdir -p "$DATA"
[ -f "$DATA/relay.log" ] && [ "$(stat -c%s "$DATA/relay.log")" -gt 1048576 ] && rm -f "$DATA/relay.log"
nohup "$BIN" >> "$DATA/relay.log" 2>&1 &
PID=$!
sleep 1
if kill -0 $PID 2>/dev/null; then
  echo "The relay is running in the background. Stop it with ./stop.sh"
else
  echo "The relay did not start. It may already be running, or the port is in use."
  tail -n 3 "$DATA/relay.log"; exit 1
fi
