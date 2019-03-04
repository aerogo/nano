# nano

[![Godoc reference][godoc-image]][godoc-url]
[![Go report card][goreportcard-image]][goreportcard-url]
[![Tests][travis-image]][travis-url]
[![Code coverage][codecov-image]][codecov-url]
[![License][license-image]][license-url]

High-performance database. Basically network and disk synchronized hashmaps.

## Benchmarks

```
BenchmarkCollectionGet-8      	200000000	         7.16 ns/op	       0 B/op	       0 allocs/op
BenchmarkCollectionSet-8      	10000000	       117 ns/op	      32 B/op	       2 allocs/op
BenchmarkCollectionDelete-8   	100000000	        19.6 ns/op	       6 B/op	       0 allocs/op
BenchmarkCollectionAll-8      	 1000000	      1822 ns/op	    2144 B/op	       2 allocs/op
```

## Features

* Low latency commands
* Every command is "local first, sync later"
* Data is stored in memory
* Data is synchronized between all nodes in a cluster
* Data is saved to disk persistently using JSON
* Timestamp based conflict resolution
* Uses the extremely fast `sync.Map`

## Terminology

* **Namespace**: Contains multiple collections (e.g. "google")
* **Collection**: Contains homogeneous data for a data type (e.g. "User")
* **Key**: The unique string that lets you look up data in a collection

All of the above require a unique name. Given namespace, collection and key, you can access the data stored for it.

## Author

| [![Eduard Urbach on Twitter](https://gravatar.com/avatar/16ed4d41a5f244d1b10de1b791657989?s=70)](https://twitter.com/eduardurbach "Follow @eduardurbach on Twitter") |
|---|
| [Eduard Urbach](https://eduardurbach.com) |

[godoc-image]: https://godoc.org/github.com/aerogo/nano?status.svg
[godoc-url]: https://godoc.org/github.com/aerogo/nano
[goreportcard-image]: https://goreportcard.com/badge/github.com/aerogo/nano
[goreportcard-url]: https://goreportcard.com/report/github.com/aerogo/nano
[travis-image]: https://travis-ci.org/aerogo/nano.svg?branch=master
[travis-url]: https://travis-ci.org/aerogo/nano
[codecov-image]: https://codecov.io/gh/aerogo/nano/branch/master/graph/badge.svg
[codecov-url]: https://codecov.io/gh/aerogo/nano
[license-image]: https://img.shields.io/badge/license-MIT-blue.svg
[license-url]: https://github.com/aerogo/nano/blob/master/LICENSE