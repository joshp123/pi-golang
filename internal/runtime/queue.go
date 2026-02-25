package runtime

import "sync"

// Queue is a blocking FIFO with close semantics.
type Queue[T any] struct {
	mu     sync.Mutex
	cond   *sync.Cond
	items  []T
	head   int
	closed bool
}

func NewQueue[T any]() *Queue[T] {
	queue := &Queue[T]{}
	queue.cond = sync.NewCond(&queue.mu)
	return queue
}

func (queue *Queue[T]) Push(item T) bool {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if queue.closed {
		return false
	}
	queue.items = append(queue.items, item)
	queue.cond.Signal()
	return true
}

func (queue *Queue[T]) Pop() (T, bool) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	for !queue.closed && queue.head >= len(queue.items) {
		queue.cond.Wait()
	}

	if queue.head >= len(queue.items) {
		var zero T
		return zero, false
	}

	item := queue.items[queue.head]
	queue.head++
	queue.compact()
	return item, true
}

func (queue *Queue[T]) Close() {
	queue.mu.Lock()
	queue.closed = true
	queue.cond.Broadcast()
	queue.mu.Unlock()
}

func (queue *Queue[T]) compact() {
	if queue.head == 0 {
		return
	}
	if queue.head < 1024 && queue.head*2 < len(queue.items) {
		return
	}
	remaining := len(queue.items) - queue.head
	copy(queue.items[:remaining], queue.items[queue.head:])
	queue.items = queue.items[:remaining]
	queue.head = 0
}
