//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func initLog() {}

func fatal(msg string) { fmt.Fprintln(os.Stderr, msg) }

// No icon on macOS/Linux: run until stopped (--stop, Ctrl+C or a signal).
func runUI(r *relay, cfg config) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-r.quit:
	case <-sig:
	}
}
