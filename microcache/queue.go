package microcache

import "sync"

type QueueItem[I any] struct {
	item I
	next *QueueItem[I]
}

type Queue[I any] struct {
	head  *QueueItem[I]
	tail  *QueueItem[I]
	size  int64
	mutex *sync.Mutex
}

func NewQueue[I any]() *Queue[I] {
	return &Queue[I]{
		mutex: &sync.Mutex{},
	}
}

func (q *Queue[I]) Push(item I) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.size++
	if q.head == nil {
		q.head = &QueueItem[I]{item: item}
		q.tail = q.head
		return
	}
	q.tail.next = &QueueItem[I]{item: item}
	q.tail = q.tail.next
}

func (q *Queue[I]) Pop() (value I) {
	if q.head == nil {
		return
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.size--
	value = q.head.item
	q.head = q.head.next
	return value
}

func (q *Queue[I]) Size() int64 {
	return q.size
}

func (q *Queue[I]) Empty() bool {
	return q.size == 0
}

func (q *Queue[I]) Peek() (value I) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if q.head == nil {
		return
	}
	return q.head.item
}

func (q *Queue[I]) Clear() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.head = nil
	q.tail = nil
	q.size = 0
}
