# Changelog (VoiceStudio OBS Connector Relay)

## 0.0.6
- **Update check:** once a day the relay asks GitHub which version is the newest. Windows: a message at the tray icon and an
  "Update available" menu entry; all systems: a line in the log and a notice in the extension's settings. Nothing is installed
  automatically. Switch it off in the tray menu or with `UPDATECHECK=off` in `settings.txt`.
- **Open folder:** tray menu "Open relay folder", `--open-folder`, and a button in the extension's settings.
- `--check-update` and `--set-updates on|off`. macOS and Linux zips get a "check for updates" script.
- New `GET /info` (version and update state) and `POST /open-folder` (this computer only, needs a header a web page cannot send).

## 0.0.5
- First public release on GitHub. Same program as 0.0.4; copyright holder and links to the Chrome
  extension added.

## 0.0.4
- Released under GPL-3.0 (free and open source). License, notices and credit to Social Stream Ninja added.

## 0.0.3
- Name: VoiceStudio OBS Connector Relay (files are named VoiceStudio-OBS-Connector-Relay...).
- Change the port from the program: Windows tray menu "Change port..." (the relay restarts by
  itself); macOS "Change Port.command"; Linux `./change-port.sh PORT`; or `--set-port PORT`.
- Dropped Windows 7/8 support (Windows 10/11 only).

## 0.0.2
- Rewritten in Go as one native program per system: no Node.js and no install needed.
- Windows: silent, notification-area icon with menu, Stop Relay.vbs, optional start with Windows.
- macOS (Intel, Apple Silicon) and Linux (x64, ARM64): start/stop scripts, optional start at login.
- `--stop` closes a running relay (from this computer only).

## 0.0.1
First separate release (Node.js based).
