package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"
)

// config vars, to be manipulated via command line flags
var (
	carbon  string
	prefix  string
	flush   int
	gen     int
	genSize int
)

// generate N metrics, prefixed with batchID for tracking
func genMetrics(prefix string, batchID, n int) []string {
	metrics := make([]string, n)
	for i := 0; i < n; i++ {
		metrics[i] = fmt.Sprintf("%s.%d.%d", prefix, batchID, i)
	}

	return metrics
}

// flush a collection of metrics to a carbon server located at addr
// epoch is set to now
// the value is randomly generated
func flushMetrics(addr string, metrics []string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	epoch := time.Now().Unix()
	for _, m := range metrics {
		err := carbonate(conn, m, rand.Intn(1000), epoch)
		if err != nil {
			return err
		}
	}

	log.Printf("flushed %d metrics\n", len(metrics))
	return nil
}

// actually write the data in carbon line format
func carbonate(w io.ReadWriteCloser, name string, value int, epoch int64) error {
	_, err := fmt.Fprintf(w, "%s %d %d\n", name, value, epoch)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	flag.StringVar(&carbon, "carbon", "localhost:2003", "address of carbon host")
	flag.StringVar(&prefix, "prefix", "bench", "prefix for metrics")
	flag.IntVar(&flush, "flush", 10000, "how often to flush metrics, in millis")
	flag.IntVar(&gen, "gen", 10000, "how often to gen new metrics, in millis")
	flag.IntVar(&genSize, "gen-size", 10000, "number of metrics to generate per batch")
}

func main() {
	flag.Parse()

	genTicker := time.NewTicker(time.Duration(gen) * time.Millisecond).C
	flushTicker := time.NewTicker(time.Duration(flush) * time.Millisecond).C

	metrics := make([]string, 0)
	batchID := 0

	// initial generation
	metrics = append(metrics, genMetrics(prefix, batchID, genSize)...)
	batchID++

	// initial flush
	err := flushMetrics(carbon, metrics)
	if err != nil {
		log.Fatal(err)
	}

	// inner loop
	for {
		select {
		case <-flushTicker:
			err := flushMetrics(carbon, metrics)
			if err != nil {
				log.Fatal(err)
			}
		case <-genTicker:
			metrics = append(metrics, genMetrics(prefix, batchID, genSize)...)
			batchID++
		}
	}
}
