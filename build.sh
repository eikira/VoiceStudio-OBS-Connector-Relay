#!/bin/bash
# Builds the relay for every system and assembles the download zips into ./dist
# Needs: Go 1.21+ and zip. No other tools or downloads.
set -e
cd "$(dirname "$0")"
VER=$(grep -o 'const version = "[^"]*"' main.go | cut -d'"' -f2)
NAME=VoiceStudio-OBS-Connector-Relay
LD="-s -w"
rm -rf dist; mkdir -p dist
build() { GOOS=$1 GOARCH=$2 go build -trimpath -ldflags "$4 $LD" -o "$3" .; }

W=dist/win/$NAME; mkdir -p $W
build windows amd64 $W/$NAME.exe "-H windowsgui"
build windows 386   $W/$NAME-32bit.exe "-H windowsgui"
cp packaging/windows/README.txt $W/; cp packaging/windows/Stop-Relay.vbs "$W/Stop Relay.vbs"; cp packaging/settings.txt $W/

M=dist/mac/$NAME; mkdir -p $M
build darwin amd64 $M/$NAME-mac-intel ""
build darwin arm64 $M/$NAME-mac-apple-silicon ""
cp packaging/macos/* $M/; cp packaging/settings.txt $M/

L=dist/linux/$NAME; mkdir -p $L
build linux amd64 $L/$NAME-linux-x64 ""
build linux arm64 $L/$NAME-linux-arm64 ""
cp packaging/linux/* $L/; cp packaging/settings.txt $L/

for d in $W $M $L; do cp CHANGELOG.md LICENSE THIRD-PARTY-NOTICES.txt $d/; done
# The folder inside every zip has the same name for every version (no version number), so an update can simply replace it.
(cd dist/win && zip -qr ../$NAME-$VER-Windows.zip $NAME)
(cd dist/mac && zip -qr ../$NAME-$VER-macOS.zip $NAME)
(cd dist/linux && zip -qr ../$NAME-$VER-Linux.zip $NAME)
ls -la dist/*.zip
