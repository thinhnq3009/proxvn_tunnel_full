package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// asyncFileLogger keeps disk I/O away from tunnel goroutines. A full queue
// drops log lines instead of making networking wait for a slow disk.
type asyncFileLogger struct {
	mu      sync.Mutex
	queue   chan []byte
	done    chan struct{}
	closed  bool
	dropped atomic.Uint64
}

const maxFileLogLineBytes = 16 * 1024

func newAsyncFileLogger(path string) (*asyncFileLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	l := &asyncFileLogger{queue: make(chan []byte, 1024), done: make(chan struct{})}
	go l.run(file)
	return l, nil
}

func (l *asyncFileLogger) Write(p []byte) (int, error) {
	l.mu.Lock()
	if !l.closed {
		// log.Logger reuses its formatting buffer after Write returns.
		var line []byte
		if len(p) > maxFileLogLineBytes {
			line = make([]byte, maxFileLogLineBytes)
			copy(line, p[:maxFileLogLineBytes-1])
			line[len(line)-1] = '\n'
		} else {
			line = append([]byte(nil), p...)
		}
		select {
		case l.queue <- line:
		default:
			l.dropped.Add(1)
		}
	}
	l.mu.Unlock()
	return len(p), nil
}

func (l *asyncFileLogger) Close() {
	l.mu.Lock()
	if !l.closed {
		l.closed = true
		close(l.queue)
	}
	l.mu.Unlock()
	<-l.done
}

func (l *asyncFileLogger) run(file *os.File) {
	defer close(l.done)
	defer file.Close()
	writer := bufio.NewWriterSize(file, 64*1024)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var reportedError bool
	write := func(p []byte) {
		if _, err := writer.Write(p); err != nil && !reportedError {
			fmt.Fprintf(os.Stderr, "[client] ghi log file lỗi: %v\n", err)
			reportedError = true
		}
	}
	flush := func() {
		if err := writer.Flush(); err != nil && !reportedError {
			fmt.Fprintf(os.Stderr, "[client] flush log file lỗi: %v\n", err)
			reportedError = true
		}
	}
	reportDropped := func() {
		if n := l.dropped.Swap(0); n > 0 {
			write([]byte(fmt.Sprintf("[client] bỏ qua %d dòng log vì hàng đợi ghi file đã đầy\n", n)))
		}
	}
	for {
		select {
		case line, ok := <-l.queue:
			if !ok {
				reportDropped()
				flush()
				return
			}
			write(line)
		case <-ticker.C:
			reportDropped()
			flush()
		}
	}
}
