package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage(t *testing.T) {
	storage := NewInMemoryStorage()

	t.Run("Set and Get", func(t *testing.T) {
		// Test setting and getting a value
		storage.Set("key1", "value1")
		value, ok := storage.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, "value1", value)

		// Test getting non-existent key
		value, ok = storage.Get("nonexistent")
		assert.False(t, ok)
		assert.Nil(t, value)
	})

	t.Run("Delete", func(t *testing.T) {
		// Test deleting a value
		storage.Set("key2", "value2")
		storage.Delete("key2")
		value, ok := storage.Get("key2")
		assert.False(t, ok)
		assert.Nil(t, value)
	})

	t.Run("Clear", func(t *testing.T) {
		// Test clearing all values
		storage.Set("key3", "value3")
		storage.Set("key4", "value4")
		storage.Clear()
		value, ok := storage.Get("key3")
		assert.False(t, ok)
		assert.Nil(t, value)
		value, ok = storage.Get("key4")
		assert.False(t, ok)
		assert.Nil(t, value)
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		// Test concurrent access to storage
		done := make(chan bool)
		go func() {
			for i := 0; i < 1000; i++ {
				storage.Set("concurrent", i)
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 1000; i++ {
				storage.Get("concurrent")
			}
			done <- true
		}()

		<-done
		<-done
		// No panic means the test passed
	})

	t.Run("Different Types", func(t *testing.T) {
		// Test storing different types
		storage.Set("string", "value")
		storage.Set("int", 42)
		storage.Set("float", 3.14)
		storage.Set("bool", true)
		storage.Set("struct", struct{}{})

		value, ok := storage.Get("string")
		assert.True(t, ok)
		assert.Equal(t, "value", value)

		value, ok = storage.Get("int")
		assert.True(t, ok)
		assert.Equal(t, 42, value)

		value, ok = storage.Get("float")
		assert.True(t, ok)
		assert.Equal(t, 3.14, value)

		value, ok = storage.Get("bool")
		assert.True(t, ok)
		assert.Equal(t, true, value)

		value, ok = storage.Get("struct")
		assert.True(t, ok)
		assert.Equal(t, struct{}{}, value)
	})
}
