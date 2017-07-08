package kernel

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/gopherjs/gopherjs/js"
)

type messagePortListener struct {
	messagePort *js.Object
	networkPort int
	incoming    chan net.Conn
}

func newMessagePortListener(networkPort int, messagePort *js.Object) *messagePortListener {
	l := &messagePortListener{
		messagePort: messagePort,
		networkPort: networkPort,
		incoming:    make(chan net.Conn, 64),
	}
	l.messagePort.Set("onmessage", func(evt *js.Object) {
		method := evt.Get("data").Index(0).String()
		switch method {
		case "close":
			l.Close()
		case "connection":
			connMessagePort := evt.Get("data").Index(1)
			conn := newMessagePortConn(l.networkPort, -1, connMessagePort)
			select {
			case l.incoming <- conn:
			default:
				conn.Close()
			}
		default:
			panic(fmt.Sprintf("method %s not implemented", method))
		}
	})
	return l
}

func (l *messagePortListener) Accept() (net.Conn, error) {
	conn := <-l.incoming
	return conn, nil
}

func (l *messagePortListener) Close() error {
	if l.messagePort != nil {
		l.messagePort.Call("postMessage", []interface{}{
			"close",
		})
		l.messagePort.Call("close")
	}
	l.messagePort = nil
	return nil
}

func (l *messagePortListener) Addr() net.Addr {
	return addr{l.networkPort}
}

type messagePortConn struct {
	srcPort     int
	dstPort     int
	messagePort *js.Object

	r *ChannelReader
	w *ChannelWriter
}

func newMessagePortConn(srcPort, dstPort int, messagePort *js.Object) *messagePortConn {
	recv := make(chan []byte, math.MaxInt32)
	send := make(chan []byte, math.MaxInt32)

	c := &messagePortConn{
		srcPort:     srcPort,
		dstPort:     dstPort,
		messagePort: messagePort,

		r: NewChannelReader(recv),
		w: NewChannelWriter(send),
	}
	go func() {
		for data := range send {
			buf := js.NewArrayBuffer(data)
			c.messagePort.Call("postMessage", []interface{}{
				"message",
				buf,
			}, []interface{}{
				buf,
			})
		}
	}()
	c.messagePort.Set("onmessage", func(evt *js.Object) {
		method := evt.Get("data").Index(0).String()
		switch method {
		case "close":
			c.Close()
		case "message":
			data := evt.Get("data").Index(1).Interface().([]byte)
			select {
			case recv <- data:
			default:
			}
		default:
			panic(fmt.Sprintf("method %s is not implemented", method))
		}
	})
	return c
}

func (c *messagePortConn) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

func (c *messagePortConn) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

func (c *messagePortConn) Close() error {
	c.messagePort.Call("postMessage", []interface{}{"close"})
	c.messagePort.Call("close")
	return nil
}

func (c *messagePortConn) LocalAddr() net.Addr {
	return addr{c.srcPort}
}

func (c *messagePortConn) RemoteAddr() net.Addr {
	return addr{c.dstPort}
}

func (c *messagePortConn) SetDeadline(t time.Time) error {
	c.r.SetDeadline(t)
	c.w.SetDeadline(t)
	return nil
}

func (c *messagePortConn) SetReadDeadline(t time.Time) error {
	c.r.SetDeadline(t)
	return nil
}

func (c *messagePortConn) SetWriteDeadline(t time.Time) error {
	c.w.SetDeadline(t)
	return nil
}
