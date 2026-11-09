#!/bin/bash
# Starts the relay at login (systemd user service).
cd "$(dirname "$0")" || exit 1
DIR="$(pwd)"
case "$(uname -m)" in aarch64|arm64) BIN="$DIR/VoiceStudio-OBS-Connector-Relay-linux-arm64" ;; *) BIN="$DIR/VoiceStudio-OBS-Connector-Relay-linux-x64" ;; esac
chmod +x "$BIN"
mkdir -p "$HOME/.config/systemd/user"
cat > "$HOME/.config/systemd/user/voicestudio-obs-relay.service" <<UNIT
[Unit]
Description=VoiceStudio OBS Connector Relay
[Service]
ExecStart=$BIN
Restart=on-failure
[Install]
WantedBy=default.target
UNIT
systemctl --user daemon-reload && systemctl --user enable --now voicestudio-obs-relay.service \
  && echo "Done. The relay now starts at login (undo with ./autostart-off.sh)."
