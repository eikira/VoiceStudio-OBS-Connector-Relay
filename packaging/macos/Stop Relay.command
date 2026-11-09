#!/bin/bash
cd "$(dirname "$0")" || exit 1
case "$(uname -m)" in arm64) BIN=./VoiceStudio-OBS-Connector-Relay-mac-apple-silicon ;; *) BIN=./VoiceStudio-OBS-Connector-Relay-mac-intel ;; esac
"$BIN" --stop
