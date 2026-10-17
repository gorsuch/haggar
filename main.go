package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// config vars, to be manipulated via command line flags
var (
	carbon        string
	prefix        string
	flushInterval time.Duration
	spawnInterval time.Duration
	metrics       int
	jitter        time.Duration
	agents        int
	cacheConns    bool
	timeout       time.Duration
)

type Agent struct {
	ID            int
	FlushInterval time.Duration
	Addr          string
	MetricNames   []string
	Connection    net.Conn
	Timeout       time.Duration
}

func (a *Agent) Start() {
	ticker := time.NewTicker(a.FlushInterval)
	for range ticker.C {
		if err := a.flush(); err != nil {
			log.Printf("agent %d: %s\n", a.ID, err)
		}
	}
}

func (a *Agent) flush() error {
	if a.Connection == nil {
		conn, err := net.DialTimeout("tcp", a.Addr, a.Timeout)
		if err != nil {
			return err
		}
		a.Connection = conn
	}

	if !cacheConns {
		defer a.closeConn()
	}

	// bound the whole flush so a stalled carbon can't block the agent forever
	if a.Timeout > 0 {
		if err := a.Connection.SetWriteDeadline(time.Now().Add(a.Timeout)); err != nil {
			a.closeConn()
			return err
		}
	}

	// buffer so a flush is a handful of large writes, not one per metric
	w := bufio.NewWriter(a.Connection)
	epoch := time.Now().Unix()
	for _, name := range a.MetricNames {
		err := carbonate(w, name, rand.Intn(1000), epoch)
		if err != nil {
			a.closeConn()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		a.closeConn()
		return err
	}

	log.Printf("agent %d: flushed %d metrics\n", a.ID, len(a.MetricNames))
	return nil
}

// closeConn closes and forgets the agent's connection, if any, so the next
// flush dials a fresh one.
func (a *Agent) closeConn() {
	if a.Connection != nil {
		a.Connection.Close()
		a.Connection = nil
	}
}

func launchAgent(id, n int, flush, timeout time.Duration, addr, prefix string) {
	metricNames := genMetricNames(prefix, id, n)

	a := &Agent{
		ID:            id,
		FlushInterval: flush,
		Addr:          addr,
		Timeout:       timeout,
		MetricNames:   metricNames,
	}
	a.Start()
}

func init() {
	flag.StringVar(&carbon, "carbon", "localhost:2003", "address of carbon host")
	flag.StringVar(&prefix, "prefix", "haggar", "prefix for metrics")
	flag.DurationVar(&flushInterval, "flush-interval", 10*time.Second, "how often to flush metrics")
	flag.DurationVar(&spawnInterval, "spawn-interval", 10*time.Second, "how often to gen new agents")
	flag.IntVar(&metrics, "metrics", 10000, "number of metrics for each agent to hold")
	flag.DurationVar(&jitter, "jitter", 10*time.Second, "max amount of jitter to introduce in between agent launches")
	flag.IntVar(&agents, "agents", 100, "max number of agents to run concurrently")
	flag.BoolVar(&cacheConns, "cache_connections", false, "if set, keep connections open between flushes (default: false)")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "max time to connect to carbon, and to write each flush (0 = no timeout)")
}

func main() {
	flag.Parse()

	if jitter > spawnInterval {
		flag.PrintDefaults()
		fmt.Print("\n\nError: jitter value must be less than or equal to spawn-interval\n\n")
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)

	log.Printf("master: pid %d\n", os.Getpid())

	spawnLoop(context.Background(), sigChan, agents, spawnInterval, jitter, func(id int) {
		go launchAgent(id, metrics, flushInterval, timeout, carbon, prefix)
	})
}

// spawnLoop calls launch once per spawnInterval (+/- jitter) until maxAgents
// have been launched. Each value received on toggle pauses or resumes spawning.
// It returns when ctx is done.
func spawnLoop(ctx context.Context, toggle <-chan os.Signal, maxAgents int, spawnInterval, jitter time.Duration, launch func(id int)) {
	spawnAgents := true

	// start timer at 1 milli so we launch an agent right away
	timer := time.NewTimer(1 * time.Millisecond)
	jitterDelay := time.Duration(0)
	curID := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-toggle:
			spawnAgents = !spawnAgents
			log.Printf("master: spawn_agents=%t\n", spawnAgents)
		case <-timer.C:
			if curID < maxAgents {
				if spawnAgents {
					launch(curID)
					log.Printf("agent %d: launched\n", curID)
					curID++
					// Subtract the last jitter value to keep spawnInterval a consistent baseline
					nextLaunch := spawnInterval - jitterDelay

					if jitter.Nanoseconds() != 0 {
						// rand(jitter * 2) - jitter gives random +/- jitter values
						jitterDelay = time.Duration(rand.Int63n(jitter.Nanoseconds()*2) - jitter.Nanoseconds())
					}
					nextLaunch += jitterDelay

					timer = time.NewTimer(nextLaunch)
				} else {
					// keep the timer running while paused so spawning picks back up on resume
					timer = time.NewTimer(spawnInterval)
				}
			}
		}
	}
}
