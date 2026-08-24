package engine

import (
	"fmt"
	"sync"
)

type completionResources struct {
	mu           sync.Mutex
	stream       chan Event
	owner        string
	reportClosed bool
}

func newCompletionResources() *completionResources {
	return &completionResources{}
}

func (r *completionResources) Attach(runID string) func() {
	r.mu.Lock()
	if r.stream == nil {
		r.stream = make(chan Event, 4)
		r.owner = runID
	}
	stream := r.stream
	owner := r.owner
	r.mu.Unlock()
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.stream == stream && !r.reportClosed {
			close(r.stream)
			r.reportClosed = true
			r.owner = owner
		}
	}
}

func (r *completionResources) RequireReportGate(runID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stream == nil {
		panic("completion report resource is not attached")
	}
	if r.reportClosed {
		panic(fmt.Sprintf("completion report resource is closed by %s before %s", r.owner, runID))
	}
}

func (r *completionResources) Notify(event Event) {
	r.mu.Lock()
	stream := r.stream
	closed := r.reportClosed
	owner := r.owner
	r.mu.Unlock()
	if closed {
		panic(fmt.Sprintf("completion event stream for %s is closed", owner))
	}
	stream <- event
}
