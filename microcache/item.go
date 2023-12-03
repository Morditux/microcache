package microcache

import "time"

type Item struct {
	// Key represents the unique identifier of the item.
	Key string
	// Value contains the data associated with the item.
	Value []byte
	// Ttl represents the time-to-live duration of the item.
	Ttl time.Duration
	// CreateAt stores the timestamp when the item was created.
	CreateAt time.Time
}

// NewItem crée un nouvel élément de cache avec la clé, la valeur et la durée de vie spécifiées.
// La fonction renvoie un pointeur vers l'élément créé.
func NewItem(key string, value []byte, timeToLive time.Duration) *Item {
	return &Item{
		Key:      key,
		Value:    value,
		Ttl:      timeToLive,
		CreateAt: time.Now(),
	}
}

// Expired returns a boolean value indicating whether the item has expired or not.
// An item is considered expired if the current time is greater than the time when it was created plus its time-to-live (TTL) duration.
func (i *Item) Expired() bool {
	return time.Since(i.CreateAt) > i.Ttl
}

// Size returns the size of the item in bytes.
// It calculates the size by summing the lengths of the key and value.
func (i *Item) Size() uint64 {
	return uint64(len(i.Key) + len(i.Value))
}
