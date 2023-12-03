package microcache

import (
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash"
	"github.com/vmihailenco/msgpack"
)

type Config struct {
	MaxSize uint64
	Buckets int
	Ttl     time.Duration
}

type Cache struct {
	config   Config
	buckets  []*Bucket
	size     atomic.Uint64
	hits     atomic.Uint64
	misses   atomic.Uint64
	Overflow atomic.Uint64
}

func New(config Config) *Cache {
	if config.Buckets == 0 {
		config.Buckets = 16
	}
	if config.Ttl == 0 {
		config.Ttl = time.Second * 60 * 5
	}

	cache := &Cache{
		config:  config,
		buckets: make([]*Bucket, config.Buckets),
	}
	cache.size.Store(0)
	for i := 0; i < config.Buckets; i++ {
		cache.buckets[i] = NewBucket()
	}
	return cache
}

func (c *Cache) Get(key string, value any) bool {
	bucketId, hashKey := c.findBucket(key)
	bucket := c.getBucket(bucketId)
	item := bucket.Get(hashKey)
	if item == nil {
		c.misses.Add(1)
		return false
	}
	c.hits.Add(1)
	err := msgpack.Unmarshal(item.Value, value)
	if err != nil {
		panic(err)
	}
	return true
}

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

func (c *Cache) findBucket(key string) (uint64, uint64) {
	hash := xxhash.Sum64String(key)
	return hash % uint64(c.config.Buckets), hash
}

func (c *Cache) getBucket(key uint64) *Bucket {
	bucket := c.buckets[key]
	return bucket
}

func (c *Cache) Hits() uint64 {
	return c.hits.Load()
}

func (c *Cache) Misses() uint64 {
	return c.misses.Load()
}

func (c *Cache) Size() uint64 {
	return c.size.Load()
}

func (c *Cache) OverflowCount() uint64 {
	return c.Overflow.Load()
}
