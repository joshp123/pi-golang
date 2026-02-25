package pi

import "github.com/joshp123/pi-golang/internal/runtime"

type eventQueue struct {
	queue *runtime.Queue[Event]
}

func newEventQueue() *eventQueue {
	return &eventQueue{queue: runtime.NewQueue[Event]()}
}

func (queue *eventQueue) push(event Event) bool {
	return queue.queue.Push(event)
}

func (queue *eventQueue) pop() (Event, bool) {
	return queue.queue.Pop()
}

func (queue *eventQueue) close() {
	queue.queue.Close()
}
