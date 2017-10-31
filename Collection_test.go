package nano_test

import (
	"strconv"
	"testing"

	"github.com/aerogo/nano"
)

func BenchmarkCollectionGet(b *testing.B) {
	node := nano.New(port)
	db := node.Namespace("test", types...)
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
	db := node.Namespace("test", types...)
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
	db := node.Namespace("test", types...)
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
	db := node.Namespace("test", types...)
	defer node.Close()
	defer node.Clear()

	users := db.Collection("User")

	for i := 0; i < 10000; i++ {
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
