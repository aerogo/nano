package database_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aerogo/database"
)

func TestDatabaseGet(t *testing.T) {
	db := database.New()
	db.Set("User", "123", &User{})
	val := db.Get("User", "123")

	assert.NotNil(t, val)
}

func TestDatabaseAll(t *testing.T) {
	db := database.New()

	for i := 0; i < 1000000; i++ {
		db.Set("User", strconv.Itoa(i), &User{})
	}

	count := 0
	start := time.Now()

	for user := range db.All("User") {
		assert.NotNil(t, user)
		count++
	}

	fmt.Println(time.Since(start))

	assert.Equal(t, 1000000, count)
}
