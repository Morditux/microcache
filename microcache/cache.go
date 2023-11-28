package microcache

import (
	"sync"

	"github.com/cespare/xxhash"
)

type Config struct {
	MaxSize int64
	Buckets int
}

type Bucket struct {
	items map[string]*Item
	m     *sync.RWMutex
	size  int64
}

type Cache struct {
	config  Config
	buckets map[uint64]*Bucket
	keys    *Queue[uint64]
	size    int64
}

func New(config Config) *Cache {
	if config.Buckets == 0 {
		config.Buckets = 16
	}
	return &Cache{
		config:  config,
		buckets: make(map[uint64]*Bucket),
		keys:    NewQueue[uint64](),
		size:    0,
	}
}

func newBucket() *Bucket {
	return &Bucket{
		items: make(map[string]*Item),
		m:     &sync.RWMutex{},
		size:  0,
	}
}

func (b *Bucket) Get(key string) *Item {
	b.m.RLock()
	defer b.m.RUnlock()
	item, ok := b.items[key]
	if !ok {
		return nil
	}
	return item
}

func (b *Bucket) Set(key string, value *Item) {
	b.m.Lock()
	defer b.m.Unlock()
	b.items[key] = value
}

func (b *Bucket) Delete(key string) {
	b.m.Lock()
	defer b.m.Unlock()
	delete(b.items, key)
}

func (c *Cache) Get(key string) *Item {
	bucket := c.getBucket(c.findBucket(key))
	return bucket.Get(key)
}

func (c *Cache) Set(key string, value *Item) {
	b := c.findBucket(key)
	bucket := c.getBucket(b)
	bucket.Set(key, value)
	c.keys.Push(b)
}

func (c *Cache) Delete(key string) {
	bucket := c.getBucket(c.findBucket(key))
	bucket.Delete(key)
}

func (c *Cache) findBucket(key string) uint64 {
	hash := xxhash.Sum64String(key)
	return hash % uint64(c.config.Buckets)
}

func (c *Cache) getBucket(key uint64) *Bucket {
	bucket, ok := c.buckets[key]
	if !ok {
		bucket = newBucket()
		c.buckets[key] = bucket
	}
	return bucket
}
