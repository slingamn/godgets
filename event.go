// Copyright (c) 2021 Shivaram Lingamneni
// released under the 0BSD license

package godgets

import "time"

// Python's threading.Event with some of the APIs removed:
// https://docs.python.org/3/library/threading.html#event-objects
// in particular, this Event can only be used once, so there is no
// equivalent of clear() to reset the event.

type Event (chan struct{})

func NewEvent() Event {
	return make(chan struct{})
}

// Mark the event as completed.
func (e Event) Done() {
	close(e)
}

// Wait for the event to be completed. A timeout of 0 means no timeout;
// use IsDone() for a non-blocking check, comparable to Python's is_set().
func (e Event) Wait(timeout time.Duration) (isDone bool) {
	if timeout == 0 {
		<-e
		return true
	} else {
		timer := time.NewTimer(timeout)
		select {
		case <-e:
			isDone = true
		case <-timer.C:
		}
		timer.Stop()
		return
	}
}

func (e Event) IsDone() bool {
	select {
	case <-e:
		return true
	default:
		return false
	}
}
