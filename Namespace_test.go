package nano_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/aerogo/nano"
	"github.com/akyoto/assert"
)

func TestNamespaceGet(t *testing.T) {
	node := nano.New(config)
	defer node.Close()
	defer node.Clear()

	db := node.NewNamespace("test", types...)
	db.Set("User", "1", newUser(1))
	db.Set("User", "2", newUser(2))

	val, err := db.Get("User", "1")
	assert.Nil(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, val.(*User).ID, 1)

	val, err = db.Get("User", "2")
	assert.Nil(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, val.(*User).ID, 2)
}

func TestNamespaceGetMany(t *testing.T) {
	node := nano.New(config)
	defer node.Close()
	defer node.Clear()

	db := node.NewNamespace("test", types...)
	db.Set("User", "1", newUser(1))
	db.Set("User", "2", newUser(2))

	objects := db.GetMany("User", []string{
		"1",
		"2",
	})

	assert.Equal(t, len(objects), 2)

	for i, object := range objects {
		user, ok := object.(*User)
		assert.True(t, ok)
		assert.Equal(t, user.ID, i+1)
	}
}

func TestNamespaceSet(t *testing.T) {
	node := nano.New(config)
	defer node.Close()
	defer node.Clear()

	db := node.NewNamespace("test", types...)
	db.Set("User", "1", newUser(1))
	db.Delete("User", "2")

	assert.True(t, db.Exists("User", "1"))
	assert.False(t, db.Exists("User", "2"))
}

func TestNamespaceClear(t *testing.T) {
	node := nano.New(config)
	defer node.Close()
	defer node.Clear()

	db := node.NewNamespace("test", types...)

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
	node := nano.New(config)
	db := node.NewNamespace("test", types...)
	defer node.Close()
	defer node.Clear()

	assert.Equal(t, db.Count("User"), uint64(0))
	newCount := 10000

	for i := 0; i < newCount; i++ {
		db.Set("User", strconv.Itoa(i), newUser(i))
	}

	count := 0

	for user := range db.All("User") {
		assert.NotNil(t, user)
		count++
	}

	assert.Equal(t, newCount, count)
	assert.Equal(t, db.Count("User"), uint64(count))
}

func TestNamespaceClose(t *testing.T) {
	node := nano.New(config)
	node.Close()
}

func TestNamespaceTypes(t *testing.T) {
	node := nano.New(config)
	defer node.Close()

	db := node.NewNamespace("test", types...)
	assert.Equal(t, reflect.TypeOf(User{}), db.Types()["User"])
}

func TestNamespacePrefetch(t *testing.T) {
	node := nano.New(config)
	defer node.Close()

	db := node.NewNamespace("test", types...)
	db.Prefetch()
}

func TestNamespaceNode(t *testing.T) {
	node := nano.New(config)
	defer node.Close()

	db := node.NewNamespace("test", types...)
	assert.Equal(t, db.Node(), node)
}
