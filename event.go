// Copyright (c) 2021 Shivaram Lingamneni
// released under the MIT license

package godgets

type Event (chan empty)

func NewEvent() Event {
	return make(chan empty)
}

func (e Event) Done() {
	close(e)
}

func (e Event) Wait() {
	<-e
}
