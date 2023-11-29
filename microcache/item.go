package microcache

type Item struct {
	Key   string
	Value []byte
}

func NewItem(key string, value []byte, timeToLive int64) *Item {
	return &Item{
		Key:   key,
		Value: value,
	}
}

func (i *Item) Size() uint64 {
	return uint64(len(i.Key) + len(i.Value))
}
