package pi

import (
	"encoding/json"
	"fmt"
	"sync"
)

type subscription struct {
	out    chan Event
	in     chan Event
	done   chan struct{}
	once   sync.Once
	policy SubscriptionPolicy
}

func newSubscription(policy SubscriptionPolicy) (*subscription, error) {
	if err := validateSubscriptionPolicy(policy); err != nil {
		return nil, err
	}
	return &subscription{
		out:    make(chan Event, policy.Buffer),
		in:     make(chan Event, policy.Buffer),
		done:   make(chan struct{}),
		policy: policy,
	}, nil
}

func validateSubscriptionPolicy(policy SubscriptionPolicy) error {
	if policy.Buffer <= 0 {
		return fmt.Errorf("%w: buffer must be > 0", ErrInvalidSubscriptionPolicy)
	}
	switch policy.Mode {
	case SubscriptionModeDrop, SubscriptionModeBlock, SubscriptionModeRing:
		return nil
	default:
		return fmt.Errorf("%w: unsupported mode %q", ErrInvalidSubscriptionPolicy, policy.Mode)
	}
}

func (sub *subscription) run() {
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

func (sub *subscription) drainBestEffort() {
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

func (sub *subscription) close() {
	sub.once.Do(func() {
		close(sub.done)
	})
}

func (sub *subscription) enqueue(event Event) (dropped bool) {
	switch sub.policy.Mode {
	case SubscriptionModeBlock:
		select {
		case <-sub.done:
			return false
		case sub.in <- event:
			return false
		}
	case SubscriptionModeRing:
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

func (sub *subscription) enqueueSystem(event Event) {
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

func newSubscriptionDropEvent(mode SubscriptionMode, droppedType string) Event {
	raw, _ := json.Marshal(map[string]any{
		"type":        EventTypeSubscriptionDrop,
		"mode":        mode,
		"droppedType": droppedType,
	})
	return Event{Type: EventTypeSubscriptionDrop, Raw: raw}
}
