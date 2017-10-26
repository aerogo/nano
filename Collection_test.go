package database_test

import (
	"strconv"
	"testing"

	"github.com/aerogo/database"
)

func BenchmarkCollectionGet(b *testing.B) {
	db := database.New()
	db.Set("User", "123", "test")

	b.ReportAllocs()
	b.ResetTimer()

	users := db.Collection("User")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			users.Get("123")
		}
	})
}

func BenchmarkCollectionAll(b *testing.B) {
	db := database.New()

	for i := 0; i < 10000; i++ {
		db.Set("User", strconv.Itoa(i), i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	users := db.Collection("User")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _ = range users.All() {
				// ...
			}
		}
	})
}
