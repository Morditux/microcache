package microcache_test

import (
	"fmt"
	"sync"
	"testing"

	cache "github.com/Morditux/microcache/microcache"
	"github.com/google/uuid"
)

func TestNewCache(tests *testing.T) {
	tests.Log("TestNewCache")
	config := cache.Config{
		MaxSize: 1024,
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
		MaxSize: 1024,
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
		MaxSize: 1024,
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

func BenchmarkCacheSet(b *testing.B) {
	config := cache.Config{
		MaxSize: 256 * 1024 * 1024,
		Buckets: 128,
	}
	cache := cache.New(config)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			value := uuid.New().String()
			key := uuid.New().String()
			cache.Set(key, &value)
			ret := ""
			cache.Get(key, &ret)
			if ret != value {
				b.Error("Cache returned wrong value " + ret + " for key test")
			}
		}
	})
}

func BenchmarkCacheGet(b *testing.B) {
	config := cache.Config{
		MaxSize: 256 * 1024 * 1024,
		Buckets: 128,
	}
	cache := cache.New(config)
	for i := 0; i < b.N; i++ {
		wg := sync.WaitGroup{}
		wg.Add(1000)
		for j := 0; j < 1000; j++ {
			go func() {
				value := uuid.New().String()
				key := uuid.New().String()
				cache.Set(key, &value)
				ret := ""
				cache.Get(key, &ret)
				if ret != value {
					b.Error("Cache returned wrong value " + ret + " for key test")
				}
				wg.Done()
			}()
		} //for j
		wg.Wait()
	}
}

func TestCacheSize(t *testing.T) {
	config := cache.Config{
		MaxSize: 1024 * 64,
		Buckets: 16,
	}
	cache := cache.New(config)
	if cache == nil {
		t.Error("NewCache returned nil")
	}
	for i := 0; i < 15000; i++ {
		value := uuid.New().String()
		key := uuid.New().String()
		cache.Set(key, &value)
	}
	fmt.Println("Cache size: ", cache.Size())
	fmt.Println("Cache overflow : ", cache.OverflowCount())
}
