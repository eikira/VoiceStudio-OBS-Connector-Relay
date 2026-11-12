# VoiceStudio OBS Connector Relay

[![License: GPL v3](https://img.shields.io/badge/license-GPLv3-blue.svg)](LICENSE)

A small background program that connects the **[VoiceStudio+SimpleVoice OBS Connector](https://github.com/eikira/VoiceStudio-SimpleVoice-OBS-Connector)** Chrome
extension to **OBS Studio**. It is the second half of a pair: the extension reads your stream chat,
this relay carries the speech request to OBS, where an invisible Browser Source plays it.

> **You need both.** This relay does nothing on its own. Get the Chrome extension here:
> **[VoiceStudio-SimpleVoice-OBS-Connector](https://github.com/eikira/VoiceStudio-SimpleVoice-OBS-Connector)**

## What it does

    Chrome extension  --->  this relay  --->  OBS Browser Source  --->  VoiceStudio (your voices)
    (reads !commands         (127.0.0.1:8787)    (plays the audio)
     from your chat)

- Gives every user an OBS Browser Source URL with their own unique ID.
- Plays requests one after another, in order, through OBS.
- Asks VoiceStudio for the audio itself when VoiceStudio refuses requests that come straight from
  a web page.
- One file, nothing to install. Runs silently in the background (Windows: notification-area icon
  with a Stop button).

## Requirements

| | |
|---|---|
| **System** | Windows 10 / 11, macOS (Intel or Apple Silicon), or Linux (x64 or ARM64) |
| **OBS Studio** | with a Browser Source (included in OBS) |
| **VoiceStudio** | running on the same computer (or in your network) with its OpenAI-compatible speech API (default `http://127.0.0.1:3900`) |
| **Chrome extension** | [VoiceStudio+SimpleVoice OBS Connector](https://github.com/eikira/VoiceStudio-SimpleVoice-OBS-Connector) (Chrome or a Chromium-based browser) |

No Node.js, Go or any other install is needed.

## Download

Get the zip for your system from the **[latest release](https://github.com/eikira/VoiceStudio-OBS-Connector-Relay/releases/latest)**:

| System | File |
|---|---|
| Windows 10 / 11 | `VoiceStudio-OBS-Connector-Relay-<version>-Windows.zip` |
| macOS | `VoiceStudio-OBS-Connector-Relay-<version>-macOS.zip` |
| Linux | `VoiceStudio-OBS-Connector-Relay-<version>-Linux.zip` |

Unzip it. You get a folder named `VoiceStudio-OBS-Connector-Relay` (the same name for every version; put it somewhere permanent). Then:

- **Windows:** double-click `VoiceStudio-OBS-Connector-Relay.exe`. Stop it with `Stop Relay.vbs` or
  the tray icon menu. Windows SmartScreen may warn about an unsigned program: *More info -> Run anyway*.
- **macOS:** right-click `Start Relay.command` -> Open (first time only). Stop with `Stop Relay.command`.
- **Linux:** `./start.sh`, stop with `./stop.sh`.

Each zip has a short README with the details (changing the port, starting at login).
The default port is **8787**; use `http://127.0.0.1:8787` as "Relay address" in the extension.

## Updating

The relay tells you when a new version is out, but **never installs anything by itself**. You update by hand:

1. Stop the relay (Windows: tray icon -> *Stop relay and exit*; macOS/Linux: the stop script).
2. Download the new zip from the [latest release](https://github.com/eikira/VoiceStudio-OBS-Connector-Relay/releases/latest) and unzip it.
3. Copy **everything inside** the unzipped folder (not the folder itself) over the old files in the folder where the relay lives. **Keep your own `settings.txt`** if you changed the port.
4. Start the relay again.

**Where is the relay's folder?** Windows: tray icon -> *Open relay folder*. Any system: in the extension's
settings click *Open relay folder* (the relay must be running), or run the program with `--open-folder`.

**How you are told:** Windows shows a message at the tray icon and adds *Update available* to its menu.
On every system the relay writes it to its log and the extension's settings show the relay's version
and whether a newer one exists. `--check-update` (or the *check-update* script in the macOS and Linux zips) asks right away.
The daily check can be switched off in the tray menu, or with `UPDATECHECK=off` in `settings.txt`.

## Build from source

    ./build.sh          # needs Go 1.21+ and zip; writes the download zips to ./dist
    go build .          # just the program for this machine

One Go program using only the standard library (no third-party code). The Windows notification-area
icon is made with direct Win32 calls. The icon on the .exe comes from `rsrc_windows_*.syso`
(regenerate with `python3 tools/mkico.py && python3 tools/mksyso.py`, needs `pip install pillow`).

Tests (need Node 18+ only to run them): `go build -o relay-test . && node --test`

## Command line

    --port N        listen on port N (default 8787)
    --host H        listen address (default 127.0.0.1 = this computer only)
    --set-port N    save the port to settings.txt (1024-65535)
    --stop          close a running relay (this computer only)
    --open-folder   open the folder the program runs from
    --check-update  ask GitHub whether a newer version exists and print the answer
    --set-updates on|off   switch the daily update check on or off (saved to settings.txt)
    --version

Settings live in `settings.txt` next to the program (`PORT`, `HOST`, `UPDATECHECK`). Command line beats
environment variables (`PORT`, `HOST`) beats the file.

## HTTP interface

| Route | Use |
|-------|-----|
| `GET /overlay.html?id=ID`, `/overlay.js` | the OBS browser source page (embedded in the program) |
| `GET /events?id=ID` | Server-Sent Events: what the OBS page listens to |
| `POST /send?id=ID` | the extension posts a speech request (validated, then sent to every OBS page with that ID). Optional fields (0.0.7): `intro` (a second clip `{text, voice, model, speed}` that is played first), `kind` (`chat` for read-all lines) and `maxQueue` (1-100, waiting chat lines before the oldest are dropped). |
| `GET /status?id=ID` | `{ok, listeners}` |
| `POST /tts?id=ID` | asks VoiceStudio for audio on behalf of the OBS page or the extension's Test buttons. Only `.../audio/speech` URLs are allowed. |
| `GET /info` | `{ok, version, protocol, updateCheck, latest, updateAvailable, downloadUrl}`: the version, the protocol number (2 since 0.0.7; extension 0.1.1+ needs 2) and what the update check found |
| `POST /open-folder` | opens the relay's own folder in the file manager. Only from this computer and only with the header `X-Relay-Open: 1`. |
| `POST /shutdown` | closes the relay. Only from this computer and only with the header `X-Relay-Stop: 1`. |

The ID is the secret: 8-64 characters `[A-Za-z0-9_-]`, generated by the extension. The relay listens on
`127.0.0.1` by default; with `HOST=0.0.0.0` anyone on the network who knows an ID can send speech.

## Support the project

This project is free and always will be. If it is useful to you and you would like to say thanks, you can:

- Follow the channel and like the videos. It costs nothing and helps a lot.
- Donate via [Ko-fi](https://ko-fi.com/V7V6175QA3), [Streamlabs](https://switzerlandplayit.com/tips) or [PayPal](https://www.paypal.com/donate?business=switzerland.playit@gmail.com&currency_code=USD)
- Send something from the wishlists: [Amazon](https://www.amazon.de/hz/wishlist/ls/3HC3NSU1P8V0Q?ref_=wl_share) or [Steam](https://store.steampowered.com/wishlist/id/eikira/)

Donations are optional and do not unlock anything. Every feature stays free for everybody.

## Privacy

The relay only talks to your own computer (or your network) and to your VoiceStudio. The one exception is the
update check: once a day it asks `api.github.com` which release is the newest (a public page). Nothing about you,
your settings or your chat is sent. GitHub sees your IP address, like any website you visit. Switch it off with
`UPDATECHECK=off` in `settings.txt` or in the tray menu.

## Free and open source

Free to use, share and modify, under the **GNU General Public License v3.0** (see [LICENSE](LICENSE)).
This software comes with no warranty. Copyright (C) 2026 eikira.

Inspired by [Social Stream Ninja](https://github.com/steveseguin/social_stream) (also GPL-3.0); no
SSN code is included. Not affiliated with or endorsed by VoiceStudio, OBS Project, Social Stream
Ninja, Twitch, YouTube, Rumble or Kick. Those names belong to their owners.
