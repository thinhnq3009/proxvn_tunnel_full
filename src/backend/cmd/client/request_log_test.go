package main

import (
	"testing"
	"time"
)

func TestRequestLogKeepsNewestEntries(t *testing.T) {
	c := &client{uiEnabled: true, requestLogLimit: 2}
	first := c.beginHTTPRequest("get", "/first")
	second := c.beginHTTPRequest("post", "/second")
	third := c.beginHTTPRequest("delete", "/third")

	c.finishHTTPRequest(first, 10*time.Millisecond, 10)
	c.finishHTTPRequest(second, 20*time.Millisecond, 20)
	c.finishHTTPRequest(third, 30*time.Millisecond, 30)

	entries := c.recentHTTPRequestSnapshot()
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].sequence != second || entries[1].sequence != third {
		t.Fatalf("got sequences %d, %d; want %d, %d", entries[0].sequence, entries[1].sequence, second, third)
	}
	if !entries[1].completed || entries[1].latency != 30*time.Millisecond || entries[1].size != 30 {
		t.Fatalf("latest entry was not completed correctly: %+v", entries[1])
	}
}

func TestRequestLogDisabledOutsideTUI(t *testing.T) {
	tests := []client{
		{uiEnabled: false, requestLogLimit: 10},
		{uiEnabled: true, requestLogLimit: 0},
	}
	for i := range tests {
		if sequence := tests[i].beginHTTPRequest("GET", "/"); sequence != 0 {
			t.Fatalf("case %d: got sequence %d, want 0", i, sequence)
		}
		if len(tests[i].recentHTTPRequestSnapshot()) != 0 {
			t.Fatalf("case %d: request was stored while logging was disabled", i)
		}
	}
}

func TestSanitizeRequestEndpoint(t *testing.T) {
	got := sanitizeRequestEndpoint(" /api/users?token=secret\x1b[2J ")
	if got != "/api/users" {
		t.Fatalf("got %q, want %q", got, "/api/users")
	}
}

func TestTruncateRequestEndpoint(t *testing.T) {
	if got := truncateRequestEndpoint("/đường-dẫn-dài", 8); got != "/đường-…" {
		t.Fatalf("got %q, want %q", got, "/đường-…")
	}
}

func TestFormatRequestLatency(t *testing.T) {
	if got := formatRequestLatency(httpRequestLogEntry{}); got != "..." {
		t.Fatalf("pending latency = %q, want ...", got)
	}
	if got := formatRequestLatency(httpRequestLogEntry{completed: true, latency: 500 * time.Microsecond}); got != "<1ms" {
		t.Fatalf("sub-millisecond latency = %q, want <1ms", got)
	}
	if got := formatRequestLatency(httpRequestLogEntry{completed: true, latency: 12*time.Millisecond + 400*time.Microsecond}); got != "12ms" {
		t.Fatalf("rounded latency = %q, want 12ms", got)
	}
}
