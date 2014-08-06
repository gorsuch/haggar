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
  -agents=100: max number of agents to run concurrently
  -carbon="localhost:2003": address of carbon host
  -flush-interval=10s: how often to flush metrics
  -jitter=10s: max amount of jitter to introduce in between agent launches
  -metrics=10000: number of metrics for each agent to hold
  -prefix="bench": prefix for metrics
  -spawn-interval=10s: how often to gen new agents
```
