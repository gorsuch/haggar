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
  -carbon="localhost:2003": address of carbon host
  -flush=10000: how often to flush metrics, in millis
  -gen=10000: how often to gen new metrics, in millis
  -gen-size=10000: number of metrics to generate per batch
  -prefix="bench": prefix for metrics
```
