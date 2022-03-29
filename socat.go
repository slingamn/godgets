// Copyright (c) 2021 Shivaram Lingamneni
// released under the 0BSD license

package godgets

import (
	"io"
	"sync"
	"time"
)

// Socat connects two io.ReadWriteCloser; reads from the first are written
// to the second, and vice versa. Compare the UNIX utility socat(1).
type Socat struct {
	c1 io.ReadWriteCloser
	c2 io.ReadWriteCloser

	errors    chan error
	closeOnce sync.Once
	closeErr  error

	timeout time.Duration
}

// NewSocat starts a two-way copy between two io.ReadWriteCloser.
// `timeout` is the amount of time to wait after an EOF from one side
// for the other side to finish copying; socat(1) defaults to 0.5 seconds,
// but in a reverse proxying context 0 is probably fine.
func NewSocat(c1, c2 io.ReadWriteCloser, timeout time.Duration) *Socat {
	s := &Socat{
		c1:      c1,
		c2:      c2,
		errors:  make(chan error, 2),
		timeout: timeout,
	}
	go s.funnel(c1, c2)
	go s.funnel(c2, c1)
	return s
}

func (s *Socat) funnel(d1, d2 io.ReadWriteCloser) {
	_, err := io.Copy(d1, d2)
	s.errors <- err
	s.Close()
}

// Wait blocks until both io.ReadWriteCloser have been closed. It is not
// necessary to call Wait to ensure that they are closed.
func (s *Socat) Wait() (err error) {
	// return the first error returned by Copy()
	err = <-s.errors
	// block on the sync.Once until close is complete
	s.Close()
	return
}

// Close closes both of the io.ReadWriteCloser.
func (s *Socat) Close() (err error) {
	s.closeOnce.Do(func() {
		if s.timeout != 0 {
			time.Sleep(s.timeout)
		}
		s.closeErr = s.realClose()
	})
	return s.closeErr
}

func (s *Socat) realClose() (err error) {
	e1 := s.c1.Close()
	e2 := s.c2.Close()
	if e1 != nil {
		return e1
	}
	return e2
}
