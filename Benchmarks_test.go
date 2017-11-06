package nano_test

import (
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aerogo/nano"
)

func BenchmarkCollectionGet(b *testing.B) {
	node := nano.New(port)
	db := node.Namespace("test").RegisterTypes(types...)
	defer node.Close()
	defer node.Clear()

	users := db.Collection("User")
	users.Set("1", newUser(1))

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			users.Get("1")
		}
	})
	b.StopTimer()
}

func BenchmarkCollectionSet(b *testing.B) {
	node := nano.New(port)
	db := node.Namespace("test").RegisterTypes(types...)
	defer node.Close()
	defer node.Clear()

	users := db.Collection("User")
	example := newUser(1)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			users.Set("1", example)
		}
	})
	b.StopTimer()
}

func BenchmarkCollectionDelete(b *testing.B) {
	node := nano.New(port)
	db := node.Namespace("test").RegisterTypes(types...)
	defer node.Close()
	defer node.Clear()

	users := db.Collection("User")

	for i := 0; i < 10000; i++ {
		users.Set(strconv.Itoa(i), newUser(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			users.Delete("42")
		}
	})
	b.StopTimer()
}

func BenchmarkCollectionAll(b *testing.B) {
	node := nano.New(port)
	db := node.Namespace("test").RegisterTypes(types...)
	defer node.Close()
	defer node.Clear()

	users := db.Collection("User")

	for i := 0; i < 1; i++ {
		users.Set(strconv.Itoa(i), newUser(i))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _ = range users.All() {
			// ...
		}
	}

	b.StopTimer()
}

func BenchmarkClusterGet(b *testing.B) {
	// Create cluster
	nodes := make([]*nano.Node, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New(port)
		nodes[i].Namespace("test").RegisterTypes(types...)
	}

	// Wait for clients to connect
	for nodes[0].Server().ClientCount() < nodeCount-1 {
		time.Sleep(10 * time.Millisecond)
	}

	i := int64(0)

	// Run benchmark
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&i, 1)
			id := int(atomic.LoadInt64(&i))
			nodes[id%nodeCount].Namespace("test").Get("User", strconv.Itoa(id))
		}
	})
	b.StopTimer()

	// Cleanup
	for i := nodeCount - 1; i >= 0; i-- {
		nodes[i].Clear()
		nodes[i].Close()
	}
}

func BenchmarkClusterSet(b *testing.B) {
	// Create cluster
	nodes := make([]*nano.Node, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New(port)
		nodes[i].Namespace("test").RegisterTypes(types...)
	}

	// Wait for clients to connect
	for nodes[0].Server().ClientCount() < nodeCount-1 {
		time.Sleep(10 * time.Millisecond)
	}

	i := int64(0)

	// Run benchmark
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&i, 1)
			id := int(atomic.LoadInt64(&i))
			nodes[id%nodeCount].Namespace("test").Set("User", strconv.Itoa(id), newUser(id))
		}
	})
	b.StopTimer()

	// Cleanup
	for i := nodeCount - 1; i >= 0; i-- {
		nodes[i].Clear()
		nodes[i].Close()
	}
}

func BenchmarkClusterDelete(b *testing.B) {
	// Create cluster
	nodes := make([]*nano.Node, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New(port)
		nodes[i].Namespace("test").RegisterTypes(types...)
	}

	// Wait for clients to connect
	for nodes[0].Server().ClientCount() < nodeCount-1 {
		time.Sleep(10 * time.Millisecond)
	}

	i := int64(0)

	// Run benchmark
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&i, 1)
			id := int(atomic.LoadInt64(&i))
			nodes[id%nodeCount].Namespace("test").Delete("User", strconv.Itoa(id))
		}
	})
	b.StopTimer()

	// Cleanup
	for i := nodeCount - 1; i >= 0; i-- {
		nodes[i].Clear()
		nodes[i].Close()
	}
}
