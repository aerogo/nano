package nano_test

import (
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/aerogo/nano"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseGet(t *testing.T) {
	db := nano.New("test", types)
	defer db.Close()
	defer db.ClearAll()

	db.Set("User", "1", newUser(1))
	db.Set("User", "2", newUser(2))

	val, err := db.Get("User", "1")
	assert.NoError(t, err)

	user, ok := val.(*User)
	assert.True(t, ok)
	assert.Equal(t, "Test User", user.Name)

	assert.NotNil(t, val)
}

func TestDatabaseSet(t *testing.T) {
	db := nano.New("test", types)
	defer db.Close()
	defer db.ClearAll()

	db.Set("User", "1", newUser(1))
	db.Delete("User", "2")

	assert.True(t, db.Exists("User", "1"))
	assert.False(t, db.Exists("User", "2"))
}

func TestDatabaseClear(t *testing.T) {
	db := nano.New("test", types)
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
	db := nano.New("test", types)
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
	db := nano.New("test", types)

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
	newDB := nano.New("test", types)
	defer newDB.Close()
	defer newDB.ClearAll()

	for i := 0; i < 10000; i++ {
		assert.True(t, newDB.Exists("User", strconv.Itoa(i)))
	}
}

func TestDatabaseCluster(t *testing.T) {
	nodeCount := 5
	nodes := make([]*nano.Database, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New("test", types)
	}

	time.Sleep(100 * time.Millisecond)

	for i := 1; i < nodeCount; i++ {
		nodes[i].Close()
	}

	time.Sleep(100 * time.Millisecond)

	for i := 0; i < nodeCount; i++ {
		nodes[i].ClearAll()
		nodes[i].Close()
	}
}
