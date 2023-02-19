// Copyright (c) 2023 Shivaram Lingamneni
// released under the 0BSD license

package godgets

import (
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsConnMaxMessage = 4 * 1024 * 1024
)

// WSStreamConn turns a websocket back into a full-duplex byte stream,
// because it's 2023 and we can't have nice things.
type WSStreamConn struct {
	conn *websocket.Conn

	openReader io.Reader
	readErr    error
}

// compile-time assertion that *WSStreamConn implements net.Conn:
var _ net.Conn = (*WSStreamConn)(nil)

func NewWSStreamConn(conn *websocket.Conn) *WSStreamConn {
	conn.SetReadLimit(wsConnMaxMessage)
	return &WSStreamConn{
		conn: conn,
	}
}

func (w *WSStreamConn) Read(b []byte) (n int, err error) {
	for {
		if w.readErr != nil {
			return 0, w.readErr
		}
		if len(b) == 0 {
			return 0, nil
		}

		if w.openReader != nil {
			n, err = w.openReader.Read(b)
			if err == nil {
				return
			} else if err == io.EOF {
				w.openReader = nil
				if n == 0 {
					// fall through and open a new reader
				} else {
					return n, nil
				}
			} else {
				w.openReader = nil
				w.readErr = err
				return
			}
		}

		// don't care about message type here
		_, w.openReader, w.readErr = w.conn.NextReader()
		if w.readErr != nil {
			return 0, w.readErr
		}
		// loop back around and read from the open reader
	}
}

func (w *WSStreamConn) Write(b []byte) (n int, err error) {
	// don't send giant messages in case it makes middleboxes angry
	for i := 0; i < len(b); i += wsConnMaxMessage {
		end := i + wsConnMaxMessage
		if end > len(b) {
			end = len(b)
		}
		var written int
		written, err = w.realWrite(b[i:end])
		n += written
		if err != nil {
			return
		}
	}
	return
}

func (w *WSStreamConn) realWrite(b []byte) (n int, err error) {
	writer, err := w.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return
	}
	n, err = writer.Write(b)
	if err != nil {
		return
	}
	err = writer.Close()
	return
}

func (w *WSStreamConn) Close() error {
	return w.conn.Close()
}

func (w *WSStreamConn) SetDeadline(t time.Time) error {
	if err := w.conn.SetReadDeadline(t); err != nil {
		return err
	}
	if err := w.conn.SetWriteDeadline(t); err != nil {
		return err
	}
	return nil
}

func (w *WSStreamConn) SetReadDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

func (w *WSStreamConn) SetWriteDeadline(t time.Time) error {
	return w.conn.SetWriteDeadline(t)
}

func (w *WSStreamConn) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *WSStreamConn) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}
