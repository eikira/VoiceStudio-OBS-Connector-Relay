VoiceStudio OBS Connector Relay 0.0.5 - macOS (Intel and Apple Silicon)
=======================================================================

What it is: a small background program that connects the VoiceStudio OBS Connector
Chrome extension to OBS. Nothing to install.

START   Double-click "Start Relay.command". It runs in the background; close the window.
STOP    Double-click "Stop Relay.command".
PORT    Double-click "Change Port.command" (default 8787). Then set the same address in the
        extension settings ("Relay address") and use the new OBS URL.
AUTO    "Start at Login (on).command" starts it when you log in; "(off)" removes that.

First time: macOS blocks files downloaded from the internet. Right-click "Start Relay.command"
-> Open -> Open. (Once only. The script then clears the block for the program itself.)
If it says a file is not executable, open Terminal and run:  chmod +x /path/to/folder/*

settings.txt: PORT=8787, HOST=127.0.0.1 (use 0.0.0.0 if OBS runs on another computer).
Log: ~/Library/Application Support/VoiceStudioOBSRelay/relay.log
There is no menu-bar icon on macOS.

Needs the VoiceStudio OBS Connector Chrome extension: https://github.com/eikira/VoiceStudio-OBS-Connector
Free and open source (GPL-3.0), copyright eikira. Free to use, share and modify. No warranty.
Inspired by Social Stream Ninja. Not affiliated with VoiceStudio, OBS, Twitch, YouTube,
Rumble or Kick. See LICENSE and THIRD-PARTY-NOTICES.txt.
