haggar
======

experimental carbon load generation tool

## Installation

```sh
$ go get github.com/gorsuch/haggar
```

## Command-line flags

```sh
$ haggar -h
Usage of haggar:
  -agent-size=10000: number of metrics for each agent to hold
  -carbon="localhost:2003": address of carbon host
  -flush=10000: how often to flush metrics, in millis
  -gen=10000: how often to gen new agents, in millis
  -jitter=10000: max amount of jitter to introduce in between agent launches
  -max-agents=100: max number of agents to run concurrently
  -prefix="bench": prefix for metrics
```
