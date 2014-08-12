package main

import (
	"fmt"
	"os"
)

func ExampleGenMetricNames() {
	names := genMetricNames("test", 1, 2)
	for _, s := range names {
		fmt.Fprintf(os.Stdout, "%s\n", s)
	}

	// Output:
	// test.agent.1.metrics.0
	// test.agent.1.metrics.1
}

func ExampleCarbonate() {
	epoch := int64(1407850160)
	carbonate(os.Stdout, "test.metric.1", 100, epoch)
	// Output:
	// test.metric.1 100 1407850160
}
