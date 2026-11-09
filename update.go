package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Update check. The relay asks GitHub once a day which release is the newest one and only TELLS the
// user (log line, tray menu, extension settings). It never downloads or installs anything.
// Nothing about the user is sent; GitHub sees the IP address like with any web request.

const repoSlug = "eikira/VoiceStudio-OBS-Connector-Relay"

// The download page that is opened for the user. Never taken from the server's answer.
const releasePage = "https://github.com/" + repoSlug + "/releases/latest"

func latestURL() string {
	if u := os.Getenv("RELAY_UPDATE_URL"); u != "" { // only used by the automatic tests
		return u
	}
	return "https://api.github.com/repos/" + repoSlug + "/releases/latest"
}

// parseVersion turns "v0.0.6" or "0.1" into numbers. Anything else (for example "1.0-beta") is refused.
func parseVersion(s string) ([]int, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(strings.TrimPrefix(s, "v"), "V")
	if s == "" {
		return nil, false
	}
	var out []int
	for _, p := range strings.Split(s, ".") {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 1000000 {
			return nil, false
		}
		out = append(out, n)
	}
	return out, true
}

// newerThan reports whether latest is a higher version than current.
func newerThan(latest, current string) bool {
	a, ok1 := parseVersion(latest)
	b, ok2 := parseVersion(current)
	if !ok1 || !ok2 {
		return false
	}
	for i := 0; i < len(a) || i < len(b); i++ {
		var x, y int
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x != y {
			return x > y
		}
	}
	return false
}

func updatesEnabled() bool {
	switch strings.ToLower(readSettings()["UPDATECHECK"]) {
	case "off", "no", "0", "false":
		return false
	}
	return true
}

type updater struct {
	mu       sync.Mutex
	latest   string
	avail    bool
	notified bool
	notifier func(latest string)
	kick     chan struct{}
}

func newUpdater() *updater { return &updater{kick: make(chan struct{}, 1)} }

func (u *updater) state() (string, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.latest, u.avail
}

// found stores a result and, the first time a newer version shows up, tells the notifier (tray balloon).
func (u *updater) found(latest string) {
	u.mu.Lock()
	u.latest = latest
	u.avail = newerThan(latest, version)
	var call func(string)
	if u.avail && !u.notified && u.notifier != nil {
		u.notified = true
		call = u.notifier
	}
	u.mu.Unlock()
	if call != nil {
		call(latest)
	}
}

func (u *updater) setNotifier(f func(string)) {
	u.mu.Lock()
	u.notifier = f
	var call func(string)
	var latest string
	if u.avail && !u.notified && f != nil {
		u.notified = true
		call, latest = f, u.latest
	}
	u.mu.Unlock()
	if call != nil {
		call(latest)
	}
}

func (u *updater) wake() {
	select {
	case u.kick <- struct{}{}:
	default:
	}
}

// fetchLatest asks GitHub for the newest release tag.
func fetchLatest() (string, error) {
	req, err := http.NewRequest(http.MethodGet, latestURL(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "VoiceStudio-OBS-Connector-Relay/"+version)
	c := &http.Client{Timeout: 10 * time.Second} // normal system proxy settings apply
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub answered %d", resp.StatusCode)
	}
	var v struct {
		Tag string `json:"tag_name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&v); err != nil {
		return "", err
	}
	if _, ok := parseVersion(v.Tag); !ok {
		return "", fmt.Errorf("unexpected version %q", v.Tag)
	}
	return strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(v.Tag), "v"), "V"), nil
}

// checkNow runs one check, remembers the result and writes it to the log.
func (r *relay) checkNow() (string, bool, error) {
	latest, err := fetchLatest()
	if err != nil {
		logf("Update check failed: %v", err)
		return "", false, err
	}
	r.upd.found(latest)
	_, avail := r.upd.state()
	if avail {
		logf("Update available: version %s (you have %s). Download: %s", latest, version, releasePage)
	} else {
		logf("No update available (you have %s, newest is %s).", version, latest)
	}
	return latest, avail, nil
}

// updateLoop checks shortly after start and then once a day, while the setting allows it.
func (r *relay) updateLoop() {
	wait := 8 * time.Second
	for {
		select {
		case <-time.After(wait):
		case <-r.upd.kick:
		case <-r.quit:
			return
		}
		wait = 24 * time.Hour
		if updatesEnabled() {
			r.checkNow()
		}
	}
}
