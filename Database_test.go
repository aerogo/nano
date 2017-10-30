package nano_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/aerogo/nano"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseGet(t *testing.T) {
	db := nano.New("test", types)
	assert.True(t, db.Node().IsServer())
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
	assert.True(t, db.Node().IsServer())
	defer db.Close()
	defer db.ClearAll()

	db.Set("User", "1", newUser(1))
	db.Delete("User", "2")

	assert.True(t, db.Exists("User", "1"))
	assert.False(t, db.Exists("User", "2"))
}

func TestDatabaseClear(t *testing.T) {
	db := nano.New("test", types)
	assert.True(t, db.Node().IsServer())
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
	assert.True(t, db.Node().IsServer())
	defer db.Close()
	defer db.ClearAll()

	db.Collection("User").Clear()
	recordCount := 10000

	for i := 0; i < recordCount; i++ {
		db.Set("User", strconv.Itoa(i), newUser(i))
	}

	count := 0

	for user := range db.All("User") {
		assert.NotNil(t, user)
		count++
	}

	assert.Equal(t, recordCount, count)
}

func TestDatabaseClose(t *testing.T) {
	db := nano.New("test", types)
	assert.True(t, db.Node().IsServer())
	assert.False(t, db.Node().IsClosed())

	db.Close()

	time.Sleep(100 * time.Millisecond)
	assert.True(t, db.Node().IsClosed())
}

// func TestDatabaseColdStart(t *testing.T) {
// 	time.Sleep(500 * time.Millisecond)
// 	db := nano.New("test", types)
// 	assert.True(t, db.Node().IsServer())

// 	for i := 0; i < 10000; i++ {
// 		db.Set("User", strconv.Itoa(i), newUser(i))
// 		assert.True(t, db.Exists("User", strconv.Itoa(i)))
// 	}

// 	db.Close()

// 	// Sync filesystem
// 	exec.Command("sync").Run()

// 	// Wait a little
// 	time.Sleep(2000 * time.Millisecond)

// 	// Cold start
// 	newDB := nano.New("test", types)
// 	assert.True(t, newdb.Node().IsServer())

// 	defer newDB.Close()
// 	defer newDB.ClearAll()

// 	for i := 0; i < 10000; i++ {
// 		if !newDB.Exists("User", strconv.Itoa(i)) {
// 			assert.FailNow(t, fmt.Sprintf("User %d does not exist after cold start", i))
// 		}
// 	}
// }
