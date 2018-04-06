package nano_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/aerogo/nano"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceGet(t *testing.T) {
	node := nano.New(port)
	defer node.Close()
	defer node.Clear()

	db := node.Namespace("test").RegisterTypes(types...)
	assert.True(t, node.IsServer())

	db.Set("User", "1", newUser(1))
	db.Set("User", "2", newUser(2))

	val, err := db.Get("User", "1")
	assert.NoError(t, err)

	user, ok := val.(*User)
	assert.True(t, ok)
	assert.Equal(t, "Test User", user.Name)

	assert.NotNil(t, val)
}

func TestNamespaceGetMany(t *testing.T) {
	node := nano.New(port)
	defer node.Close()
	defer node.Clear()

	db := node.Namespace("test").RegisterTypes(types...)
	assert.True(t, node.IsServer())

	db.Set("User", "1", newUser(1))
	db.Set("User", "2", newUser(2))

	objects := db.GetMany("User", []string{
		"1",
		"2",
	})

	assert.Len(t, objects, 2)

	for _, object := range objects {
		user, ok := object.(*User)
		assert.True(t, ok)
		assert.Equal(t, "Test User", user.Name)
	}
}

func TestNamespaceSet(t *testing.T) {
	node := nano.New(port)
	defer node.Close()
	defer node.Clear()

	db := node.Namespace("test").RegisterTypes(types...)
	assert.True(t, node.IsServer())

	db.Set("User", "1", newUser(1))
	db.Delete("User", "2")

	assert.True(t, db.Exists("User", "1"))
	assert.False(t, db.Exists("User", "2"))
}

func TestNamespaceClear(t *testing.T) {
	node := nano.New(port)
	defer node.Close()
	defer node.Clear()

	db := node.Namespace("test").RegisterTypes(types...)
	assert.True(t, node.IsServer())

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

func TestNamespaceAll(t *testing.T) {
	node := nano.New(port)
	db := node.Namespace("test").RegisterTypes(types...)
	assert.True(t, node.IsServer())
	defer node.Close()
	defer node.Clear()

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

func TestNamespaceClose(t *testing.T) {
	node := nano.New(port)

	assert.True(t, node.IsServer())
	assert.False(t, node.IsClosed())

	node.Close()

	assert.True(t, node.IsClosed())
}

func TestNamespaceTypes(t *testing.T) {
	node := nano.New(port)
	defer node.Close()

	db := node.Namespace("test").RegisterTypes(types...)
	assert.Equal(t, reflect.TypeOf(User{}), db.Types()["User"])
}

func TestNamespacePrefetch(t *testing.T) {
	node := nano.New(port)
	defer node.Close()

	db := node.Namespace("test").RegisterTypes(types...)
	db.Prefetch()
}

func TestNamespaceNode(t *testing.T) {
	node := nano.New(port)
	defer node.Close()

	db := node.Namespace("test").RegisterTypes(types...)
	assert.Equal(t, db.Node(), node)
}

// func TestNamespaceColdStart(t *testing.T) {
// 	time.Sleep(500 * time.Millisecond)
// 	db := nano.New(port).Namespace("test").RegisterTypes(types...)
// 	assert.True(t, node.IsServer())

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
// 	newDB := nano.New(port).Namespace("test").RegisterTypes(types...)
// 	assert.True(t, newnode.IsServer())

// 	defer newDB.Close()
// 	defer newDB.ClearAll()

// 	for i := 0; i < 10000; i++ {
// 		if !newDB.Exists("User", strconv.Itoa(i)) {
// 			assert.FailNow(t, fmt.Sprintf("User %d does not exist after cold start", i))
// 		}
// 	}
// }
