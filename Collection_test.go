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
	Text      string
	Created   string
	Edited    string
	Following []string
}

func newUser(id int) *User {
	return &User{
		ID:        strconv.Itoa(id),
		Name:      "Test User",
		BirthYear: "1991",
		Text: `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam sit amet ante interdum, congue est vel, gravida odio. Praesent consequat, sem id convallis tincidunt, turpis dolor varius justo, sed consequat ante urna ac tortor. Nullam a tellus ac velit condimentum semper. Nulla et dolor a justo dignissim consectetur vel eu urna. Quisque molestie tincidunt mi non consectetur. Nulla eget faucibus lacus. Suspendisse dui lacus, volutpat vel quam ac, vehicula egestas est. Quisque a malesuada velit, mollis ullamcorper neque. Cras lobortis vitae tortor eget vehicula. Sed dictum augue vel risus eleifend, non venenatis mi vulputate. Sed laoreet accumsan enim ac porttitor. Ut blandit nibh ut ipsum ullamcorper, ut congue massa eleifend. Vivamus condimentum pharetra lorem, eget bibendum nunc porta id. Ut nulla orci, commodo id odio ac, mollis molestie ipsum.
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas sit amet dolor sit amet sem volutpat iaculis. Nunc viverra est quis sodales dictum. Fusce elementum nunc ac aliquet efficitur. Morbi id nunc sed urna dictum mattis at vitae est. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec vestibulum mauris non metus fermentum molestie.
Integer eu tortor a tellus tincidunt pretium et a urna. Integer tortor felis, rutrum vitae ipsum rutrum, laoreet maximus purus. Aliquam diam ipsum, pulvinar a leo eu, convallis ultricies urna. In consequat et eros id porta. Sed congue quam eu turpis vestibulum hendrerit. Suspendisse massa arcu, placerat sit amet tempor lobortis, ornare ut magna. Nunc sit amet gravida mi, aliquam laoreet metus. Cras non dolor at sapien euismod pulvinar ultrices eget turpis. Morbi vitae enim venenatis lacus tincidunt mollis eu non lectus. Proin nec libero porttitor, gravida turpis sed, bibendum mi. In venenatis molestie dapibus. Phasellus molestie tincidunt arcu, vel vestibulum orci dapibus in. Aliquam maximus justo eros, eu efficitur leo porta eu. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Vivamus efficitur, libero accumsan molestie blandit, velit augue mattis enim, ut elementum arcu nulla quis eros. Sed aliquet, nibh sed dapibus porta, nisi sem efficitur purus, in sollicitudin ipsum nisi sit amet nisl. In hac habitasse platea dictumst. Nulla eu odio sit amet turpis mollis mollis finibus vitae diam. Mauris vel lorem vitae erat accumsan rhoncus eget porta libero. Curabitur tincidunt id dolor ut consectetur. Donec ornare elit sed metus malesuada fringilla. Cras purus nisl, laoreet ac risus et, consectetur consequat neque. Aliquam porttitor viverra aliquam.
Etiam condimentum justo mi, eu hendrerit mauris ornare eget. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Morbi ac metus diam. Proin pulvinar orci at ex commodo, a blandit metus mollis. Phasellus tincidunt vel purus feugiat consequat. Proin vel accumsan massa. Cras suscipit neque dolor. Sed sit amet aliquet metus, non fringilla libero. Donec sagittis neque vel purus euismod, in tempus libero dignissim. Cras scelerisque vehicula bibendum. Maecenas augue orci, blandit posuere metus at, consectetur consectetur urna. Duis consectetur posuere est, vitae rutrum augue vehicula vitae. Aliquam id ornare odio.`,
		Created:   "2017-01-01",
		Edited:    "2017-01-01",
		Following: []string{"Vy2Hk5yvx", "VJOK1ckvx", "VJCuoQevx", "41oBveZPx", "41w5sjZKg", "4y1WgNMDx", "NyQph5bFe", "NJ3kffzwl", "Vy2We3bKe", "VkVaI_MPl", "V1eSUNSYx", "BJdJDFgc", "r1nTQ8Ko", "BkXadrU5"},
	}
}

var types = []interface{}{
	(*User)(nil),
}

func BenchmarkCollectionGet(b *testing.B) {
	db := database.New("db", types)
	defer db.Close()
	defer db.ClearAll()

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
	db := database.New("db", types)
	defer db.Close()
	defer db.ClearAll()

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
	db := database.New("db", types)
	defer db.Close()
	defer db.ClearAll()

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
	db := database.New("db", types)
	defer db.Close()
	defer db.ClearAll()

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
