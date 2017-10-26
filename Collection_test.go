package database_test

import (
	"strconv"
	"testing"

	"github.com/aerogo/database"
)

type User struct {
	Name string
}

func BenchmarkCollectionGet(b *testing.B) {
	db := database.New()
	db.Set("User", "123", &User{})

	b.ReportAllocs()
	b.ResetTimer()

	users := db.Collection("User")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			users.Get("123")
		}
	})
}

func BenchmarkCollectionSet(b *testing.B) {
	db := database.New()
	users := db.Collection("User")
	example := &User{}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			users.Set("123", example)
		}
	})
}

func BenchmarkCollectionAll(b *testing.B) {
	db := database.New()
	users := db.Collection("User")

	for i := 0; i < 10000; i++ {
		users.Set(strconv.Itoa(i), &User{})
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _ = range users.All() {
				// ...
			}
		}
	})
}
