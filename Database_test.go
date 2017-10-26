package database_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aerogo/database"
)

func TestDatabaseGet(t *testing.T) {
	db := database.New()
	db.Set("User", "1", newUser(1))
	val := db.Get("User", "1")

	assert.NotNil(t, val)
}

func TestDatabaseSet(t *testing.T) {
	db := database.New()
	db.Set("User", "1", newUser(1))

	assert.True(t, db.Exists("User", "1"))
	assert.False(t, db.Exists("User", "2"))
}

func TestDatabaseAll(t *testing.T) {
	db := database.New()

	for i := 0; i < 10000; i++ {
		db.Set("User", strconv.Itoa(i), newUser(i))
	}

	count := 0

	for user := range db.All("User") {
		assert.NotNil(t, user)
		count++
	}

	assert.Equal(t, 10000, count)
}
