# AGENTS.md: VoiceStudio OBS Connector Relay

Notes for AI coding assistants and new contributors. Read this before changing code.
User-facing docs are in README.md. Version history is in CHANGELOG.md.

## What this is

A small native program written in Go (standard library only, one binary, no installer, no runtime needed).
It connects the browser extension (Chrome, Firefox) to OBS and runs on Windows 10/11 (tray icon, no window), macOS and Linux.
The extension is a separate project: https://github.com/eikira/VoiceStudio-SimpleVoice-OBS-Connector

```
Extension --POST /send?id=ID--> Relay --SSE /events?id=ID--> OBS Browser Source (overlay.html)
                                  |                                        |
                                  +<--- POST /vs/v1/audio/speech ----------+
                                  |
                                  +--- /vs/* (voices, models, health, speech) ---> VoiceStudio (address: VOICESTUDIO setting)
```

- Default address `127.0.0.1:8787`. Only this computer can reach it unless the user chooses a LAN address (page `/network`, `--set-host`; see `network.go`). Private ranges are offered, VPN-looking adapters are flagged, and the saved address falls back to 127.0.0.1 if the PC no longer has it. Requests "from this computer" include the PC's own addresses (`isLocalClient`), because a LAN-only relay is reached through that address.
- The ID is the secret that pairs one extension with one OBS source (8-64 chars `[A-Za-z0-9_-]`).
- **Why the relay exists:** VoiceStudio refuses requests that come from web pages. A request made by this program is
  not one. Since 0.1.1 the extension and the OBS page talk **only to the relay**: `vs.go` passes four calls (`GET /vs/health`,
  `/vs/v1/audio/voices`, `/vs/v1/models`, `POST /vs/v1/audio/speech?id=`) to the address in `VOICESTUDIO` (default `http://127.0.0.1:3900`, set
  with `getVoiceStudio`/`setVoiceStudio`, saved by `/api/voicestudio`, `--set-voicestudio` or the tray menu). Nothing else under `/vs/` is
  passed. The older `/tts` proxy (endpoint given in the request, only paths ending in `/audio/speech`) stays for older extensions.

## Files

| Path | Role |
|---|---|
| `main.go` | Version, flags (`--port --host --set-port --set-host --yes-not-private --list-addresses --open-network --set-voicestudio --stop --open-folder --check-update --set-updates --restarted --version`), settings file, start and restart logic. |
| `network.go`, `network.html` | Listening address: adapter list (`listAdapters`, `classifyIP`, `looksVirtual`), `isLocalClient`, `dialHost`, `resolveHost` (fallback to 127.0.0.1), the page `/network` and `/api/network*`, `/api/voicestudio`. The page has the two prominent notices and the firewall note. |
| `vs.go` | VoiceStudio pass-through (`handleVS`), the VoiceStudio address (`cleanVoiceStudioURL`), `describeErr`, `vsStatus`. |
| `probe.go`, `gateway_windows.go`, `gateway_other.go` | The firewall connection attempt after start on a network address (`firewallProbe`) and the gateway lookup. |
| `proc_windows.go`, `proc_other.go` | `restartSelf` / `restartEnabled` (`RELAY_NO_RESTART=1` switches the restart off for tests). |
| `server.go` | HTTP server: `/send`, `/events` (SSE hub), `/status`, `/info`, `/tts` proxy, `/open-folder`, `/shutdown`, static overlay files. |
| `update.go`, `update_test.go` | Update check: asks GitHub for the newest release tag once a day, compares versions, remembers the result. `go test` covers the version comparison. |
| `overlay.html`, `overlay.js` | The OBS Browser Source page. Embedded into the binary with `//go:embed`. Plays the speech. |
| `tray_windows.go` | Windows tray icon, menu (status page, open folder, log, change port, update entries, autostart), "Change port" window, autostart registry key. Raw Win32 calls through `syscall`, no libraries. |
| `tray_other.go` | macOS and Linux: no tray, runs in the terminal; `openPath` uses `open` / `xdg-open`. |
| `relay.ico`, `rsrc_windows_*.syso` | Icon and Windows resources. `tools/mkico.py` and `tools/mksyso.py` regenerate them. |
| `build.sh` | Cross-compiles all targets and makes the zips (`VoiceStudio-OBS-Connector-Relay-<version>-<OS>.zip`). The folder inside each zip is always `VoiceStudio-OBS-Connector-Relay`, without version or OS, so an update can replace it. |
| `packaging/` | Per-OS files that go into the zips (start/stop/change-port scripts, README.txt). |
| `*_test.go` | Go unit tests (network, VoiceStudio pass-through, probe, update version comparison). |
| `tests/relay.test.js`, `tests/update.test.js`, `tests/protocol2.test.js`, `tests/network.test.js` | Tests that start the compiled binary and talk to it over HTTP. The update tests use a fake GitHub server and a fake file manager. |

