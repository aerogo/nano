package nano_test

import (
	"github.com/aerogo/nano"
)

var config = nano.Configuration{
	Port: 3000,
}

var types = []interface{}{
	(*User)(nil),
}

type User struct {
	ID        int
	Name      string
	BirthYear string
	Text      string
	Created   string
	Edited    string
}

func newUser(id int) *User {
	//nolint:misspell
	return &User{
		ID:        id,
		Name:      "Test User",
		BirthYear: "1991",
		Text:      "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		Created:   "2017-01-01",
		Edited:    "2017-01-01",
	}
}
