package microcache

import (
	"errors"
	"sync"
)

type Bucket struct {
	items map[uint64]*Item
	keys  *Queue[uint64]
	m     *sync.RWMutex
	size  uint64
}

func NewBucket() *Bucket {
	return &Bucket{
		items: make(map[uint64]*Item),
		m:     &sync.RWMutex{},
		keys:  NewQueue[uint64](),
		size:  0,
	}
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
	b.keys.Push(key)
}

func (b *Bucket) Delete(key uint64) uint64 {
	b.m.Lock()
	defer b.m.Unlock()
	item := b.items[key]
	if item == nil {
		return 0
	}
	b.size -= item.Size()
	delete(b.items, key)
	b.keys.Delete(key)
	return item.Size()
}

func (b *Bucket) Size() uint64 {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.size
}

func (b *Bucket) DeleteFirst() uint64 {
	b.m.Lock()
	defer b.m.Unlock()
	key := b.keys.Pop()
	item := b.items[key]
	if item == nil {
		return 0
	}
	delete(b.items, key)
	return item.Size()
}

func (b *Bucket) DeleteLast() uint64 {
	b.m.Lock()
	defer b.m.Unlock()
	key, err := b.keys.PopLeast()
	if err != nil {
		if errors.Is(err, ErrEmptyQueueError) {
			return 0
		}
	}
	item := b.items[key]
	if item == nil {
		return 0
	}
	delete(b.items, key)
	return item.Size()
}
