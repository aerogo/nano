# nano

[![Godoc][godoc-image]][godoc-url]
[![Report][report-image]][report-url]
[![Tests][tests-image]][tests-url]
[![Coverage][coverage-image]][coverage-url]
[![Sponsor][sponsor-image]][sponsor-url]

High-performance database. Basically network and disk synchronized hashmaps.

## Benchmarks

```text
BenchmarkCollectionGet-12               317030264                3.75 ns/op            0 B/op          0 allocs/op
BenchmarkCollectionSet-12               11678318               102 ns/op              32 B/op          2 allocs/op
BenchmarkCollectionDelete-12            123748969                9.50 ns/op            5 B/op          0 allocs/op
BenchmarkCollectionAll-12                1403905               859 ns/op            2144 B/op          2 allocs/op
```

## API

```go
// Initialization
node := nano.New(config)
google := node.NewNamespace("google", types...)

// Usage
google.Set("User", &User{Name: "Eduard Urbach"})
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

| [![Cedric Fung](https://avatars3.githubusercontent.com/u/2269238?s=70&v=4)](https://github.com/cedricfung) | [![Scott Rayapoullé](https://avatars3.githubusercontent.com/u/11772084?s=70&v=4)](https://github.com/soulcramer) | [![Eduard Urbach](https://avatars3.githubusercontent.com/u/438936?s=70&v=4)](https://eduardurbach.com) |
| --- | --- | --- |
| [Cedric Fung](https://github.com/cedricfung) | [Scott Rayapoullé](https://github.com/soulcramer) | [Eduard Urbach](https://eduardurbach.com) |

Want to see [your own name here?](https://github.com/users/akyoto/sponsorship)

[godoc-image]: https://godoc.org/github.com/aerogo/nano?status.svg
[godoc-url]: https://godoc.org/github.com/aerogo/nano
[report-image]: https://goreportcard.com/badge/github.com/aerogo/nano
[report-url]: https://goreportcard.com/report/github.com/aerogo/nano
[tests-image]: https://cloud.drone.io/api/badges/aerogo/nano/status.svg
[tests-url]: https://cloud.drone.io/aerogo/nano
[coverage-image]: https://codecov.io/gh/aerogo/nano/graph/badge.svg
[coverage-url]: https://codecov.io/gh/aerogo/nano
[sponsor-image]: https://img.shields.io/badge/github-donate-green.svg
[sponsor-url]: https://github.com/users/akyoto/sponsorship
