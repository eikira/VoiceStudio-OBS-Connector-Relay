#!/bin/bash
# Asks GitHub if a newer version of the relay exists (it only tells you; nothing is installed).
cd "$(dirname "$0")" || exit 1
case "$(uname -m)" in aarch64|arm64) BIN=./VoiceStudio-OBS-Connector-Relay-linux-arm64 ;; *) BIN=./VoiceStudio-OBS-Connector-Relay-linux-x64 ;; esac
"$BIN" --check-update
