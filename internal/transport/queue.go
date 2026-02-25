package transport

import "github.com/joshp123/pi-golang/internal/runtime"

// Queue is a typed wrapper around runtime.Queue.
type Queue[T any] struct {
	queue *runtime.Queue[T]
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{queue: runtime.NewQueue[T]()}
}

func (queue *Queue[T]) Push(event T) bool {
	return queue.queue.Push(event)
}

func (queue *Queue[T]) Pop() (T, bool) {
	return queue.queue.Pop()
}

func (queue *Queue[T]) Close() {
	queue.queue.Close()
}
