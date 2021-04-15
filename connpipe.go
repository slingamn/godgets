// Copyright (c) 2021 Shivaram Lingamneni
// released under the MIT license

package godgets

import (
	"net"
	"sync"
)

const (
	connPipeBufferSize = 4096
)

// connects two net.Conn; reads from the first are written to the second,
// and vice versa
type ConnPipe struct {
	c1 net.Conn
	c2 net.Conn

	done      chan error
	closeOnce sync.Once
}

func NewConnPipe(c1, c2 net.Conn) *ConnPipe {
	c := &ConnPipe{
		c1:   c1,
		c2:   c2,
		done: make(chan error, 1),
	}
	go c.funnel(c1, c2)
	go c.funnel(c2, c1)
	return c
}

func (t *ConnPipe) funnel(d1, d2 net.Conn) {
	buf := make([]byte, connPipeBufferSize)
	for {
		n, err := d1.Read(buf)
		if err != nil {
			select {
			case t.done <- err:
			default:
			}
			return
		}
		_, err = d2.Write(buf[:n])
		if err != nil {
			select {
			case t.done <- err:
			default:
			}
			return
		}
	}
}

func (t *ConnPipe) Wait() (err error) {
	err = <-t.done
	t.Close()
	return
}

func (t *ConnPipe) Close() {
	t.closeOnce.Do(func() {
		t.realClose()
	})
}

func (t *ConnPipe) realClose() (err error) {
	e1 := t.c1.Close()
	e2 := t.c2.Close()
	if e1 != nil {
		return e1
	}
	return e2
}
