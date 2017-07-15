package kernel

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gopherjs/gopherjs/js"
)

type messagePortListener struct {
	messagePort *js.Object
	networkPort int
	incoming    chan net.Conn
	onClose     func()
}

func newMessagePortListener(networkPort int, messagePort *js.Object, onClose func()) *messagePortListener {
	l := &messagePortListener{
		messagePort: messagePort,
		networkPort: networkPort,
		incoming:    make(chan net.Conn, 64),
		onClose:     onClose,
	}
	l.messagePort.Set("onmessage", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "listen message", evt.Get("data"))
		method := evt.Get("data").Index(0).String()
		switch method {
		case "close":
			l.Close()
		case "connection":
			connPort := evt.Get("data").Index(1).Int()
			connMessagePort := evt.Get("data").Index(2)
			conn := newAckedMessagePortConn(l.networkPort, connPort, connMessagePort, nil)
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
	if l.onClose != nil {
		l.onClose()
		l.onClose = nil
	}
	return nil
}

func (l *messagePortListener) Addr() net.Addr {
	return addr{l.networkPort}
}

type messagePortConn struct {
	srcPort     int
	dstPort     int
	messagePort *js.Object

	ctx    context.Context
	cancel context.CancelFunc

	readCtx, writeCtx       context.Context
	readCancel, writeCancel context.CancelFunc

	qr *inOrderQueueReader
	qw *queue
}

func newMessagePortConn(srcPort, dstPort int, messagePort *js.Object) *messagePortConn {
	ctx, cancel := context.WithCancel(context.Background())

	send := newQueue()
	recv := newInOrderQueue()

	c := &messagePortConn{
		srcPort:     srcPort,
		dstPort:     dstPort,
		messagePort: messagePort,

		ctx:    ctx,
		cancel: cancel,

		readCtx:  ctx,
		writeCtx: ctx,

		qr: newInOrderQueueReader(recv),
		qw: send,
	}
	go func() {
		defer func() {
			c.messagePort.Call("close")
		}()
		var counter uint16
		for {
			msg, err := send.dequeue(c.ctx)
			if err != nil {
				return
			}
			fmt.Printf("send %05d:%05d %04d %x\n", c.srcPort, c.dstPort, counter, msg)
			buf := js.NewArrayBuffer(msg)
			c.messagePort.Call("postMessage", []interface{}{
				"message",
				counter,
				buf,
			}, []interface{}{
				buf,
			})
			counter++
		}
	}()

	c.messagePort.Set("onmessage", func(evt *js.Object) {
		//js.Global.Get("console").Call("log", fmt.Sprintf("conn %v:%v message", c.srcPort, c.dstPort), evt.Get("data"))
		method := evt.Get("data").Index(0).String()
		switch method {
		case "close":
			c.Close()
		case "message":
			counter := uint16(evt.Get("data").Index(1).Int())
			msg := toBytes(evt.Get("data").Index(2))
			fmt.Printf("recv %05d:%05d %04d %x\n", c.srcPort, c.dstPort, counter, msg)
			recv.enqueue(counter, msg)
		default:
			panic(fmt.Sprintf("method %s is not implemented", method))
		}
	})
	return c
}

func (c *messagePortConn) Read(b []byte) (n int, err error) {
	return c.qr.Read(c.readCtx, b)
}

func (c *messagePortConn) Write(b []byte) (n int, err error) {
	c.qw.enqueue(b)
	return len(b), nil
}

func (c *messagePortConn) Close() error {
	panic("CLOSE!")
	if c.readCancel != nil {
		c.readCancel()
		c.readCancel = nil
	}
	if c.writeCancel != nil {
		c.writeCancel()
		c.writeCancel = nil
	}
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	return nil
}

func (c *messagePortConn) LocalAddr() net.Addr {
	return addr{c.srcPort}
}

func (c *messagePortConn) RemoteAddr() net.Addr {
	return addr{c.dstPort}
}

func (c *messagePortConn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}

func (c *messagePortConn) SetReadDeadline(t time.Time) error {
	if c.readCancel != nil {
		c.readCancel()
	}
	if t.IsZero() {
		c.readCtx = c.ctx
		c.readCancel = nil
	} else {
		c.readCtx, c.readCancel = context.WithDeadline(c.ctx, t)
	}
	return nil
}

func (c *messagePortConn) SetWriteDeadline(t time.Time) error {
	if c.writeCancel != nil {
		c.writeCancel()
	}
	if t.IsZero() {
		c.writeCtx = c.ctx
		c.writeCancel = nil
	} else {
		c.writeCtx, c.writeCancel = context.WithDeadline(c.ctx, t)
	}
	return nil
}

func toBytes(obj *js.Object) []byte {
	return js.Global.Get("Uint8Array").New(obj).Interface().([]byte)
}
