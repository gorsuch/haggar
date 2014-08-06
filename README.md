haggar
======

experimental carbon load generation tool

## Installation

```sh
$ go get github.com/gorsuch/haggar
```

## Command-line flags

```sh
$ hagger -h
Usage of hagger:
  -carbon="localhost:2003": address of carbon host
  -flush=10000: how often to flush metrics, in millis
  -gen=10000: how often to gen new metrics, in millis
  -prefix="bench": prefix for metrics
  -size=10000: when we do gen metrics, how many should we make
```
