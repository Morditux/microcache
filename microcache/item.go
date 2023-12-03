package microcache

import "time"

type Item struct {
	Key      string
	Value    []byte
	Ttl      time.Duration
	CreateAt time.Time
}

func NewItem(key string, value []byte, timeToLive time.Duration) *Item {
	return &Item{
		Key:      key,
		Value:    value,
		Ttl:      timeToLive,
		CreateAt: time.Now(),
	}
}

func (i *Item) Expired() bool {
	return time.Since(i.CreateAt) > i.Ttl
}

func (i *Item) Size() uint64 {
	return uint64(len(i.Key) + len(i.Value))
}
