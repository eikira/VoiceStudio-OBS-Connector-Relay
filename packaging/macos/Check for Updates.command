#!/bin/bash
# Asks GitHub if a newer version of the relay exists (it only tells you; nothing is installed).
cd "$(dirname "$0")" || exit 1
case "$(uname -m)" in arm64) BIN=./VoiceStudio-OBS-Connector-Relay-mac-apple-silicon ;; *) BIN=./VoiceStudio-OBS-Connector-Relay-mac-intel ;; esac
xattr -dr com.apple.quarantine "$BIN" 2>/dev/null
chmod +x "$BIN"
"$BIN" --check-update
