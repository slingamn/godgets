// Copyright (c) 2021 Shivaram Lingamneni
// released under the MIT license

package godgets

import (
	"io"
	"sync"
)

// connects two io.ReadWriteCloser; reads from the first are written to the second,
// and vice versa
type Socat struct {
	c1 io.ReadWriteCloser
	c2 io.ReadWriteCloser

	done      chan error
	closeOnce sync.Once
	closeErr  error
}

func NewSocat(c1, c2 io.ReadWriteCloser) *Socat {
	c := &Socat{
		c1:   c1,
		c2:   c2,
		done: make(chan error, 2),
	}
	go c.funnel(c1, c2)
	go c.funnel(c2, c1)
	return c
}

func (t *Socat) funnel(d1, d2 io.ReadWriteCloser) {
	_, err := io.Copy(d1, d2)
	t.done <- err
}

func (t *Socat) Wait() (err error) {
	// close immediately upon receiving an error / EOF from either Copy() call
	err = <-t.done
	t.Close()
	return
}

func (t *Socat) Close() (err error) {
	t.closeOnce.Do(func() {
		t.closeErr = t.realClose()
	})
	return t.closeErr
}

func (t *Socat) realClose() (err error) {
	e1 := t.c1.Close()
	e2 := t.c2.Close()
	if e1 != nil {
		return e1
	}
	return e2
}
