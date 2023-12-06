package microcache

import (
	"sync"

	"github.com/emirpasic/gods/queues/arrayqueue"
)

// Bucket is a thread-safe map of items.
type Bucket struct {
	items map[uint64]*Item
	keys  *arrayqueue.Queue
	m     *sync.RWMutex
}

// NewBucket creates a new instance of Bucket.
// It initializes the bucket with an empty map of items,
// a read-write mutex, and an empty array queue for keys.
// Returns a pointer to the newly created Bucket.
func NewBucket() *Bucket {
	bucket := &Bucket{
		items: make(map[uint64]*Item),
		m:     &sync.RWMutex{},
		keys:  arrayqueue.New(),
	}
	return bucket
}

// Get retrieves the item with the specified key from the bucket.
// The key is a unique identifier for the item.
// Returns a pointer to the item if found, otherwise returns nil.

func (b *Bucket) Get(key uint64) *Item {
	b.m.RLock()
	defer b.m.RUnlock()
	item, ok := b.items[key]
	if !ok {
		return nil
	}
	return item
}

// Set adds or updates an item in the bucket with the specified key.
// It acquires a lock to ensure thread safety and then adds the item to the bucket's items map.
// The key is also enqueued in the bucket's keys queue.
func (b *Bucket) Set(key uint64, value *Item) {
	b.m.Lock()
	defer b.m.Unlock()
	b.items[key] = value
	b.keys.Enqueue(key)
}

// Size returns the number of items in the bucket.
func (b *Bucket) Size() uint64 {
	b.m.RLock()
	defer b.m.RUnlock()
	return uint64(len(b.items))
}

func (b *Bucket) Delete(key uint64) {
	b.m.Lock()
	defer b.m.Unlock()
	item := b.items[key]
	if item == nil {
		return
	}
	// mark the item as invalid, i will be deleted by the next clean
	item.Valide = false
}

// clean applies the time-to-live (TTL) logic to the items in the bucket.
// It removes expired items from the bucket and returns the number of removed items and their total size.
func (b *Bucket) clean() (nbRemoved uint64, removedSize uint64) {
	removed := uint64(0)
	b.m.Lock()
	defer b.m.Unlock()
	tmp := arrayqueue.New()
	for !b.keys.Empty() {
		key, ok := b.keys.Dequeue()
		if !ok {
			continue
		}
		item := b.items[key.(uint64)]
		if item == nil {
			continue
		}
		if item.Expired() || !item.Valide {
			removed += item.Size()
			delete(b.items, key.(uint64))
		} else {
			tmp.Enqueue(item)
		}
		b.keys = tmp
	}
	return uint64(len(b.items)), removed
}

// DeleteLast deletes the last item in the bucket
// and returns the size of the deleted item.
func (b *Bucket) DeleteLast() uint64 {
	b.m.Lock()
	defer b.m.Unlock()
	key, ok := b.keys.Dequeue()
	if !ok {
		return 0
	}
	item := b.items[key.(uint64)]
	if item == nil {
		return 0
	}
	return item.Size()
}
