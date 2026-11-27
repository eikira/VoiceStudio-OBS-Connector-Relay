VoiceStudio OBS Connector Relay 0.1.2 - Windows 10 / 11
=======================================================

What it is: a small background program that connects the VoiceStudio+SimpleVoice OBS Connector
Chrome extension to OBS. One file, nothing to install.

START   Double-click  VoiceStudio-OBS-Connector-Relay.exe
        (32-bit Windows: VoiceStudio-OBS-Connector-Relay-32bit.exe)
        Nothing opens in the taskbar. A small icon (with a white arrow) appears in the
        notification area next to the clock. On Windows 10/11 it may be under the ^ arrow:
        drag it out to keep it visible.
STOP    Double-click "Stop Relay.vbs", or right-click the tray icon -> "Stop relay and exit".

Tray icon menu: Open status page, Open relay folder, Show log, Change port..., Listening address (network)...,
Check for updates daily, Check for updates now, Start with Windows, Stop relay and exit.
Double-click the icon to open the status page.

UPDATES
  Once a day the relay asks GitHub if a newer version exists. If so, a message pops up at the
  icon and the menu gets an entry "Update available: x.y.z - open download page".
  NOTHING is installed automatically. To update:
   1. Right-click the icon -> "Stop relay and exit".
   2. Download the new zip, unzip it, and copy everything INSIDE the unzipped folder (not the folder itself)
      over the old files in this folder.
      (Keep your own settings.txt if you changed the port.)
   3. Start VoiceStudio-OBS-Connector-Relay.exe again.
  "Open relay folder" in the menu shows you where the program is (the extension settings have
  the same button). Switch the daily check off in the menu or with UPDATECHECK=off in settings.txt.

VOICESTUDIO ADDRESS (where this relay finds VoiceStudio)
  The extension and OBS talk only to this relay. The relay passes the voices, the models and the speech
  on to VoiceStudio. Default http://127.0.0.1:3900 (VoiceStudio on this computer). If VoiceStudio runs on
  another computer, enter its address on the page of the listening address, section "VoiceStudio address"
  (Windows: tray menu "VoiceStudio address..."), or run  --set-voicestudio http://ADDRESS:3900 .
  VoiceStudio itself must accept connections from the network in that case.

LISTENING ADDRESS (for OBS or the extension on another computer)
  By default the relay answers only on this PC. To use it from another computer, choose the
  address of your real home or office network (NOT a VPN adapter):
  Right-click the tray icon -> "Listening address (network)...". The page opens in your browser.
  The page lists the addresses of this computer with the name of each adapter. VPN and virtual
  adapters are marked. After saving, the relay restarts. Then set "Relay address" in the extension
  (on the other computer) to the address shown on the page, for example http://192.168.1.20:8787,
  and copy the new OBS Browser Source URL.
  IMPORTANT: the extension's "Detect" button only looks on one port. If you changed the port, set
  the same port in the extension, or Detect will not find the relay.
  There is no password: anyone on that network who knows your OBS ID can send speech. Use only a
  network you trust.

CHANGE THE PORT (default 8787)
  Right-click the tray icon -> "Change port...", type a number from 1024 to 65535, OK.
  The relay restarts by itself. Then:
   1. In the extension settings set "Relay address" to http://127.0.0.1:NEWPORT and Save.
   2. In OBS replace the Browser Source URL with the new one (Copy in the extension settings).

First time:
 - Windows SmartScreen may say "Windows protected your PC" because the file is new and not
   signed. Click "More info" -> "Run anyway".
 - If you downloaded the zip: right-click the zip -> Properties -> Unblock, then extract.
 - If Windows Firewall asks, allow it for Private networks (only needed if OBS runs on
   another computer).

settings.txt (next to the program; stop and start the relay after editing by hand):
 PORT=8787
 HOST=127.0.0.1   only this PC. For another PC see LISTENING ADDRESS above.
 UPDATECHECK=on   or off: the daily update check.

Log: %LOCALAPPDATA%\VoiceStudioOBSRelay\relay.log  (tray menu -> Show log)
Needs Windows 10 or 11.

Needs the VoiceStudio+SimpleVoice OBS Connector Chrome extension: https://github.com/eikira/VoiceStudio-SimpleVoice-OBS-Connector
Free and open source (GPL-3.0), copyright eikira. Free to use, share and modify. No warranty.
Inspired by Social Stream Ninja. Not affiliated with VoiceStudio, OBS, Twitch, YouTube,
Rumble or Kick. See LICENSE and THIRD-PARTY-NOTICES.txt.
