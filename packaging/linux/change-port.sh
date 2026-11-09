#!/bin/bash
# Usage: ./change-port.sh 9000   (1024-65535). Saves the port and restarts the relay.
cd "$(dirname "$0")" || exit 1
[ -z "$1" ] && { echo "Usage: ./change-port.sh PORT   (a number from 1024 to 65535)"; exit 1; }
case "$(uname -m)" in aarch64|arm64) BIN=./VoiceStudio-OBS-Connector-Relay-linux-arm64 ;; *) BIN=./VoiceStudio-OBS-Connector-Relay-linux-x64 ;; esac
chmod +x "$BIN"
"$BIN" --stop >/dev/null 2>&1   # stop first: it looks for the relay on the CURRENT port
"$BIN" --set-port "$1" || exit 1
./start.sh
echo "Now: in the extension settings set \"Relay address\" to http://127.0.0.1:$1 and Save,"
echo "and replace the OBS Browser Source URL with the new one."
