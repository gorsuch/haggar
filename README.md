haggar [![test](https://github.com/gorsuch/haggar/actions/workflows/test.yml/badge.svg)](https://github.com/gorsuch/haggar/actions/workflows/test.yml)
======

An experimental [carbon](https://github.com/graphite-project/carbon) load generation tool named after [Haggar the Witch](https://en.wikipedia.org/wiki/List_of_Voltron_characters), a real menace to [Voltron](https://voltron.com/).

It's also named after pants.

![](https://slimages.macys.com/is/image/MCY/products/8/optimized/1096328_fpx.tif?01AD=3T7DyZyp_siLqj1q-neozCxIommQ92M1GsNc5fe_xTNqBcjyGG2gMxA&01RI=C624DC2009B77F9&01NA=&$filterlrg$&wid=370)

It behaves like a swarm of [collectd](https://collectd.org/) agents firing a fixed number of metrics at a fixed interval to your [carbon-compatible](https://github.com/graphite-project/carbon) endpoint.  The number of agents increases over time until a maximum number is reached.  At any given time, you can pause the spawning of new agents by sending `SIGUSR1`.  Spawning can be resumed by doing the same.

## Installation

```sh
$ go install github.com/gorsuch/haggar@latest
```

## Usage

```sh
$ haggar -carbon=graphite.example.com:2003 -agents=50 -metrics=1000
```

Toggle spawning of new agents on and off while it runs:

```sh
$ kill -USR1 <pid>   # pid is logged at startup as "master: pid N"
```

## Command-line flags

```sh
$ haggar -h
Usage of haggar:
  -agents int
    	max number of agents to run concurrently (default 100)
  -cache_connections
    	if set, keep connections open between flushes (default: false)
  -carbon string
    	address of carbon host (default "localhost:2003")
  -flush-interval duration
    	how often to flush metrics (default 10s)
  -jitter duration
    	max amount of jitter to introduce in between agent launches (default 10s)
  -metrics int
    	number of metrics for each agent to hold (default 10000)
  -prefix string
    	prefix for metrics (default "haggar")
  -spawn-interval duration
    	how often to gen new agents (default 10s)
  -timeout duration
    	max time to connect to carbon, and to write each flush (0 = no timeout) (default 5s)
```

Metrics are named `<prefix>.agent.<id>.metrics.<n>` and sent in the carbon plaintext format (`name value timestamp`), with random values from 0 to 999.

## Development

Go is managed with [mise](https://mise.jdx.dev/):

```sh
$ mise install       # installs the pinned Go toolchain
$ mise run test      # go vet + go test
$ mise run build     # builds bin/haggar-<os>-<arch>
```

## Credits

This tool was designed and developed by [@gorsuch](https://github.com/gorsuch) and [@obfuscurity](https://github.com/obfuscurity).
