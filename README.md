# nano

[![Godoc reference][godoc-image]][godoc-url]
[![Go report card][goreportcard-image]][goreportcard-url]
[![Tests][travis-image]][travis-url]
[![Code coverage][codecov-image]][codecov-url]
[![License][license-image]][license-url]

High-performance database. Basically network synchronized hashmaps.

## Benchmarks

```
BenchmarkCollectionGet-8      	200000000	         7.85 ns/op	       0 B/op	       0 allocs/op
BenchmarkCollectionSet-8      	20000000	       107 ns/op	      32 B/op	       2 allocs/op
BenchmarkCollectionDelete-8   	100000000	        18.8 ns/op	       1 B/op	       0 allocs/op
BenchmarkCollectionAll-8      	    1000	   1099789 ns/op	   96553 B/op	      42 allocs/op
```

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