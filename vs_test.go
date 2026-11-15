package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanVoiceStudioURL(t *testing.T) {
	good := map[string]string{"http://127.0.0.1:3900": "http://127.0.0.1:3900", " http://10.10.10.16:3900/ ": "http://10.10.10.16:3900", "https://vs.local:3900": "https://vs.local:3900", "http://pc": "http://pc"}
	for in, want := range good {
		if got := cleanVoiceStudioURL(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
	for _, in := range []string{"", "127.0.0.1:3900", "ftp://x:1", "http://", "http://x:3900/v1", "http://x:3900?a=b", "http://u:p@x:3900", "javascript:alert(1)", "http://x:99999"} {
		if got := cleanVoiceStudioURL(in); got != "" {
			t.Errorf("%q must be refused, got %q", in, got)
		}
	}
}

func TestVoiceStudioPassThrough(t *testing.T) {
	var seen []string
	var gotBody string
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		if r.Header.Get("Origin") != "" {
			t.Error("the relay must not forward a web page's Origin")
		}
		switch r.URL.Path {
		case "/health":
			w.Write([]byte(`{"status":"ok","version":"1.2.3"}`))
		case "/v1/audio/voices":
			w.Write([]byte(`{"voices":[{"id":"a","name":"A","type":"profile"}]}`))
		case "/v1/models":
			w.Write([]byte(`{"data":[{"id":"tts-1-hd"}]}`))
		case "/v1/audio/speech":
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			w.Header().Set("Content-Type", "audio/wav")
			w.Write([]byte("RIFFxxxx"))
		default:
			w.WriteHeader(404)
		}
	}))
	defer fake.Close()
	old := getVoiceStudio()
	defer setVoiceStudio(old)
	setVoiceStudio(fake.URL)
	r := newRelay()
	srv := httptest.NewServer(r)
	defer srv.Close()

	get := func(p string) (int, string) {
		resp, err := http.Get(srv.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}
	for p, want := range map[string]string{"/vs/health": `"status":"ok"`, "/vs/v1/audio/voices": `"voices"`, "/vs/v1/models": `tts-1-hd`} {
		if code, body := get(p); code != 200 || !strings.Contains(body, want) {
			t.Errorf("%s: %d %s", p, code, body)
		}
	}
	// nothing else is passed through
	for _, p := range []string{"/vs/", "/vs/admin", "/vs/v1/audio/speech", "/vs/../health", "/vs/v1/models/extra"} {
		if code, _ := get(p); code != 404 {
			t.Errorf("%s must be 404, got %d", p, code)
		}
	}
	// speech: needs an id, passes the body through and the audio back
	resp, _ := http.Post(srv.URL+"/vs/v1/audio/speech", "application/json", strings.NewReader(`{"input":"hi"}`))
	if resp.StatusCode != 400 {
		t.Errorf("speech without id: %d", resp.StatusCode)
	}
	resp, _ = http.Post(srv.URL+"/vs/v1/audio/speech?id=abcdefgh1234", "application/json", strings.NewReader(`{"input":"hi","voice":"a"}`))
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || string(b) != "RIFFxxxx" || resp.Header.Get("Content-Type") != "audio/wav" || gotBody != `{"input":"hi","voice":"a"}` {
		t.Errorf("speech: %d %q %q", resp.StatusCode, b, gotBody)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Error("the OBS page and the extension need CORS")
	}
	// the relay's /info says where it looks
	if _, body := get("/info"); !strings.Contains(body, fake.URL) {
		t.Errorf("/info: %s", body)
	}
	// VoiceStudio not there: a clear message
	setVoiceStudio("http://127.0.0.1:1")
	code, body := get("/vs/health")
	if code != 502 || !strings.Contains(body, "Could not reach VoiceStudio at 127.0.0.1:1") || !strings.Contains(body, "ECONNREFUSED") {
		t.Errorf("unreachable: %d %s", code, body)
	}
}

func TestVoiceStudioSettingPage(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	os.Setenv("HOME", dir)
	os.Setenv("APPDATA", dir)
	exe, _ := os.Executable()
	primary := filepath.Join(filepath.Dir(exe), "settings.txt")
	if _, err := os.Stat(primary); err == nil {
		t.Skip("settings.txt exists next to the test binary")
	}
	defer os.Remove(primary)
	old := getVoiceStudio()
	defer setVoiceStudio(old)
	r := newRelay()
	srv := httptest.NewServer(r)
	defer srv.Close()
	post := func(body string, header bool) int {
		req, _ := http.NewRequest("POST", srv.URL+"/api/voicestudio", strings.NewReader(body))
		if header {
			req.Header.Set("X-Relay-Open", "1")
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if post(`{"url":"http://10.10.10.16:3900"}`, false) != 403 {
		t.Error("without the header must be refused")
	}
	if post(`{"url":"not a url"}`, true) != 400 || post(`nope`, true) != 400 {
		t.Error("bad input must be refused")
	}
	if getVoiceStudio() != old {
		t.Error("refused input must change nothing")
	}
	if post(`{"url":"http://10.10.10.16:3900/"}`, true) != 200 {
		t.Fatal("a good address must be saved")
	}
	if getVoiceStudio() != "http://10.10.10.16:3900" || readSettings()["VOICESTUDIO"] != "http://10.10.10.16:3900" {
		t.Errorf("not saved: %q %q", getVoiceStudio(), readSettings()["VOICESTUDIO"])
	}
}
