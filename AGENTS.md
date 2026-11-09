# AGENTS.md: VoiceStudio OBS Connector Relay

Notes for AI coding assistants and new contributors. Read this before changing code.
User-facing docs are in README.md. Version history is in CHANGELOG.md.

## What this is

A small native program written in Go (standard library only, one binary, no installer, no runtime needed).
It connects the Chrome extension to OBS and runs on Windows 10/11 (tray icon, no window), macOS and Linux.
The extension is a separate project: https://github.com/eikira/VoiceStudio-OBS-Connector

```
Extension --POST /send?id=ID--> Relay --SSE /events?id=ID--> OBS Browser Source (overlay.html)
                                  ^                                        |
                                  +---------- POST /tts (proxy) -----------+--> VoiceStudio /v1/audio/speech
```

- Default address `127.0.0.1:8787`. Only this computer can reach it unless the user sets `--host 0.0.0.0`.
- The ID is the secret that pairs one extension with one OBS source (8-64 chars `[A-Za-z0-9_-]`).
- **Why the relay exists:** VoiceStudio refuses requests that come from web pages. A request made by this program is
  not one, so `/tts` forwards it. Only paths ending in `/audio/speech` are allowed.

## Files

| Path | Role |
|---|---|
| `main.go` | Version, flags (`--port --host --set-port --stop --restarted --version`), settings file, start and restart logic. |
| `server.go` | HTTP server: `/send`, `/events` (SSE hub), `/status`, `/tts` proxy, `/shutdown`, static overlay files. |
| `overlay.html`, `overlay.js` | The OBS Browser Source page. Embedded into the binary with `//go:embed`. Plays the speech. |
| `tray_windows.go` | Windows tray icon, menu, "Change port" window, autostart registry key. Raw Win32 calls through `syscall`, no libraries. |
| `tray_other.go` | macOS and Linux: no tray, runs in the terminal. |
| `relay.ico`, `rsrc_windows_*.syso` | Icon and Windows resources. `tools/mkico.py` and `tools/mksyso.py` regenerate them. |
| `build.sh` | Cross-compiles all targets and makes the zips (`VoiceStudio-OBS-Connector-Relay-<version>-<OS>.zip`). |
| `packaging/` | Per-OS files that go into the zips (start/stop/change-port scripts, README.txt). |
| `tests/relay.test.js` | Tests that start the compiled binary and talk to it over HTTP. |

## Design rules (keep these)

- **Standard library only.** No third-party Go modules. `go.mod` has no requirements.
- **Small binary, quiet.** Windows build uses `-H windowsgui` so no console window or taskbar button appears.
- **Windows 10 and 11 only.** Windows 7/8 are not supported on purpose (current Go needs Windows 10).
- **Loopback by default.** The `/shutdown` endpoint only works from loopback and needs header `X-Relay-Stop: 1`.
- **Proxy errors** are mapped to the same short codes the overlay expects: `ECONNREFUSED`, `ENOTFOUND`, `ETIMEDOUT`.
- **GPL-3.0.** Keep the license, the copyright header (eikira), `THIRD-PARTY-NOTICES.txt` (Go's BSD license) and the credit to
  Social Stream Ninja (inspiration only; no SSN code).
- Keep the HTTP interface compatible: the extension and overlay depend on `/send`, `/events`, `/status`, `/tts`.

## Settings

The chosen port is saved in `settings.txt` next to the executable, or in the user config folder if that is not writable.
`--set-port N` writes it, `--stop` stops a running relay. Changing the port in the packaged scripts stops the old relay first, then saves.

## Build and test

```
go vet ./...
./build.sh                 # all platforms into dist/
go build -o relay-test .   # test binary the tests look for in the repo root
node --test                # Node 18+, run from the repo root, do not pass a directory
```

The tests start the binary `relay-test` in the repo root, or the one named in the `RELAY_BIN` environment variable.

## Known gaps

- The **Windows tray icon, menu and "Change port" window were only compiled and checked for struct sizes.**
  They were never run on a real Windows machine by the author's tooling. Test on Windows 10/11 after any change in `tray_windows.go`.
- macOS and Linux builds are tested with the tests above but there is no tray on those systems by design.
- There is no automatic update. People download new releases by hand.

## Releasing

Raise the version in `main.go`, add a short entry to `CHANGELOG.md`, run `./build.sh`, and attach the three zips to a new
GitHub release with a matching tag. The changelog starts at 0.0.5, the first public version.
