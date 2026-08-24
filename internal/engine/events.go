package engine

import (
	"sync"
	"time"
)

type Event struct {
	RunID  string    `json:"run_id"`
	Kind   string    `json:"kind"`
	Detail string    `json:"detail,omitempty"`
	At     time.Time `json:"at"`
}
type Observer func(Event)
type ObserverHub struct {
	mu        sync.RWMutex
	next      int
	listeners map[int]Observer
	history   []Event
	limit     int
}

func NewObserverHub() *ObserverHub { return &ObserverHub{listeners: map[int]Observer{}, limit: 256} }
func (h *ObserverHub) Subscribe(observer Observer) func() {
	h.mu.Lock()
	id := h.next
	h.next++
	h.listeners[id] = observer
	h.mu.Unlock()
	return func() { h.mu.Lock(); delete(h.listeners, id); h.mu.Unlock() }
}
func (h *ObserverHub) Publish(event Event) {
	h.mu.Lock()
	h.history = append(h.history, event)
	if len(h.history) > h.limit {
		start := len(h.history) - h.limit
		h.history = append([]Event(nil), h.history[start:]...)
	}
	listeners := make([]Observer, 0, len(h.listeners))
	for _, listener := range h.listeners {
		listeners = append(listeners, listener)
	}
	h.mu.Unlock()
	for _, listener := range listeners {
		listener(event)
	}
}
func (h *ObserverHub) History(runID string) []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()
	events := make([]Event, 0)
	for _, event := range h.history {
		if runID == "" || event.RunID == runID {
			events = append(events, event)
		}
	}
	return events
}
func (c *Controller) Events(runID string) []Event      { return c.observers.History(runID) }
func (c *Controller) Observe(observer Observer) func() { return c.observers.Subscribe(observer) }
