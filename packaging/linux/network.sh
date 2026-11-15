#!/bin/bash
# Opens the page where you choose which network address the relay listens on (for OBS on another PC).
# Choose your real home or office network, NOT a VPN. Needs the relay to be running.
# Without a browser: ./VoiceStudio-OBS-Connector-Relay-linux-x64 --list-addresses  and  --set-host ADDRESS
cd "$(dirname "$0")" || exit 1
case "$(uname -m)" in aarch64|arm64) BIN=./VoiceStudio-OBS-Connector-Relay-linux-arm64 ;; *) BIN=./VoiceStudio-OBS-Connector-Relay-linux-x64 ;; esac
chmod +x "$BIN"
"$BIN" --open-network
