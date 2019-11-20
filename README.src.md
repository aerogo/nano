# {name}

{go:header}

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

{go:footer}
