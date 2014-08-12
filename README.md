haggar
======

An experimental [carbon](https://github.com/graphite-project/carbon) load generation tool named after [Haggar the Witch](http://www.cheezey.org/voltron/haggar.htm), a real menace to [Voltron](http://www.voltron.com/).

![](http://f.cl.ly/items/050Y473L1x0j1y1s0744/Image%202014-08-07%20at%2015.08.35.png)

It behaves like a swarm of [collectd](https://collectd.org/) agents firing a fixed number of metrics at a fixed interval to your [carbon-compatible](https://github.com/graphite-project/carbon) endpoint.  The number of agents increase over time until a maximum number is reached.  At any given time, you can pause the spawning of new agents by sending `SIGUSR1`.  Spawning can be resumed by doing the same.

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
  -prefix="haggar": prefix for metrics
  -spawn-interval=10s: how often to gen new agents
```

## Credits

This tool was designed and developed by [@gorsuch](https://github.com/gorsuch) and [@obfuscurity](https://github.com/obfuscurity).
