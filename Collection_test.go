package database_test

import (
	"strconv"
	"testing"

	"github.com/aerogo/database"
)

type User struct {
	ID        string
	Name      string
	BirthYear string
}

func newUser(id int) *User {
	return &User{
		ID:        strconv.Itoa(id),
		Name:      "Test User",
		BirthYear: "1991",
	}
}

func BenchmarkCollectionGet(b *testing.B) {
	db := database.New()
	db.Set("User", "1", newUser(1))

	b.ReportAllocs()
	b.ResetTimer()

	users := db.Collection("User")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			users.Get("1")
		}
	})
}

func BenchmarkCollectionSet(b *testing.B) {
	db := database.New()
	users := db.Collection("User")
	example := newUser(1)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			users.Set("1", example)
		}
	})
}

func BenchmarkCollectionDelete(b *testing.B) {
	db := database.New()
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
}

func BenchmarkCollectionAll(b *testing.B) {
	db := database.New()
	users := db.Collection("User")

	for i := 0; i < 10000; i++ {
		users.Set(strconv.Itoa(i), newUser(i))
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
