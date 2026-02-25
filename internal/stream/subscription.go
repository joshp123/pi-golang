package stream

import "sync"

type subscription[T any] struct {
	out    chan T
	in     chan T
	done   chan struct{}
	once   sync.Once
	policy Policy
}

func newSubscription[T any](policy Policy) *subscription[T] {
	return &subscription[T]{
		out:    make(chan T, policy.Buffer),
		in:     make(chan T, policy.Buffer),
		done:   make(chan struct{}),
		policy: policy,
	}
}

func (sub *subscription[T]) run() {
	defer close(sub.out)
	for {
		select {
		case event := <-sub.in:
			select {
			case sub.out <- event:
			case <-sub.done:
				select {
				case sub.out <- event:
				default:
				}
				sub.drainBestEffort()
				return
			}
		case <-sub.done:
			sub.drainBestEffort()
			return
		}
	}
}

func (sub *subscription[T]) drainBestEffort() {
	for {
		select {
		case event := <-sub.in:
			select {
			case sub.out <- event:
			default:
			}
		default:
			return
		}
	}
}

func (sub *subscription[T]) close() {
	sub.once.Do(func() {
		close(sub.done)
	})
}

func (sub *subscription[T]) enqueue(event T) (dropped bool) {
	switch sub.policy.Mode {
	case ModeBlock:
		select {
		case <-sub.done:
			return false
		case sub.in <- event:
			return false
		}
	case ModeRing:
		for {
			select {
			case <-sub.done:
				return dropped
			case sub.in <- event:
				return dropped
			default:
			}
			select {
			case <-sub.done:
				return dropped
			case <-sub.in:
				dropped = true
			default:
			}
		}
	default:
		select {
		case <-sub.done:
			return false
		case sub.in <- event:
			return false
		default:
			return true
		}
	}
}

func (sub *subscription[T]) enqueueSystem(event T) {
	for {
		select {
		case <-sub.done:
			return
		case sub.in <- event:
			return
		default:
		}
		select {
		case <-sub.done:
			return
		case <-sub.in:
		default:
		}
	}
}
