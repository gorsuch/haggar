package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// listen starts a TCP listener on a random local port, closed at test end.
func listen(tb testing.TB) net.Listener {
	tb.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { ln.Close() })
	return ln
}

func TestAgentFlush(t *testing.T) {
	ln := listen(t)
	names := genMetricNames("test", 1, 5000)
	a := &Agent{ID: 1, Addr: ln.Addr().String(), MetricNames: names}

	errc := make(chan error, 1)
	go func() { errc <- a.flush() }()

	conn, err := ln.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// without -cache_connections, flush closes the connection when done
	data, err := io.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}
	if err := <-errc; err != nil {
		t.Fatalf("flush: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) != len(names) {
		t.Fatalf("got %d lines, want %d", len(lines), len(names))
	}
	epoch := strings.Fields(lines[0])[2]
	for i, line := range lines {
		f := strings.Fields(line)
		if len(f) != 3 {
			t.Fatalf("line %d: %q is not \"name value epoch\"", i, line)
		}
		if f[0] != names[i] {
			t.Fatalf("line %d: name %q, want %q", i, f[0], names[i])
		}
		if v, err := strconv.Atoi(f[1]); err != nil || v < 0 || v >= 1000 {
			t.Fatalf("line %d: value %q not in [0, 1000)", i, f[1])
		}
		if f[2] != epoch {
			t.Fatalf("line %d: epoch %s, want %s", i, f[2], epoch)
		}
	}
}

// failConn is a net.Conn whose writes always fail. It records whether it was closed.
type failConn struct {
	net.Conn
	closed bool
}

func (c *failConn) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func (c *failConn) Close() error              { c.closed = true; return nil }

func TestAgentFlushClosesConnOnWriteError(t *testing.T) {
	for _, cache := range []bool{false, true} {
		t.Run(fmt.Sprintf("cache_connections=%t", cache), func(t *testing.T) {
			cacheConns = cache
			t.Cleanup(func() { cacheConns = false })

			conn := &failConn{}
			a := &Agent{Connection: conn, MetricNames: genMetricNames("test", 1, 10)}

			if err := a.flush(); err == nil {
				t.Fatal("flush succeeded, want write error")
			}
			if !conn.closed {
				t.Error("connection was not closed after write error")
			}
			if a.Connection != nil {
				t.Error("agent kept the failed connection")
			}
		})
	}
}

func TestAgentFlushTimesOutOnStalledServer(t *testing.T) {
	ln := listen(t)
	// accept the connection but never read from it, so writes stall once
	// the socket buffers fill
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		t.Cleanup(func() { conn.Close() })
	}()

	a := &Agent{
		Addr:        ln.Addr().String(),
		Timeout:     200 * time.Millisecond,
		MetricNames: genMetricNames("test", 1, 500000), // ~20MB, far more than socket buffers hold
	}

	errc := make(chan error, 1)
	go func() { errc <- a.flush() }()

	select {
	case err := <-errc:
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			t.Fatalf("flush error = %v, want deadline exceeded", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("flush still blocked after 10s; write timeout not applied")
	}
}

func BenchmarkAgentFlush(b *testing.B) {
	ln := listen(b)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go io.Copy(io.Discard, conn)
		}
	}()

	cacheConns = true
	log.SetOutput(io.Discard)
	b.Cleanup(func() {
		cacheConns = false
		log.SetOutput(os.Stderr)
	})

	for _, n := range []int{100, 10000} {
		b.Run(fmt.Sprintf("metrics=%d", n), func(b *testing.B) {
			a := &Agent{Addr: ln.Addr().String(), MetricNames: genMetricNames("bench", 0, n)}
			for b.Loop() {
				if err := a.flush(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
