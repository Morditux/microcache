package microcache

import (
	"sync"
	"time"

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
	bucket.start()
	return bucket
}

func (b *Bucket) start() {
	ticker := time.NewTicker(time.Second * 60)
	go func() {
		for {
			<-ticker.C
			now := time.Now()
			b.m.Lock()
			tmp := arrayqueue.New()
			for !b.keys.Empty() {
				key, _ := b.keys.Dequeue()
				item := b.items[key.(uint64)]
				if item == nil {
					continue
				}
				if now.Sub(item.CreateAt) < item.Ttl {
					tmp.Enqueue(item)
				} else {
					delete(b.items, key.(uint64))
				}
				b.keys = tmp
			}
			b.m.Unlock()
			ticker.Reset(time.Second * 60)
		}
	}()
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
