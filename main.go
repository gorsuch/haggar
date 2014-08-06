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
	carbon    string
	prefix    string
	flush     int
	gen       int
	agentSize int
	jitter    int
	maxAgents int
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
	flag.StringVar(&prefix, "prefix", "bench", "prefix for metrics")
	flag.IntVar(&flush, "flush", 10000, "how often to flush metrics, in millis")
	flag.IntVar(&gen, "gen", 10000, "how often to gen new agents, in millis")
	flag.IntVar(&agentSize, "agent-size", 10000, "number of metrics for each agent to hold")
	flag.IntVar(&jitter, "jitter", 10000, "max amount of jitter to introduce in between agent launches")
	flag.IntVar(&maxAgents, "max-agents", 100, "max number of agents to run concurrently")
}

func main() {
	flag.Parse()

	spawnAgents := true
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)

	curID := 0

	log.Printf("master: pid %d\n", os.Getpid())

	go launchAgent(curID, agentSize, time.Duration(flush)*time.Millisecond, carbon, prefix)
	log.Printf("agent %d: launched\n", curID)
	curID++

	for {
		select {
		case <-sigChan:
			spawnAgents = !spawnAgents
			log.Printf("master: spawn_agents=%t\n", spawnAgents)
		case <-time.NewTicker(time.Duration(gen) * time.Millisecond).C:
			if curID < maxAgents {
				if spawnAgents {
					// sleep for some jitter
					time.Sleep(time.Duration(rand.Intn(jitter)) * time.Millisecond)

					go launchAgent(curID, agentSize, time.Duration(flush)*time.Millisecond, carbon, prefix)
					log.Printf("agent %d: launched\n", curID)
					curID++
				}
			}
		}
	}
}
