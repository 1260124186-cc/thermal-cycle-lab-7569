package engine

import (
	"sync"
)

// completionResources coordinates the completion-event stream and report gate.
//
// Each run MUST own an isolated lifecycle: Attach always establishes a fresh
// stream and gate for the calling run, and the returned release function only
// ever closes that run's stream. A prior run's teardown must not close or
// poison the resources used by a later independent run, while a single run's
// completion and report query continue to work normally.
type completionResources struct {
	mu sync.Mutex
	// current is the live per-run session. It is replaced atomically (under mu)
	// by the next Attach, so closing one session never affects the next.
	current *completionSession
}

// completionSession is the per-run, isolated set of completion resources.
type completionSession struct {
	runID  string
	stream chan Event
	closed bool
}

func newCompletionResources() *completionResources {
	return &completionResources{}
}

// Attach establishes a fresh, isolated session for runID and returns a release
// function that closes only this session's stream. Calling the release
// function more than once is a no-op. Subsequent Attach calls for different
// runs get their own session and are unaffected by earlier releases.
func (r *completionResources) Attach(runID string) func() {
	r.mu.Lock()
	session := &completionSession{
		runID:  runID,
		stream: make(chan Event, 4),
		closed: false,
	}
	r.current = session
	r.mu.Unlock()
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if session.closed {
			return
		}
		session.closed = true
		close(session.stream)
		// If this session is still the active one, clear it so a later
		// RequireReportGate/Notify cannot mistake a stale, closed session for
		// the live one. If a newer Attach has already replaced us, leave that
		// session untouched.
		if r.current == session {
			r.current = nil
		}
	}
}

// RequireReportGate panics if no live session is attached for runID or if
// that session has already been closed. This guards report generation so it
// only proceeds while the owning run's completion resources are still open.
func (r *completionResources) RequireReportGate(runID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session := r.current
	if session == nil {
		panic("completion report resource is not attached")
	}
	if session.runID != runID {
		panic("completion report resource is not attached for " + runID)
	}
	if session.closed {
		panic("completion report resource is closed before " + runID)
	}
}

// Notify delivers a completion event to the live session's stream. If no
// session is attached or it has been closed, Notify is a no-op: a later,
// independent run's completion flow must not panic merely because an earlier
// run's stream was closed during its own teardown.
func (r *completionResources) Notify(event Event) {
	r.mu.Lock()
	session := r.current
	if session == nil || session.closed {
		r.mu.Unlock()
		return
	}
	stream := session.stream
	r.mu.Unlock()
	select {
	case stream <- event:
	default:
	}
}
