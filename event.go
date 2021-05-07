// Copyright (c) 2021 Shivaram Lingamneni
// released under the MIT license

package godgets

type Event (chan struct{})

func NewEvent() Event {
	return make(chan struct{})
}

func (e Event) Done() {
	close(e)
}

func (e Event) Wait() {
	<-e
}
