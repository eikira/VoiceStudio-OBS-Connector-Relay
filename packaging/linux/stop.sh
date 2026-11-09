#!/bin/bash
cd "$(dirname "$0")" || exit 1
case "$(uname -m)" in aarch64|arm64) BIN=./VoiceStudio-OBS-Connector-Relay-linux-arm64 ;; *) BIN=./VoiceStudio-OBS-Connector-Relay-linux-x64 ;; esac
"$BIN" --stop
