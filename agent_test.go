package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
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
