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

## Coding style

Please take a look at the [style guidelines](https://github.com/akyoto/quality/blob/master/STYLE.md) if you'd like to make a pull request.

## Patrons

| [![Scott Rayapoullé](https://avatars3.githubusercontent.com/u/11772084?s=70&v=4)](https://github.com/soulcramer) |
|---|
| [Scott Rayapoullé](https://github.com/soulcramer) |

Want to see [your own name here](https://www.patreon.com/eduardurbach)?

## Author

| [![Eduard Urbach on Twitter](https://gravatar.com/avatar/16ed4d41a5f244d1b10de1b791657989?s=70)](https://twitter.com/eduardurbach "Follow @eduardurbach on Twitter") |
|---|
| [Eduard Urbach](https://eduardurbach.com) |

[godoc-image]: https://godoc.org/github.com/blitzprog/home?status.svg
[godoc-url]: https://godoc.org/github.com/blitzprog/home
[report-image]: https://goreportcard.com/badge/github.com/blitzprog/home
[report-url]: https://goreportcard.com/report/github.com/blitzprog/home
[tests-image]: https://cloud.drone.io/api/badges/blitzprog/home/status.svg
[tests-url]: https://cloud.drone.io/blitzprog/home
[coverage-image]: https://codecov.io/gh/blitzprog/home/graph/badge.svg
[coverage-url]: https://codecov.io/gh/blitzprog/home
[patreon-image]: https://img.shields.io/badge/patreon-donate-green.svg
[patreon-url]: https://www.patreon.com/eduardurbach
