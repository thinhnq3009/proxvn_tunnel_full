package main

import (
	"strings"
	"sync/atomic"
	"time"
	"unicode"
)

const maxStoredEndpointRunes = 256

type httpRequestLogEntry struct {
	sequence  uint64
	method    string
	endpoint  string
	latency   time.Duration
	size      uint64
	completed bool
}

// beginHTTPRequest adds the request immediately so the TUI can show in-flight
// work. Metrics are intentionally cheap: no body/header parsing is performed.
func (c *client) beginHTTPRequest(method, endpoint string) uint64 {
	if !c.uiEnabled || c.requestLogLimit <= 0 {
		return 0
	}

	sequence := atomic.AddUint64(&c.httpRequestSeq, 1)
	entry := httpRequestLogEntry{
		sequence: sequence,
		method:   strings.ToUpper(strings.TrimSpace(method)),
		endpoint: sanitizeRequestEndpoint(endpoint),
	}

	c.requestLogMu.Lock()
	if len(c.recentRequests) < c.requestLogLimit {
		c.recentRequests = append(c.recentRequests, entry)
	} else {
		copy(c.recentRequests, c.recentRequests[1:])
		c.recentRequests[len(c.recentRequests)-1] = entry
	}
	c.requestLogMu.Unlock()

	return sequence
}

func (c *client) finishHTTPRequest(sequence uint64, latency time.Duration, size uint64) {
	if sequence == 0 {
		return
	}

	c.requestLogMu.Lock()
	defer c.requestLogMu.Unlock()
	for i := len(c.recentRequests) - 1; i >= 0; i-- {
		if c.recentRequests[i].sequence != sequence {
			continue
		}
		c.recentRequests[i].latency = latency
		c.recentRequests[i].size = size
		c.recentRequests[i].completed = true
		return
	}
}

func (c *client) recentHTTPRequestSnapshot() []httpRequestLogEntry {
	c.requestLogMu.RLock()
	defer c.requestLogMu.RUnlock()

	entries := make([]httpRequestLogEntry, len(c.recentRequests))
	copy(entries, c.recentRequests)
	return entries
}

func sanitizeRequestEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if query := strings.IndexByte(endpoint, '?'); query >= 0 {
		endpoint = endpoint[:query]
	}
	endpoint = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, endpoint)
	if endpoint == "" {
		return "/"
	}

	runes := []rune(endpoint)
	if len(runes) > maxStoredEndpointRunes {
		return string(runes[:maxStoredEndpointRunes-1]) + "…"
	}
	return endpoint
}

func truncateRequestEndpoint(endpoint string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(endpoint)
	if len(runes) <= width {
		return endpoint
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}

func formatRequestLatency(entry httpRequestLogEntry) string {
	if !entry.completed {
		return "..."
	}
	if entry.latency < time.Millisecond {
		return "<1ms"
	}
	return entry.latency.Round(time.Millisecond).String()
}
