#!/bin/bash
systemctl --user disable --now voicestudio-obs-relay.service 2>/dev/null
rm -f "$HOME/.config/systemd/user/voicestudio-obs-relay.service"
systemctl --user daemon-reload
echo "Done. The relay no longer starts at login."
