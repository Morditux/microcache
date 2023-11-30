package microcache

type QueueItem[I any] struct {
	item I
	next *QueueItem[I]
	prev *QueueItem[I]
}

type Queue[I any] struct {
	head *QueueItem[I]
	tail *QueueItem[I]
	size int64
}

func NewQueue[I any]() *Queue[I] {
	return &Queue[I]{}
}

func (q *Queue[I]) Push(item I) {
	q.size++
	if q.head == nil {
		q.head = &QueueItem[I]{item: item, next: nil, prev: nil}
		q.tail = q.head
		return
	}
	q.tail.next = &QueueItem[I]{item: item, next: nil, prev: q.tail}
	q.tail = q.tail.next
}

func (q *Queue[I]) Pop() (value I) {
	if q.head == nil {
		return
	}
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
	if q.head == nil {
		return
	}
	return q.head.item
}

func (q *Queue[I]) PopLeast() (value I) {
	if q.tail == nil {
		return
	}
	item := q.tail
	q.tail = item.prev
	return item.item
}

func (q *Queue[I]) Clear() {
	q.head = nil
	q.tail = nil
	q.size = 0
}
