package microcache_test

import (
	"testing"

	cache "github.com/Morditux/microcache/microcache"
)

func TestNewCache(tests *testing.T) {
	tests.Log("TestNewCache")
	config := cache.Config{
		MaxSize: 1024 * 1024,
		Buckets: 16,
	}
	cache := cache.New(config)
	if cache == nil {
		tests.Error("NewCache returned nil")
	}
}

func TestCacheGet(tests *testing.T) {
	tests.Log("TestCacheGet")
	config := cache.Config{
		MaxSize: 1024 * 1024,
		Buckets: 16,
	}
	cache := cache.New(config)
	if cache == nil {
		tests.Error("NewCache returned nil")
	}
	value := ""
	foud := cache.Get("test", &value)
	if foud {
		tests.Error("Cache returned value for non existing key")
	}
}

func TestCacheSet(tests *testing.T) {
	tests.Log("TestCacheSet")
	config := cache.Config{
		MaxSize: 1024 * 1024,
		Buckets: 16,
	}
	cache := cache.New(config)
	if cache == nil {
		tests.Error("NewCache returned nil")
	}
	value := "Test value"
	cache.Set("test", &value)
	ret := ""
	found := cache.Get("test", &ret)
	if !found {
		tests.Error("Cache did not return value for existing key")
	}
	if ret != value {
		tests.Error("Cache returned wrong value")
	}
}
