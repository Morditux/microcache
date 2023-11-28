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

func (i *Item) Size() int {
	return len(i.Key) + len(i.Value)
}
