package security

import (
	"sync"
	"sync/atomic"
)

// defaultReplaySize bounds the in-memory buffer used to answer a reconnect's
// Last-Event-ID. Beyond it the client re-syncs from the REST history.
const defaultReplaySize = 500

// Subscriber is one connected administrator's view of the stream.
type Subscriber struct {
	// Events closes when the broker shuts down or the subscriber unsubscribes.
	Events <-chan SecurityEvent
}

type brokerSubscriber struct {
	ch chan SecurityEvent
}

// Broker fans security events out to every connected admin.
//
// Publish is non-blocking: a subscriber whose bounded queue is full simply drops
// that event (and the drop is counted), so one slow admin can never stall event
// collection or another admin's stream.
type Broker struct {
	mu       sync.Mutex
	subs     map[*brokerSubscriber]struct{}
	ring     []SecurityEvent
	ringSize int
	closed   bool
	dropped  atomic.Int64
}

// NewBroker builds a broker with a bounded replay window.
func NewBroker(replaySize int) *Broker {
	if replaySize <= 0 {
		replaySize = defaultReplaySize
	}
	return &Broker{
		subs:     make(map[*brokerSubscriber]struct{}),
		ringSize: replaySize,
	}
}

// Publish delivers ev to every subscriber without blocking.
func (b *Broker) Publish(ev SecurityEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}

	if ev.Persisted && b.ringSize > 0 {
		if len(b.ring) == b.ringSize {
			copy(b.ring, b.ring[1:])
			b.ring = b.ring[:len(b.ring)-1]
		}
		b.ring = append(b.ring, ev)
	}

	for sub := range b.subs {
		select {
		case sub.ch <- ev:
		default:
			// Bounded queue full: drop for this subscriber only.
			b.dropped.Add(1)
		}
	}
}

// Subscribe registers a subscriber whose queue holds at most buffer events.
func (b *Broker) Subscribe(buffer int) *Subscriber {
	if buffer <= 0 {
		buffer = 64
	}
	sub := &brokerSubscriber{ch: make(chan SecurityEvent, buffer)}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		close(sub.ch)
		return &Subscriber{Events: sub.ch}
	}
	b.subs[sub] = struct{}{}
	return &Subscriber{Events: sub.ch}
}

// Unsubscribe removes a subscriber and closes its channel. Safe to call twice.
func (b *Broker) Unsubscribe(s *Subscriber) {
	if s == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for sub := range b.subs {
		if sub.ch == s.Events {
			delete(b.subs, sub)
			close(sub.ch)
			return
		}
	}
}

// Replay returns the events published after the one identified by afterID.
// It returns nil when that id is outside the replay window, which tells the
// caller to fall back to the REST history rather than pretend to be complete.
func (b *Broker) Replay(afterID string) []SecurityEvent {
	if afterID == "" {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, ev := range b.ring {
		if ev.EventID == afterID {
			out := make([]SecurityEvent, len(b.ring)-i-1)
			copy(out, b.ring[i+1:])
			return out
		}
	}
	return nil
}

// Close disconnects every subscriber and stops accepting events.
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for sub := range b.subs {
		close(sub.ch)
		delete(b.subs, sub)
	}
}

// Dropped is the number of per-subscriber deliveries dropped because a queue
// was full. Exposed so monitoring degradation is visible, never silent.
func (b *Broker) Dropped() int64 { return b.dropped.Load() }

// Subscribers is the current number of connected streams (used by tests).
func (b *Broker) Subscribers() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.subs)
}
