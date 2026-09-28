package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAsyncFileLoggerFlushesOnCloseAndAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "client.log")
	for _, line := range []string{"first\n", "second\n"} {
		logger, err := newAsyncFileLogger(path)
		if err != nil {
			t.Fatal(err)
		}
		p := []byte(line)
		if _, err := logger.Write(p); err != nil {
			t.Fatal(err)
		}
		p[0] = 'X' // The caller may reuse the buffer immediately.
		logger.Close()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "first\nsecond\n" {
		t.Fatalf("unexpected log: %q", data)
	}
}

func TestAsyncFileLoggerDropsWhenQueueIsFull(t *testing.T) {
	logger := &asyncFileLogger{queue: make(chan []byte, 1)}
	logger.Write([]byte("first"))
	logger.Write([]byte("second"))
	if got := logger.dropped.Load(); got != 1 {
		t.Fatalf("dropped %d lines, want 1", got)
	}
	if got := string(<-logger.queue); got != "first" {
		t.Fatalf("queued %q, want first", got)
	}
}

func TestAsyncFileLoggerOpenError(t *testing.T) {
	_, err := newAsyncFileLogger(filepath.Join(t.TempDir(), "missing", "client.log"))
	if err == nil || !strings.Contains(err.Error(), "client.log") {
		t.Fatalf("expected path error, got %v", err)
	}
}
