package main

import (
	"context"
	"os"
	"slices"
	"sync"
	"syscall"
	"testing"
	"testing/synctest"
	"time"
)

// runSpawnLoop starts spawnLoop in the current synctest bubble and returns
// the toggle channel, a func reporting launched IDs, and a func to stop it.
func runSpawnLoop(t *testing.T, maxAgents int, interval, jitter time.Duration) (chan<- os.Signal, func() []int, func()) {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())
	toggle := make(chan os.Signal)
	done := make(chan struct{})

	var mu sync.Mutex
	var launched []int

	go func() {
		defer close(done)
		spawnLoop(ctx, toggle, maxAgents, interval, jitter, func(id int) {
			mu.Lock()
			defer mu.Unlock()
			launched = append(launched, id)
		})
	}()

	got := func() []int {
		synctest.Wait()
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(launched)
	}
	stop := func() {
		cancel()
		<-done
	}
	return toggle, got, stop
}

func TestSpawnLoopLaunchesUpToMax(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		_, got, stop := runSpawnLoop(t, 3, time.Second, 0)
		defer stop()

		time.Sleep(10 * time.Second)
		if want := []int{0, 1, 2}; !slices.Equal(got(), want) {
			t.Fatalf("launched %v, want %v", got(), want)
		}
	})
}

func TestSpawnLoopPauseResume(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		toggle, got, stop := runSpawnLoop(t, 10, time.Second, 0)
		defer stop()

		// launches at 0s, 1s, 2s
		time.Sleep(2500 * time.Millisecond)
		if n := len(got()); n != 3 {
			t.Fatalf("before pause: launched %d, want 3", n)
		}

		toggle <- syscall.SIGUSR1 // pause
		time.Sleep(3 * time.Second)
		if n := len(got()); n != 3 {
			t.Fatalf("while paused: launched %d, want 3", n)
		}

		toggle <- syscall.SIGUSR1 // resume
		time.Sleep(2 * time.Second)
		if want := []int{0, 1, 2, 3, 4}; !slices.Equal(got(), want) {
			t.Fatalf("after resume: launched %v, want %v", got(), want)
		}
	})
}

func TestSpawnLoopJitterStaysInBounds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		interval, jitter := time.Second, 500*time.Millisecond

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		times := make(chan time.Time, 20)
		go spawnLoop(ctx, nil, 20, interval, jitter, func(int) { times <- time.Now() })

		prev := <-times
		for range 19 {
			next := <-times
			// each gap is interval, minus the previous jitter, plus the new jitter
			if gap := next.Sub(prev); gap < interval-2*jitter || gap > interval+2*jitter {
				t.Fatalf("gap between launches %v outside %v ± %v", gap, interval, 2*jitter)
			}
			prev = next
		}
	})
}
