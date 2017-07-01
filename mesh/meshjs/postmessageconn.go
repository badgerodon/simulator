package meshjs

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"golang.org/x/net/context"
)

// A PostMessageConn is a net.Conn implemented using postMessage
type PostMessageConn struct {
	object    *js.Object
	recv      chan []byte
	send      chan []byte
	recvbuf   []byte
	closed    chan struct{}
	closeonce sync.Once

	mu                          sync.Mutex
	readDeadline, writeDeadline time.Time
}

// NewPostMessageConn creates a new PostMessageConn
func NewPostMessageConn(object *js.Object) *PostMessageConn {
	conn := &PostMessageConn{
		object: object,
		recv:   make(chan []byte),
	}
	conn.object.Set("onmessage", func(evt *js.Object) {
		data, ok := evt.Get("data").Interface().([]byte)
		if !ok {
			warn("invalid data", data)
			return
		}

		go func() {
			conn.recv <- data
		}()
	})
	go func() {
		// close the connection if for any reason sending fails
		defer conn.Close()

		for data := range conn.send {
			arr := js.NewArrayBuffer(data)
			conn.object.Call("postMessage", arr, []interface{}{arr})
		}
	}()
	return conn
}

// Read reads data posted to the connection
func (c *PostMessageConn) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, io.ErrShortBuffer
	}

	for len(c.recvbuf) == 0 {
		var waiter <-chan time.Time
		c.mu.Lock()
		if !c.readDeadline.IsZero() {
			wait := time.Now().Sub(c.readDeadline)
			if wait < 0 {
				wait = 0
			}
			waiter = time.After(wait)
		}
		c.mu.Unlock()

		select {
		case <-c.closed:
			return 0, context.Canceled
		case <-waiter:
			return 0, context.DeadlineExceeded
		case c.recvbuf = <-c.recv:
		}
	}

	if len(c.recvbuf) <= len(b) {
		n = len(c.recvbuf)
		copy(b, c.recvbuf)
		c.recvbuf = nil
	} else {
		n = len(b)
		copy(b, c.recvbuf)
		c.recvbuf = c.recvbuf[len(b):]
	}

	return n, nil
}

// Write writes data using postMessage
func (c *PostMessageConn) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}

	var waiter <-chan time.Time
	c.mu.Lock()
	if !c.writeDeadline.IsZero() {
		wait := time.Now().Sub(c.readDeadline)
		if wait < 0 {
			wait = 0
		}
		waiter = time.After(wait)
	}
	c.mu.Unlock()

	select {
	case <-c.closed:
		return 0, context.Canceled
	case <-waiter:
		return 0, context.DeadlineExceeded
	case c.send <- b:
	}

	return len(b), nil
}

// Close closes the connection
func (c *PostMessageConn) Close() error {
	c.closeonce.Do(func() {
		close(c.closed)
	})
	return nil
}

// LocalAddr returns the local address
func (c *PostMessageConn) LocalAddr() net.Addr {
	panic("not implemented")
}

// RemoteAddr returns the remote address
func (c *PostMessageConn) RemoteAddr() net.Addr {
	panic("not implemented")
}

// SetDeadline sets the deadlines
func (c *PostMessageConn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readDeadline = t
	c.writeDeadline = t
	return nil
}

// SetReadDeadline sets the read deadline
func (c *PostMessageConn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readDeadline = t
	return nil
}

// SetWriteDeadline sets the write deadline
func (c *PostMessageConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeDeadline = t
	return nil
}
