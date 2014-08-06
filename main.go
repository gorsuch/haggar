package main

import (
	"flag"
	"fmt"
	"io"
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
)

type Agent struct {
	ID            int
	FlushInterval time.Duration
	Addr          string
	MetricNames   []string
}

func (a *Agent) Loop() {
	for {
		select {
		case <-time.NewTicker(a.FlushInterval).C:
			err := a.Flush()
			if err != nil {
				log.Printf("agent %d: %s\n", a.ID, err)
			}
		}
	}
}

func (a *Agent) Flush() error {
	conn, err := net.Dial("tcp", a.Addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	epoch := time.Now().Unix()
	for _, name := range a.MetricNames {
		err := carbonate(conn, name, rand.Intn(1000), epoch)
		if err != nil {
			return err
		}
	}

	log.Printf("agent %d: flushed %d metrics\n", a.ID, len(a.MetricNames))
	return nil
}

// generate N metric names, prefixed with batchID for tracking
func genMetricNames(prefix string, id, n int) []string {
	names := make([]string, n)
	for i := 0; i < n; i++ {
		names[i] = fmt.Sprintf("%s.agent.%d.metrics.%d", prefix, id, i)
	}

	return names
}

// actually write the data in carbon line format
func carbonate(w io.ReadWriteCloser, name string, value int, epoch int64) error {
	_, err := fmt.Fprintf(w, "%s %d %d\n", name, value, epoch)
	if err != nil {
		return err
	}
	return nil
}

func launchAgent(id, n int, flush time.Duration, addr, prefix string) {
	metricNames := genMetricNames(prefix, id, n)

	a := &Agent{ID: id, FlushInterval: time.Duration(flush), Addr: addr, MetricNames: metricNames}
	a.Loop()
}

func init() {
	flag.StringVar(&carbon, "carbon", "localhost:2003", "address of carbon host")
	flag.StringVar(&prefix, "prefix", "haggar", "prefix for metrics")
	flag.DurationVar(&flushInterval, "flush-interval", 10*time.Second, "how often to flush metrics")
	flag.DurationVar(&spawnInterval, "spawn-interval", 10*time.Second, "how often to gen new agents")
	flag.IntVar(&metrics, "metrics", 10000, "number of metrics for each agent to hold")
	flag.DurationVar(&jitter, "jitter", 10*time.Second, "max amount of jitter to introduce in between agent launches")
	flag.IntVar(&agents, "agents", 100, "max number of agents to run concurrently")
}

func main() {
	flag.Parse()

	spawnAgents := true
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)

	// start timer at 1 milli so we launch an agent right away
	timer := time.NewTimer(1 * time.Millisecond)

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
					go launchAgent(curID, agents, flushInterval, carbon, prefix)
					log.Printf("agent %d: launched\n", curID)
					curID++

					timer = time.NewTimer(spawnInterval + (time.Duration(rand.Int63n(jitter.Nanoseconds()))))
				}
			}
		}
	}
}
