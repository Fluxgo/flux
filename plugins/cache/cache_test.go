package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/Fluxgo/flux/pkg/flux"
)

func TestCachePlugin(t *testing.T) {
	
	config := &Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		Prefix:   "test:",
	}

	
	app := &flux.Application{}

	
	cache, err := New(app, config)
	if err != nil {
		t.Fatalf("Failed to create cache plugin: %v", err)
	}
	defer cache.Shutdown()

	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		
		err := cache.Set(ctx, "string_key", "test_value", time.Hour)
		assert.NoError(t, err)

		var value string
		err = cache.Get(ctx, "string_key", &value)
		assert.NoError(t, err)
		assert.Equal(t, "test_value", value)

		
		type TestStruct struct {
			Name  string
			Value int
		}
		testStruct := TestStruct{Name: "test", Value: 42}
		err = cache.Set(ctx, "struct_key", testStruct, time.Hour)
		assert.NoError(t, err)

		var retrievedStruct TestStruct
		err = cache.Get(ctx, "struct_key", &retrievedStruct)
		assert.NoError(t, err)
		assert.Equal(t, testStruct, retrievedStruct)
	})

	t.Run("Delete", func(t *testing.T) {
		
		err := cache.Set(ctx, "delete_key", "test_value", time.Hour)
		assert.NoError(t, err)

		
		err = cache.Delete(ctx, "delete_key")
		assert.NoError(t, err)

		
		var value string
		err = cache.Get(ctx, "delete_key", &value)
		assert.Equal(t, flux.ErrNotFound, err)
	})

	t.Run("Exists", func(t *testing.T) {
		
		err := cache.Set(ctx, "exists_key", "test_value", time.Hour)
		assert.NoError(t, err)

		
		exists, err := cache.Exists(ctx, "exists_key")
		assert.NoError(t, err)
		assert.True(t, exists)

		
		exists, err = cache.Exists(ctx, "non_existent_key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Increment and Decrement", func(t *testing.T) {
		
		value, err := cache.Increment(ctx, "counter")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), value)

		value, err = cache.Increment(ctx, "counter")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), value)

		
		value, err = cache.Decrement(ctx, "counter")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), value)
	})

	t.Run("SetNX", func(t *testing.T) {
		
		success, err := cache.SetNX(ctx, "setnx_key", "test_value", time.Hour)
		assert.NoError(t, err)
		assert.True(t, success)

		
		success, err = cache.SetNX(ctx, "setnx_key", "new_value", time.Hour)
		assert.NoError(t, err)
		assert.False(t, success)

		
		var value string
		err = cache.Get(ctx, "setnx_key", &value)
		assert.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("GetOrSet", func(t *testing.T) {
		
		var value string
		err := cache.GetOrSet(ctx, "getorset_key", &value, time.Hour, func() (interface{}, error) {
			return "computed_value", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "computed_value", value)

		
		var cachedValue string
		err = cache.GetOrSet(ctx, "getorset_key", &cachedValue, time.Hour, func() (interface{}, error) {
			return "new_value", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "computed_value", cachedValue) // Should still be the original value
	})

	t.Run("Clear", func(t *testing.T) {
		
		err := cache.Set(ctx, "clear_key1", "value1", time.Hour)
		assert.NoError(t, err)
		err = cache.Set(ctx, "clear_key2", "value2", time.Hour)
		assert.NoError(t, err)

		
		err = cache.Clear(ctx)
		assert.NoError(t, err)

		
		exists, err := cache.Exists(ctx, "clear_key1")
		assert.NoError(t, err)
		assert.False(t, exists)

		exists, err = cache.Exists(ctx, "clear_key2")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
} 
