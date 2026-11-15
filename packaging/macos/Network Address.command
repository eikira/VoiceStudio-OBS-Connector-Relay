#!/bin/bash
# Opens the page where you choose which network address the relay listens on (for OBS on another computer).
# Choose your real home or office network, NOT a VPN. Needs the relay to be running.
cd "$(dirname "$0")" || exit 1
case "$(uname -m)" in arm64) BIN=./VoiceStudio-OBS-Connector-Relay-mac-apple-silicon ;; *) BIN=./VoiceStudio-OBS-Connector-Relay-mac-intel ;; esac
chmod +x "$BIN"
"$BIN" --open-network
