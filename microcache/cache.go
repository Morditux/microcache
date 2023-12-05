package microcache

import (
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash"
	"github.com/vmihailenco/msgpack"
)

type Config struct {
	MaxSize  uint64        // MaxSize is the maximum size of the cache in bytes.
	Buckets  int           // Buckets is the number of buckets used for sharding the cache.
	Ttl      time.Duration // Ttl is the time-to-live duration for cache entries.
	Eviction time.Duration // Eviction is the duration after which cache entries are evicted.
}

// Cache represents a cache implementation.
type Cache struct {
	config   Config        // The configuration of the cache.
	buckets  []*Bucket     // The buckets used for storing cache entries.
	size     atomic.Uint64 // The current size of the cache.
	hits     atomic.Uint64 // The number of cache hits.
	misses   atomic.Uint64 // The number of cache misses.
	Overflow atomic.Uint64 // The number of cache overflows.
}

// New creates a new instance of Cache with the provided configuration.
// If the number of buckets is not specified in the configuration, it defaults to 16.
// If the time-to-live (TTL) is not specified in the configuration, it defaults to 5 minutes.
// The Cache starts with an initial size of 0 and initializes the buckets accordingly.
// If the eviction duration is specified in the configuration, a TTL evictor goroutine is started to periodically remove expired items from the cache.
// The size of the cache is updated accordingly after each eviction.
// Returns a pointer to the newly created Cache instance.
func New(config Config) *Cache {
	if config.Buckets == 0 {
		config.Buckets = 16
	}
	if config.Eviction == 0 {
		config.Eviction = 5 * time.Minute
	}
	cache := &Cache{
		config:  config,
		buckets: make([]*Bucket, config.Buckets),
	}
	cache.size.Store(0)
	for i := 0; i < config.Buckets; i++ {
		cache.buckets[i] = NewBucket()
	}
	// Start TTL evictor
	tickerChan := time.NewTicker(config.Eviction).C
	go func() {
		for range tickerChan {
			time.Sleep(config.Eviction)
			for _, bucket := range cache.buckets {
				_, removedSize := bucket.clean()
				cache.size.Add(-removedSize)
			}
		}
	}()
	return cache
}

// Get retrieves the value associated with the given key from the cache.
// It returns true if the key was found and the value was successfully retrieved,
// and false otherwise. The retrieved value is stored in the 'value' parameter.
// If the value is expired or cannot be unmarshaled into the 'value' parameter,
// it returns false.
func (c *Cache) Get(key string, value any) bool {
	bucketId, hashKey := c.findBucket(key)
	bucket := c.getBucket(bucketId)
	item := bucket.Get(hashKey)
	if item == nil {
		c.misses.Add(1)
		return false
	}

	if item.Expired() {
		c.misses.Add(1)
		return false
	}
	if !item.Valide {
		c.misses.Add(1)
		return false
	}

	c.hits.Add(1)
	if c.config.Eviction != 0 {
		item.CreateAt = time.Now() // Update last access
	}
	err := msgpack.Unmarshal(item.Value, value)
	if err != nil {
		panic(err)
	}
	return true
}

// Set adds a new key-value pair to the cache.
// The key is a string that uniquely identifies the value.
// The value can be of any type and will be serialized using MessagePack before being stored in the cache.
// If the value cannot be serialized, a panic will occur.
// If the cache is full and adding the new item would exceed the maximum size,
// the cache will evict items until there is enough space.
// If an item is too large to fit in any bucket, it will not be added to the cache.
func (c *Cache) Set(key string, value any) {
	data, err := msgpack.Marshal(value)
	if err != nil {
		panic(err)
	}
	item := NewItem(key, data, c.config.Ttl)
	size := item.Size()
	emptyBucket := 0
	if c.size.Load()+size > c.config.MaxSize {
		c.Overflow.Add(1)
		for c.size.Load()+size > c.config.MaxSize {
			if emptyBucket == c.config.Buckets {
				// Item to large for cache
				return
			}
			for _, bucket := range c.buckets {
				if bucket.Size() == 0 {
					emptyBucket++
					continue
				}
				c.size.Add(-bucket.DeleteLast())
			}
		}
	}
	b, keyHash := c.findBucket(key)
	bucket := c.getBucket(b)
	bucket.Set(keyHash, item)
	c.size.Add(size)
}

func (c Cache) Delete(key string) {
	b, keyHash := c.findBucket(key)
	bucket := c.getBucket(b)
	bucket.Delete(keyHash)
}

// findBucket is a method that calculates the bucket index and hash value for a given key.
// It uses the xxhash algorithm to generate a hash value for the key and then calculates the bucket index
// by taking the modulus of the hash value with the number of buckets in the cache.
// The method returns the bucket index and the hash value.
func (c *Cache) findBucket(key string) (uint64, uint64) {
	hash := xxhash.Sum64String(key)
	return hash % uint64(c.config.Buckets), hash
}

// getBucket returns the bucket associated with the given key.
// If the bucket does not exist, it returns nil.
func (c *Cache) getBucket(key uint64) *Bucket {
	bucket := c.buckets[key]
	return bucket
}

// Hits returns the number of cache hits.
func (c *Cache) Hits() uint64 {
	return c.hits.Load()
}

// Misses returns the number of cache misses.
func (c *Cache) Misses() uint64 {
	return c.misses.Load()
}

// Size returns the current size of the cache in bytes.
func (c *Cache) Size() uint64 {
	return c.size.Load()
}

// OverflowCount returns the current value of the overflow count in the cache.
func (c *Cache) OverflowCount() uint64 {
	return c.Overflow.Load()
}
