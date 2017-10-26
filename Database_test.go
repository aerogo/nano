package database_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aerogo/database"
)

func TestDatabaseGet(t *testing.T) {
	db := database.New()
	db.Set("User", "123", newUser())
	val := db.Get("User", "123")

	assert.NotNil(t, val)
}

func TestDatabaseSet(t *testing.T) {
	db := database.New()
	db.Set("User", "123", newUser())

	assert.True(t, db.Exists("User", "123"))
	assert.False(t, db.Exists("User", "456"))
}

func TestDatabaseAll(t *testing.T) {
	db := database.New()

	for i := 0; i < 10000; i++ {
		db.Set("User", strconv.Itoa(i), newUser())
	}

	count := 0

	for user := range db.All("User") {
		assert.NotNil(t, user)
		count++
	}

	assert.Equal(t, 10000, count)
}
