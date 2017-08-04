package main

import (
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
)

type Agent struct {
	ID            int
	FlushInterval time.Duration
	Addr          string
	MetricNames   []string
	Connection    net.Conn
}

func (a *Agent) Start() {
	ticker := time.NewTicker(a.FlushInterval)
	for {
		select {
		case <-ticker.C:
			err := a.flush()
			if err != nil {
				log.Printf("agent %d: %s\n", a.ID, err)
			}
		}
	}
}

func (a *Agent) flush() error {
	if a.Connection == nil {
		conn, err := net.Dial("tcp", a.Addr)
		if err != nil {
			return err
		}
		a.Connection = conn
	}

	if !cacheConns {
		defer func() {
			if a.Connection != nil {
				a.Connection.Close()
				a.Connection = nil
			}
		}()
	}

	epoch := time.Now().Unix()
	for _, name := range a.MetricNames {
		err := carbonate(a.Connection, name, rand.Intn(1000), epoch)
		if err != nil {
			a.Connection = nil
			return err
		}
	}

	log.Printf("agent %d: flushed %d metrics\n", a.ID, len(a.MetricNames))
	return nil
}

func launchAgent(id, n int, flush time.Duration, addr, prefix string) {
	metricNames := genMetricNames(prefix, id, n)

	a := &Agent{
		ID:            id,
		FlushInterval: time.Duration(flush),
		Addr:          addr,
		MetricNames:   metricNames}
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
}

func main() {
	flag.Parse()

	if jitter > spawnInterval {
		flag.PrintDefaults()
		fmt.Print("\n\nError: jitter value must be less than or equal to spawn-interval\n\n")
		os.Exit(1)
	}

	spawnAgents := true
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)

	// start timer at 1 milli so we launch an agent right away
	timer := time.NewTimer(1 * time.Millisecond)
	jitterDelay := time.Duration(0)
	curID := 0

	log.Printf("master: pid %d\n", os.Getpid())

	for {
		select {
		case <-sigChan:
			spawnAgents = !spawnAgents
			log.Printf("master: spawn_agents=%t\n", spawnAgents)
		case <-timer.C:
			if curID < agents {
				if spawnAgents {
					go launchAgent(curID, metrics, flushInterval, carbon, prefix)
					log.Printf("agent %d: launched\n", curID)
					curID++
					// Subtract the last jitter val ue to keep spawnInterval a consistent baseline
					nextLaunch := spawnInterval - jitterDelay

					if jitter.Nanoseconds() != 0 {
						// rand(jitter * 2) - jitter gives random +/- jitter values
						jitterDelay = time.Duration(rand.Int63n(jitter.Nanoseconds()*2) - jitter.Nanoseconds())
					}
					nextLaunch += jitterDelay

					timer = time.NewTimer(nextLaunch)
				}
			}
		}
	}
}
