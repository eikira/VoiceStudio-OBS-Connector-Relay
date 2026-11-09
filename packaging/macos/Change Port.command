#!/bin/bash
# Asks for a new port, saves it and restarts the relay.
cd "$(dirname "$0")" || exit 1
case "$(uname -m)" in arm64) BIN=./VoiceStudio-OBS-Connector-Relay-mac-apple-silicon ;; *) BIN=./VoiceStudio-OBS-Connector-Relay-mac-intel ;; esac
chmod +x "$BIN"
PORT=$(osascript -e 'text returned of (display dialog "New port (1024 - 65535):" default answer "8787" with title "VoiceStudio OBS Connector Relay")' 2>/dev/null)
[ -z "$PORT" ] && exit 0
"$BIN" --stop >/dev/null 2>&1   # stop first: it looks for the relay on the CURRENT port
"$BIN" --set-port "$PORT" || exit 1
"$(dirname "$0")/Start Relay.command"
echo
echo "Now: in the extension settings set \"Relay address\" to http://127.0.0.1:$PORT and Save,"
echo "and replace the OBS Browser Source URL with the new one."
