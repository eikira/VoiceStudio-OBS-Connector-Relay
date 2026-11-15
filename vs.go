// Copyright (C) 2026 eikira.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// VoiceStudio through the relay. The relay knows where VoiceStudio is (setting VOICESTUDIO, default
// http://127.0.0.1:3900, which is right when VoiceStudio runs on the same computer as the relay) and passes
// a few calls through: the voice list, the model list, the health check and the speech request. The
// extension and the OBS page then only need the relay's address. Nothing else is passed through.

package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultVoiceStudio = "http://127.0.0.1:3900"

var (
	vsMu   sync.RWMutex
	vsBase = defaultVoiceStudio
)

func getVoiceStudio() string  { vsMu.RLock(); defer vsMu.RUnlock(); return vsBase }
func setVoiceStudio(u string) { vsMu.Lock(); vsBase = u; vsMu.Unlock() }

// cleanVoiceStudioURL accepts http(s)://host[:port] (a trailing slash is dropped). Returns "" if it is not that.
func cleanVoiceStudioURL(s string) string {
	s = strings.TrimRight(strings.TrimSpace(s), "/")
	if len(s) > 200 {
		return ""
	}
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Hostname() == "" || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.User != nil {
		return ""
	}
	if u.Port() != "" {
		if _, err := net.LookupPort("tcp", u.Port()); err != nil {
			return ""
		}
	}
	return u.Scheme + "://" + u.Host
}

// describeErr turns a connection error into the short words the extension shows.
func describeErr(err error) string {
	why := err.Error()
	if ne, ok := err.(*url.Error); ok {
		why = ne.Err.Error()
	}
	low := strings.ToLower(why)
	switch {
	case strings.Contains(low, "refused"):
		return "ECONNREFUSED"
	case strings.Contains(low, "no such host") || strings.Contains(low, "no address associated"):
		return "ENOTFOUND"
	case strings.Contains(low, "deadline exceeded") || strings.Contains(low, "timeout") || strings.Contains(low, "timed out"):
		return "ETIMEDOUT"
	}
	return why
}

// vsGetPaths are the only GET calls that are passed through.
var vsGetPaths = map[string]string{
	"/vs/health":          "/health",
	"/vs/v1/audio/voices": "/v1/audio/voices",
	"/vs/v1/models":       "/v1/models",
}

// handleVS passes the calls through. Returns true if the request was one of them.
func (r *relay) handleVS(w http.ResponseWriter, req *http.Request, path, id string) bool {
	if !strings.HasPrefix(path, "/vs/") {
		return false
	}
	var upPath string
	var body io.Reader
	switch {
	case req.Method == http.MethodGet && vsGetPaths[path] != "":
		upPath = vsGetPaths[path]
	case req.Method == http.MethodPost && path == "/vs/v1/audio/speech":
		if !idRe.MatchString(id) {
			fail(w, 400, "bad id")
			return true
		}
		raw, err := readBody(req)
		if err != nil {
			fail(w, 413, "bad body")
			return true
		}
		upPath = "/v1/audio/speech"
		body = strings.NewReader(string(raw))
	default:
		fail(w, 404, "not found")
		return true
	}
	base := getVoiceStudio()
	timeout := 60 * time.Second
	ctx, cancel := context.WithTimeout(req.Context(), timeout)
	defer cancel()
	up, err := http.NewRequestWithContext(ctx, req.Method, base+upPath, body)
	if err != nil {
		fail(w, 500, "bad address")
		return true
	}
	if body != nil {
		up.Header.Set("Content-Type", "application/json")
	}
	resp, err := r.http.Do(up)
	if err != nil {
		u, _ := url.Parse(base)
		msg := "Could not reach VoiceStudio at " + u.Host + " (" + describeErr(err) + "). The relay looks for VoiceStudio at that address: change it on the relay's network page."
		logf("%s", msg)
		fail(w, 502, msg)
		return true
	}
	defer resp.Body.Close()
	cors(w)
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, io.LimitReader(resp.Body, 64<<20))
	return true
}

// vsStatus asks VoiceStudio's health address once (for the relay's own page).
func (r *relay) vsStatus() (ok bool, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	base := getVoiceStudio()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, base+"/health", nil)
	resp, err := r.http.Do(req)
	if err != nil {
		return false, "Not reachable (" + describeErr(err) + ")"
	}
	resp.Body.Close()
	if resp.StatusCode == 200 {
		return true, "VoiceStudio answers."
	}
	if resp.StatusCode == 503 {
		return false, "VoiceStudio is starting up (try again in a moment)."
	}
	return false, "Something answered with HTTP " + http.StatusText(resp.StatusCode) + ", but not like VoiceStudio."
}
