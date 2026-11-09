VoiceStudio OBS Connector Relay 0.0.6 - Linux (x64 and ARM64)
=============================================================

What it is: a small background program that connects the VoiceStudio OBS Connector
Chrome extension to OBS. Nothing to install (no Node.js needed).

START   ./start.sh            (runs in the background)
STOP    ./stop.sh
PORT    ./change-port.sh 9000 (1024-65535; default 8787). Then set the same address in the
        extension settings ("Relay address") and use the new OBS URL.
AUTO    ./autostart-on.sh     starts the relay at login (systemd user service)
        ./autostart-off.sh    removes it
If the scripts are not executable:  chmod +x *.sh

UPDATES The relay asks GitHub once a day if a newer version exists and writes it to the log.
        The extension settings also show it. NOTHING is installed automatically.
        ./check-update.sh     asks right now
        To update: ./stop.sh, then copy the new files over the old ones (keep your own
        settings.txt if you changed the port), then ./start.sh
        Switch the daily check off: put UPDATECHECK=off in settings.txt.
FOLDER  The extension settings have an "Open relay folder" button (relay must be running).

settings.txt: PORT=8787, HOST=127.0.0.1 (use 0.0.0.0 if OBS runs on another computer),
              UPDATECHECK=on (or off).
Log: ~/.local/state/voicestudio-obs-relay/relay.log

Needs the VoiceStudio OBS Connector Chrome extension: https://github.com/eikira/VoiceStudio-OBS-Connector
Free and open source (GPL-3.0), copyright eikira. Free to use, share and modify. No warranty.
Inspired by Social Stream Ninja. Not affiliated with VoiceStudio, OBS, Twitch, YouTube,
Rumble or Kick. See LICENSE and THIRD-PARTY-NOTICES.txt.
