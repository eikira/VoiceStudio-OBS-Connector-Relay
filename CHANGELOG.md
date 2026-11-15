# Changelog (VoiceStudio OBS Connector Relay)

## 0.1.1
Works with extension 0.3.0 (for VoiceStudio the extension needs this relay, 0.1.1 or newer) and, as before, with extension 0.1.1 and 0.2.0 for everything except the new network features.

- **VoiceStudio goes through the relay.** The relay knows where VoiceStudio is (setting `VOICESTUDIO`, default `http://127.0.0.1:3900`, right when VoiceStudio runs on the same computer as the relay) and passes four calls through: `GET /vs/health`, `GET /vs/v1/audio/voices`, `GET /vs/v1/models` and `POST /vs/v1/audio/speech` (needs `?id=`). Nothing else is passed through. The extension (0.3.0) and the OBS page now only need the relay's address: no second address for VoiceStudio, no `&endpoint=` in the OBS URL, and VoiceStudio never sees a web page's request.
- **Set where the relay finds VoiceStudio:** Windows tray menu "VoiceStudio address...", the `/network` page (section "VoiceStudio address", with a check), `--set-voicestudio ADDRESS`, or `VOICESTUDIO=` in `settings.txt`. The change applies at once, no restart.
- **Listening address:** choose which network address the relay answers on, for OBS or the extension on another computer. A page that only this computer can open (`/network`) lists the addresses of this computer with the adapter name of each. Private ranges (192.168.x.x, 10.x.x.x, 172.16-31.x.x) are offered; adapters whose names look like a VPN or a virtual adapter, and the 100.64.x.x range, are marked in red; other (not private) addresses are hidden in a section with a red warning and a confirmation box. Two prominent notices: choose your real network, not a VPN; and the extension's Detect looks on one port only, so the port must match.
- Opened from: Windows tray menu "Listening address (network)...", macOS `Network Address.command`, Linux `network.sh`, or `--open-network`. Without a browser: `--list-addresses` and `--set-host ADDRESS` (`--yes-not-private` confirms a not-private address). Saving restarts the relay. The page also shows the address to enter in the extension (with the adapter name and a Copy button).
- If the saved address no longer belongs to this computer (the network changed), the relay starts on this PC only (127.0.0.1) and says why on the page and in the log.
- **Firewall question:** when the relay starts on a network address, it makes one short connection attempt from that address to the gateway of that network (port 80; the answer does not matter, and nothing goes to the internet). A firewall such as Windows Firewall or simplewall only asks about a program when it sees its traffic; listening alone shows nothing. The log says what happened ("Firewall check: ..."). Never done in "this PC only" mode. The gateway comes from Windows (adapter list) or, on Linux, from the default route of that adapter; otherwise the .1 address of the network is used.
- The panel-like calls (`/shutdown`, `/open-folder`, the network page) accept requests from this computer under any of its own addresses, not only 127.0.0.1, because with a LAN address the local PC reaches the relay through that address. The tray menu and `--stop` use the listening address.
- `GET /info` also reports `name`, `listening` (the address it listens on), `adapter` and `voicestudio` (the address it uses for VoiceStudio), so the extension can recognise this relay when it searches the network and show where it listens.
- `HOST=0.0.0.0` in `settings.txt` still works but is no longer recommended.
- Older extensions that send their own `endpoint` keep working (the OBS page still understands it).
- **Tested by the owner:** relay 0.1.1 together with extension 0.3.0 on Windows; he reported that it works as wanted (what exactly was tried is not recorded here).
- **Not verified by the author's tooling:** macOS and the Windows tray menu entries (built and checked with `go vet` only); the Windows gateway lookup; the exact behaviour of Windows Firewall and simplewall; real LAN and VPN adapters (adapter detection is by name and address range and was tested with simulated lists); a VoiceStudio on a second computer (tested with a fake VoiceStudio on one computer: voices, models, health and speech through the relay, played by the OBS page in Chrome).

## 0.0.8
- Names and links only, nothing else changed: the Chrome extension is now called **VoiceStudio+SimpleVoice OBS Connector** (it also works with SimpleVoice OBS) and lives at github.com/eikira/VoiceStudio-SimpleVoice-OBS-Connector. This relay is still the one for VoiceStudio, and its own name and address are unchanged. Works with extension 0.1.1 and 0.2.0.

## 0.0.7
- **Two-voice messages:** a request can carry an `intro` clip (for example "Paul says" in its own voice). The OBS page requests both clips, waits until both are ready, then plays intro and message one after the other. If only the intro fails, the message is still read.
- **Chat lines (`kind: chat`):** when more than `maxQueue` chat lines are waiting, the oldest are dropped. Commands are never dropped.
- `GET /info` now also reports `protocol: 2`. **Extension 0.1.1 needs this relay (0.0.7 or newer) and refuses to work with an older one.**

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
