package database_test

import (
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aerogo/database"
)

func TestDatabaseGet(t *testing.T) {
	db := database.New("db", types)
	defer db.Close()
	defer db.ClearAll()

	db.Set("User", "1", newUser(1))
	db.Set("User", "2", newUser(2))

	val := db.Get("User", "1")
	user, ok := val.(*User)

	assert.True(t, ok)
	assert.Equal(t, "Test User", user.Name)

	assert.NotNil(t, val)
}

func TestDatabaseSet(t *testing.T) {
	db := database.New("db", types)
	defer db.Close()
	defer db.ClearAll()

	db.Set("User", "1", newUser(1))
	db.Delete("User", "2")

	assert.True(t, db.Exists("User", "1"))
	assert.False(t, db.Exists("User", "2"))
}

func TestDatabaseClear(t *testing.T) {
	db := database.New("db", types)
	defer db.Close()
	defer db.ClearAll()

	db.Set("User", "1", newUser(1))
	db.Set("User", "2", newUser(2))
	db.Set("User", "3", newUser(3))

	assert.True(t, db.Exists("User", "1"))
	assert.True(t, db.Exists("User", "2"))
	assert.True(t, db.Exists("User", "3"))

	db.Clear("User")

	assert.False(t, db.Exists("User", "1"))
	assert.False(t, db.Exists("User", "2"))
	assert.False(t, db.Exists("User", "3"))
}

func TestDatabaseAll(t *testing.T) {
	db := database.New("db", types)
	defer db.Close()
	defer db.ClearAll()

	db.Collection("User").Clear()

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

func TestDatabaseColdStart(t *testing.T) {
	db := database.New("db", types)

	for i := 0; i < 10000; i++ {
		db.Set("User", strconv.Itoa(i), newUser(i))
		assert.True(t, db.Exists("User", strconv.Itoa(i)))
	}

	db.Close()

	// Sync filesystem
	exec.Command("sync").Run()

	// Wait a little
	time.Sleep(1000 * time.Millisecond)

	// Cold start
	newDB := database.New("db", types)
	defer newDB.Close()
	defer newDB.ClearAll()

	for i := 0; i < 10000; i++ {
		assert.True(t, newDB.Exists("User", strconv.Itoa(i)))
	}
}