## Design rules (keep these)

- **Standard library only.** No third-party Go modules. `go.mod` has no requirements.
- **Small binary, quiet.** Windows build uses `-H windowsgui` so no console window or taskbar button appears.
- **Windows 10 and 11 only.** Windows 7/8 are not supported on purpose (current Go needs Windows 10).
- **This computer only by default.** The listening address is `127.0.0.1` until the user chooses a LAN address. `/shutdown`, `/open-folder`, `/network` and `/api/*` only work from this computer (`isLocalClient`: loopback or one of the PC's own addresses) and the changing ones need a custom header. Offer private ranges only; flag VPN-looking adapters; non-private addresses need an explicit confirmation. Never offer `0.0.0.0` on the page.
- **Firewall probe:** one TCP attempt from the listening address to the gateway (port 80) after start on a network address, never in loopback mode, never to the internet.
- **Proxy errors** are mapped to the same short codes the overlay expects: `ECONNREFUSED`, `ENOTFOUND`, `ETIMEDOUT`.
- **GPL-3.0.** Keep the license, the copyright header (eikira), `THIRD-PARTY-NOTICES.txt` (Go's BSD license) and the credit to
  Social Stream Ninja (inspiration only; no SSN code).
- **Updates are manual.** The update check only informs (log, tray, `/info`, extension settings). Never download, replace or run files
  from the internet. The download page is the fixed `releasePage` constant, never a URL taken from GitHub's answer.
- The update check is on by default, can be switched off (`UPDATECHECK=off`, tray menu, `--set-updates`), and must stay quiet when GitHub is unreachable.
- `RELAY_UPDATE_URL` (environment variable) points the check at a fake server. It exists for the automatic tests only.
- `/open-folder` and `/shutdown` must stay protected: this computer only plus a custom header. Never add those headers to the CORS allow-list.
- **Protocol version:** `/info` reports `protocol` (`protocolVersion` in `server.go`, 2 since 0.0.7). Extension 0.1.1+ refuses to speak unless it is >= 2. Raise it when the extension needs something new from the relay; keep old fields working.
- `/send` payload extras (0.0.7): `intro`, `kind` (only `chat`), `maxQueue` (1-100, only for chat). `validate()` rejects a malformed `intro` with 400. The overlay requests both clips in parallel, waits for both, plays intro then message; chat-kind items beyond `maxQueue` drop the oldest waiting chat item.
- Keep the HTTP interface compatible: the extension and overlay depend on `/send`, `/events`, `/status`, `/vs/*`, `/tts` (older extensions) and `/info` (which the extension reads: `name`, `listening`, `adapter` and `voicestudio` too; the extension decides from `voicestudio` whether this relay can pass VoiceStudio through).

## Settings

`PORT`, `HOST` and `VOICESTUDIO` are saved in `settings.txt` next to the executable, or in the user config folder if that is not writable.
`--set-port N`, `--set-host` and `--set-voicestudio` write it, `--stop` stops a running relay. Changing the port in the packaged scripts stops the old relay first, then saves.

## Build and test

```
go vet ./...
./build.sh                 # all platforms into dist/
go build -o relay-test .   # test binary the tests look for in the repo root
go test -race ./...        # unit tests
node --test tests/*.test.js   # Node 18+, run from the repo root with the file list (the update tests take about 20 s)
```

The tests start the binary `relay-test` in the repo root, or the one named in the `RELAY_BIN` environment variable.

## Known gaps

- The **Windows tray icon, menu ("Open relay folder", the update entries, "Listening address (network)...", "VoiceStudio address...", the balloon messages) and "Change port" window were only compiled
  and checked for struct sizes by the author's tooling** (the owner tested relay 0.1.1 with extension 0.3.0 on Windows and reported that it works). The Windows gateway lookup (`gateway_windows.go`) and the firewall prompt of Windows Firewall / simplewall depend on a real Windows machine. Test on Windows 10/11 after any change in `tray_windows.go`.
- macOS and Linux builds are tested with the tests above but there is no tray on those systems by design.
- There is no automatic update on purpose (see the design rules). People download new releases by hand; the relay only tells them.

## Releasing

Raise the version in `main.go`, add a short entry to `CHANGELOG.md`, run `./build.sh`, and attach the three zips to a new
GitHub release with a matching tag. The changelog starts at 0.0.5, the first public version. Versions go 0.0.9, 0.1.0, 0.1.1 ...
