// VoiceStudio OBS Connector Relay: connects the VoiceStudio+SimpleVoice OBS Connector Chrome extension to OBS.
// One small program, no install. See README.
//
// Copyright (C) 2026 eikira.
// SPDX-License-Identifier: GPL-3.0-or-later
// Free software: you can redistribute it and/or modify it under the terms of the GNU General
// Public License v3 or later. It comes with NO WARRANTY. See the LICENSE file.
// Inspired by Social Stream Ninja.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const version = "0.1.1"

var logOut io.Writer = os.Stdout

func logf(format string, args ...interface{}) {
	fmt.Fprintf(logOut, "["+time.Now().Format("15:04:05")+"] "+format+"\n", args...)
}

type config struct{ host, port string }

// activeCfg is what the running relay uses (set once at start; the network page shows it).
var activeCfg config

// appDir is the folder the program runs from (where settings.txt normally lives too).
func appDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}

// settings.txt lives next to the program (PORT=8787 / HOST=127.0.0.1). If that folder cannot be
// written (for example Program Files), the settings are kept in the user's config folder instead.
func settingsPaths() (primary, fallback string) {
	if exe, err := os.Executable(); err == nil {
		primary = filepath.Join(filepath.Dir(exe), "settings.txt")
	}
	if d, err := os.UserConfigDir(); err == nil {
		fallback = filepath.Join(d, "VoiceStudioOBSConnectorRelay", "settings.txt")
	}
	return
}

func parseSettings(path string, into map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.Index(line, "="); i > 0 {
			into[strings.ToUpper(strings.TrimSpace(line[:i]))] = strings.TrimSpace(line[i+1:])
		}
	}
}

func readSettings() map[string]string {
	out := map[string]string{}
	p, f := settingsPaths()
	parseSettings(p, out)
	parseSettings(f, out) // the config-folder file only exists if the program folder was read-only
	return out
}

// writeSetting changes one KEY=value line, keeping the comments around it.
func writeSetting(key, value string) error {
	try := func(path string) error {
		if path == "" {
			return fmt.Errorf("no path")
		}
		var lines []string
		if data, err := os.ReadFile(path); err == nil {
			lines = strings.Split(strings.TrimRight(string(data), "\r\n"), "\n")
		}
		found := false
		for i, l := range lines {
			t := strings.TrimSpace(l)
			if !strings.HasPrefix(t, "#") && strings.HasPrefix(strings.ToUpper(t), key+"=") {
				lines[i] = key + "=" + value
				found = true
			}
		}
		if !found {
			lines = append(lines, key+"="+value)
		}
		os.MkdirAll(filepath.Dir(path), 0o755)
		return os.WriteFile(path, []byte(strings.Join(lines, "\r\n")+"\r\n"), 0o644)
	}
	p, f := settingsPaths()
	if err := try(p); err == nil {
		os.Remove(f) // so there is only one place that decides
		return nil
	}
	return try(f)
}

func validPort(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n >= 1024 && n <= 65535
}

func loadConfig(flagHost, flagPort string) config {
	s := readSettings()
	pick := func(flagVal, env, key, def string) string {
		if flagVal != "" {
			return flagVal
		}
		if v := os.Getenv(env); v != "" {
			return v
		}
		if v := s[key]; v != "" {
			return v
		}
		return def
	}
	return config{host: pick(flagHost, "HOST", "HOST", "127.0.0.1"), port: pick(flagPort, "PORT", "PORT", "8787")}
}

func printAddresses(cfg config) {
	fmt.Println("Addresses of this computer (adapters that are up):")
	list := listAdapters()
	if len(list) == 0 {
		fmt.Println("  none found")
	}
	for _, a := range list {
		tag := "private network"
		switch {
		case a.Kind == "shared":
			tag = "PROBABLY A VPN"
		case a.Kind == "public":
			tag = "NOT PRIVATE"
		case a.Virtual:
			tag = "LOOKS LIKE A VPN OR VIRTUAL ADAPTER"
		}
		fmt.Printf("  %-16s %-28s %s\n", a.Address, a.Adapter, tag)
	}
	fmt.Println("Choose your real home or office network, not a VPN. Save one with:  --set-host ADDRESS")
	fmt.Println("The extension's Detect looks only on one port (now " + cfg.port + "): keep the same port in the extension.")
}

// doStop asks a running relay (on this computer) to close. Returns the exit code.
func doStop(cfg config) int {
	req, _ := http.NewRequest(http.MethodPost, "http://"+dialHost(cfg)+":"+cfg.port+"/shutdown", nil)
	req.Header.Set("X-Relay-Stop", "1")
	c := &http.Client{Timeout: 4 * time.Second, Transport: &http.Transport{Proxy: nil}}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println("The relay was not running.")
		return 1
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("Something else is using port " + cfg.port + " (not this relay).")
		return 2
	}
	// wait until it is really gone
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if conn, err := net.DialTimeout("tcp", net.JoinHostPort(dialHost(cfg), cfg.port), 300*time.Millisecond); err != nil {
			break
		} else {
			conn.Close()
		}
	}
	fmt.Println("The relay has been stopped.")
	return 0
}

