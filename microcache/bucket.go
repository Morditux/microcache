package microcache

import (
	"sync"

	"github.com/emirpasic/gods/queues/arrayqueue"
)

type Bucket struct {
	items map[uint64]*Item
	keys  *arrayqueue.Queue
	m     *sync.RWMutex
}

func NewBucket() *Bucket {
	bucket := &Bucket{
		items: make(map[uint64]*Item),
		m:     &sync.RWMutex{},
		keys:  arrayqueue.New(),
	}
	return bucket
}

func (b *Bucket) Get(key uint64) *Item {
	b.m.RLock()
	defer b.m.RUnlock()
	item, ok := b.items[key]
	if !ok {
		return nil
	}
	return item
}

func (b *Bucket) Set(key uint64, value *Item) {
	b.m.Lock()
	defer b.m.Unlock()
	b.items[key] = value
	b.keys.Enqueue(key)
}

func (b *Bucket) Size() uint64 {
	b.m.RLock()
	defer b.m.RUnlock()
	return uint64(len(b.items))
}

func (b *Bucket) applyTTL() uint64 {
	b.m.Lock()
	defer b.m.Unlock()
	tmp := arrayqueue.New()
	for !b.keys.Empty() {
		key, _ := b.keys.Dequeue()
		item := b.items[key.(uint64)]
		if item == nil {
			continue
		}
		if item.Expired() {
			delete(b.items, key.(uint64))
		} else {
			tmp.Enqueue(item)
		}
		b.keys = tmp
	}
	return uint64(len(b.items))
}

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
