package main

import "testing"

func TestNewerThan(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"0.0.6", "0.0.5", true},
		{"v0.0.6", "0.0.5", true},
		{"0.0.5", "0.0.5", false},
		{"0.0.4", "0.0.5", false},
		{"0.1.0", "0.0.9", true},
		{"0.0.10", "0.0.9", true},
		{"1.0", "0.9.9", true},
		{"1.0", "1.0.0", false},
		{"1.0.1", "1.0", true},
		{"1.0-beta", "0.0.5", false},
		{"", "0.0.5", false},
		{"latest", "0.0.5", false},
		{"0.0.6", "garbage", false},
	}
	for _, c := range cases {
		if got := newerThan(c.latest, c.current); got != c.want {
			t.Errorf("newerThan(%q,%q)=%v want %v", c.latest, c.current, got, c.want)
		}
	}
}