func main() {
	showVersion := flag.Bool("version", false, "print the version")
	stop := flag.Bool("stop", false, "stop a running relay")
	setPort := flag.String("set-port", "", "save a new port (1024-65535) to settings.txt")
	restarted := flag.Bool("restarted", false, "internal: started by the port-change restart")
	openFolder := flag.Bool("open-folder", false, "open the folder this program runs from")
	checkUpdate := flag.Bool("check-update", false, "ask GitHub if a newer version exists and print the answer")
	setUpdates := flag.String("set-updates", "", "on or off: the daily update check (saved to settings.txt)")
	setHost := flag.String("set-host", "", "save the address to listen on (an address from --list-addresses, or 127.0.0.1)")
	yesPublic := flag.Bool("yes-not-private", false, "with --set-host: confirm an address that is not in a private range")
	listAddr := flag.Bool("list-addresses", false, "list the addresses of this computer with their adapters")
	setVS := flag.String("set-voicestudio", "", "save where this relay finds VoiceStudio (default http://127.0.0.1:3900)")
	openNet := flag.Bool("open-network", false, "open the page where the listening address is chosen")
	host := flag.String("host", "", "address to listen on (default 127.0.0.1)")
	port := flag.String("port", "", "port to listen on (default 8787)")
	flag.Parse()
	cfg := loadConfig(*host, *port)

	if *showVersion {
		fmt.Println("VoiceStudio OBS Connector Relay " + version)
		fmt.Println("Free and open source software (GPL-3.0). Inspired by Social Stream Ninja.")
		return
	}
	if *setPort != "" {
		if !validPort(*setPort) {
			fmt.Println("The port must be a number from 1024 to 65535.")
			os.Exit(2)
		}
		if err := writeSetting("PORT", *setPort); err != nil {
			fmt.Println("Could not save the setting:", err)
			os.Exit(1)
		}
		fmt.Println("Port saved: " + *setPort + ". Restart the relay. Then set the same address in the extension settings (Relay address) and use the new OBS URL.")
		return
	}
	if *setVS != "" {
		u := cleanVoiceStudioURL(*setVS)
		if u == "" {
			fmt.Println("Enter an address like http://127.0.0.1:3900")
			os.Exit(2)
		}
		if err := writeSetting("VOICESTUDIO", u); err != nil {
			fmt.Println("Could not save the setting:", err)
			os.Exit(1)
		}
		fmt.Println("VoiceStudio address saved: " + u + ". Restart the relay.")
		return
	}
	if *listAddr {
		printAddresses(cfg)
		return
	}
	if *setHost != "" {
		h := strings.TrimSpace(*setHost)
		if msg := checkHost(h, *yesPublic); msg != "" {
			fmt.Println(msg)
			os.Exit(2)
		}
		if err := writeSetting("HOST", h); err != nil {
			fmt.Println("Could not save the setting:", err)
			os.Exit(1)
		}
		fmt.Println("Listening address saved: " + h + ". Restart the relay.")
		if h != "127.0.0.1" {
			fmt.Println("Then set the Relay address in the extension (on the other PC) to http://" + h + ":" + cfg.port)
		}
		return
	}
	if *openNet {
		if err := openPath("http://" + dialHost(cfg) + ":" + cfg.port + "/network"); err != nil {
			fmt.Println("Could not open the browser. Open this address yourself: http://" + dialHost(cfg) + ":" + cfg.port + "/network")
			os.Exit(1)
		}
		return
	}
	if *stop {
		os.Exit(doStop(cfg))
	}
	if *openFolder {
		if err := openPath(appDir()); err != nil {
			fmt.Println("Could not open the folder:", err)
			fmt.Println("It is: " + appDir())
			os.Exit(1)
		}
		fmt.Println("Opened " + appDir())
		return
	}
	if *setUpdates != "" {
		v := strings.ToLower(*setUpdates)
		if v != "on" && v != "off" {
			fmt.Println("Use --set-updates on  or  --set-updates off")
			os.Exit(2)
		}
		if err := writeSetting("UPDATECHECK", v); err != nil {
			fmt.Println("Could not save the setting:", err)
			os.Exit(1)
		}
		fmt.Println("Update check: " + v + ". Restart the relay for it to take effect.")
		return
	}
	if *checkUpdate {
		latest, err := fetchLatest()
		if err != nil {
			fmt.Println("Could not check for updates:", err)
			os.Exit(1)
		}
		if newerThan(latest, version) {
			fmt.Println("An update is available: " + latest + " (you have " + version + ").")
			fmt.Println("Download: " + releasePage)
			fmt.Println("To update: stop the relay, replace the files with the new download, start it again.")
		} else {
			fmt.Println("You have the newest version (" + version + ").")
		}
		return
	}

	initLog()
	if v := cleanVoiceStudioURL(firstNonEmpty(os.Getenv("VOICESTUDIO"), readSettings()["VOICESTUDIO"])); v != "" {
		setVoiceStudio(v)
	}
	hostFallbackNote = resolveHost(&cfg)
	activeCfg = cfg
	if hostFallbackNote != "" {
		logf("%s", hostFallbackNote)
	}
	r := newRelay()
	var ln net.Listener
	var err error
	for try := 0; try < 40; try++ { // after a port change the old copy needs a moment to let go
		ln, err = net.Listen("tcp", net.JoinHostPort(cfg.host, cfg.port))
		if err == nil || !*restarted {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil {
		fatal("Port " + cfg.port + " is already in use. Is the relay already running?")
		os.Exit(1)
	}
	srv := &http.Server{Handler: r}
	go srv.Serve(ln)
	go firewallProbe(cfg.host) // lets a firewall ask about this program when it listens on a network address
	go func() {
		for {
			select {
			case <-time.After(15 * time.Second):
				r.hub.ping() // keeps OBS connections alive
			case <-r.quit:
				return
			}
		}
	}()
	go r.updateLoop()
	logf("VoiceStudio OBS Connector Relay %s running on http://%s:%s", version, cfg.host, cfg.port)
	if !isLocalOnly(cfg.host) {
		logf("WARNING: listening beyond this PC (%s). Anyone on that network who knows your ID can send speech. Use only a network you trust.", cfg.host)
	}

	runUI(r, cfg) // blocks until the relay should close
	srv.Close()
	logf("Relay stopped.")
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
