package microcache

import (
	"sync"

	"github.com/cespare/xxhash"
	"github.com/vmihailenco/msgpack"
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
	buckets []*Bucket
	keys    *Queue[uint64]
	size    int64
	mutex   *sync.Mutex
}

func New(config Config) *Cache {
	if config.Buckets == 0 {
		config.Buckets = 16
	}
	cache := &Cache{
		config:  config,
		buckets: make([]*Bucket, config.Buckets),
		keys:    NewQueue[uint64](),
		size:    0,
		mutex:   &sync.Mutex{},
	}
	for i := 0; i < config.Buckets; i++ {
		cache.buckets[i] = newBucket()
	}
	return cache
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

func (c *Cache) Get(key string, value any) bool {
	bucket := c.getBucket(c.findBucket(key))
	item := bucket.Get(key)
	if item == nil {
		return false
	}
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
	item := &Item{
		Key:   key,
		Value: data,
	}
	b := c.findBucket(key)
	bucket := c.getBucket(b)
	bucket.Set(key, item)
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
	bucket := c.buckets[key]
	return bucket
}
