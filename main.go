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

var (
	carbon  string
	prefix  string
	flush   int
	gen     int
	genSize int
)

func genMetrics(batch, size int) []string {
	metrics := make([]string, size)
	for i := 0; i < size; i++ {
		metrics[i] = fmt.Sprintf("%d.%d", batch, i)
	}

	return metrics
}

func flushMetrics(addr, prefix string, metrics []string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, m := range metrics {
		epoch := time.Now().Unix()
		err := carbonate(conn, prefix, m, rand.Intn(1000), epoch)
		if err != nil {
			return err
		}
	}

	log.Printf("flushed %d metrics\n", len(metrics))
	return nil
}

func carbonate(w io.ReadWriteCloser, prefix, name string, value int, epoch int64) error {
	_, err := fmt.Fprintf(w, "%s.%s %d %d\n", prefix, name, value, epoch)
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
	batch := 0

	// gen a batch of metric names
	metrics = append(metrics, genMetrics(batch, genSize)...)
	batch++

	// initial flush
	err := flushMetrics(carbon, prefix, metrics)
	if err != nil {
		log.Fatal(err)
	}

	// inner loop
	for {
		select {
		case <-flushTicker:
			err := flushMetrics(carbon, prefix, metrics)
			if err != nil {
				log.Fatal(err)
			}
		case <-genTicker:
			metrics = append(metrics, genMetrics(batch, genSize)...)
			batch++
		}
	}
}
