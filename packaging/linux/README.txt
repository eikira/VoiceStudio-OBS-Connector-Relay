VoiceStudio OBS Connector Relay 0.1.2 - Linux (x64 and ARM64)
=============================================================

What it is: a small background program that connects the VoiceStudio+SimpleVoice OBS Connector
Chrome extension to OBS. Nothing to install (no Node.js needed).

START   ./start.sh            (runs in the background)
STOP    ./stop.sh
PORT    ./change-port.sh 9000 (1024-65535; default 8787). Then set the same address in the
        extension settings ("Relay address") and use the new OBS URL.
AUTO    ./autostart-on.sh     starts the relay at login (systemd user service)
        ./autostart-off.sh    removes it
If the scripts are not executable:  chmod +x *.sh

VOICESTUDIO ADDRESS (where this relay finds VoiceStudio)
  The extension and OBS talk only to this relay. The relay passes the voices, the models and the speech
  on to VoiceStudio. Default http://127.0.0.1:3900 (VoiceStudio on this computer). If VoiceStudio runs on
  another computer, enter its address on the page of the listening address, section "VoiceStudio address"
  (./network.sh opens the page), or run  --set-voicestudio http://ADDRESS:3900 .
  VoiceStudio itself must accept connections from the network in that case.

LISTENING ADDRESS (for OBS or the extension on another computer)
  By default the relay answers only on this PC. To use it from another computer, choose the
  address of your real home or office network (NOT a VPN adapter):
  Run ./network.sh (opens the page in your browser; the relay must be running).
  Without a browser: --list-addresses shows the addresses, --set-host ADDRESS saves one.
  The page lists the addresses of this computer with the name of each adapter. VPN and virtual
  adapters are marked. After saving, the relay restarts. Then set "Relay address" in the extension
  (on the other computer) to the address shown on the page, for example http://192.168.1.20:8787,
  and copy the new OBS Browser Source URL.
  IMPORTANT: the extension's "Detect" button only looks on one port. If you changed the port, set
  the same port in the extension, or Detect will not find the relay.
  There is no password: anyone on that network who knows your OBS ID can send speech. Use only a
  network you trust.

UPDATES The relay asks GitHub once a day if a newer version exists and writes it to the log.
        The extension settings also show it. NOTHING is installed automatically.
        ./check-update.sh     asks right now
        To update: ./stop.sh, then copy everything INSIDE the unzipped folder (not the
        folder itself) over the old ones (keep your own settings.txt if you changed the port), then ./start.sh
        Switch the daily check off: put UPDATECHECK=off in settings.txt.
FOLDER  The extension settings have an "Open relay folder" button (relay must be running).

settings.txt: PORT=8787, HOST=127.0.0.1 (see LISTENING ADDRESS above),
              UPDATECHECK=on (or off).
Log: ~/.local/state/voicestudio-obs-relay/relay.log

Needs the VoiceStudio+SimpleVoice OBS Connector Chrome extension: https://github.com/eikira/VoiceStudio-SimpleVoice-OBS-Connector
Free and open source (GPL-3.0), copyright eikira. Free to use, share and modify. No warranty.
Inspired by Social Stream Ninja. Not affiliated with VoiceStudio, OBS, Twitch, YouTube,
Rumble or Kick. See LICENSE and THIRD-PARTY-NOTICES.txt.
