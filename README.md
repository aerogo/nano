# nano

[![Godoc][godoc-image]][godoc-url]
[![Report][report-image]][report-url]
[![Tests][tests-image]][tests-url]
[![Coverage][coverage-image]][coverage-url]
[![Patreon][patreon-image]][patreon-url]

High-performance database. Basically network and disk synchronized hashmaps.

## Benchmarks

```text
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
* **Key**: The string that lets you look up a single object in a collection

All of the above require a unique name. Given namespace, collection and key, you can access the data stored for it.

## Style

Please take a look at the [style guidelines](https://github.com/akyoto/quality/blob/master/STYLE.md) if you'd like to make a pull request.

## Sponsors

| [![Scott Rayapoullé](https://avatars3.githubusercontent.com/u/11772084?s=70&v=4)](https://github.com/soulcramer) | [![Eduard Urbach](https://avatars2.githubusercontent.com/u/438936?s=70&v=4)](https://twitter.com/eduardurbach) |
| --- | --- |
| [Scott Rayapoullé](https://github.com/soulcramer) | [Eduard Urbach](https://eduardurbach.com) |

Want to see [your own name here?](https://www.patreon.com/eduardurbach)

[godoc-image]: https://godoc.org/github.com/aerogo/nano?status.svg
[godoc-url]: https://godoc.org/github.com/aerogo/nano
[report-image]: https://goreportcard.com/badge/github.com/aerogo/nano
[report-url]: https://goreportcard.com/report/github.com/aerogo/nano
[tests-image]: https://cloud.drone.io/api/badges/aerogo/nano/status.svg
[tests-url]: https://cloud.drone.io/aerogo/nano
[coverage-image]: https://codecov.io/gh/aerogo/nano/graph/badge.svg
[coverage-url]: https://codecov.io/gh/aerogo/nano
[patreon-image]: https://img.shields.io/badge/patreon-donate-green.svg
[patreon-url]: https://www.patreon.com/eduardurbach
