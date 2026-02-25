package pi

import "sync"

type eventQueue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	items  []Event
	head   int
	closed bool
}

func newEventQueue() *eventQueue {
	queue := &eventQueue{}
	queue.cond = sync.NewCond(&queue.mu)
	return queue
}

func (queue *eventQueue) push(event Event) bool {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if queue.closed {
		return false
	}
	queue.items = append(queue.items, event)
	queue.cond.Signal()
	return true
}

func (queue *eventQueue) pop() (Event, bool) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	for !queue.closed && queue.head >= len(queue.items) {
		queue.cond.Wait()
	}

	if queue.head >= len(queue.items) {
		return Event{}, false
	}

	event := queue.items[queue.head]
	queue.head++
	queue.compact()
	return event, true
}

func (queue *eventQueue) close() {
	queue.mu.Lock()
	queue.closed = true
	queue.cond.Broadcast()
	queue.mu.Unlock()
}

func (queue *eventQueue) compact() {
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
