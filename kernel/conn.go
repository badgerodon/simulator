package kernel

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/gopherjs/gopherjs/js"
)

type ackedMessagePortConn struct {
	srcPort     int
	dstPort     int
	messagePort *js.Object

	ack  chan struct{}
	recv chan []byte
	rdr  *ChannelReader
}

func newAckedMessagePortConn(srcPort, dstPort int, messagePort *js.Object) *ackedMessagePortConn {
	c := &ackedMessagePortConn{
		srcPort:     srcPort,
		dstPort:     dstPort,
		messagePort: messagePort,

		ack:  make(chan struct{}, 1),
		recv: make(chan []byte, math.MaxInt16),
	}
	c.rdr = NewChannelReader(c.recv)
	c.messagePort.Set("onmessage", func(evt *js.Object) {
		//js.Global.Get("console").Call("log", fmt.Sprintf("conn %v:%v message", c.srcPort, c.dstPort), evt.Get("data"))
		method := evt.Get("data").Index(0).String()
		switch method {
		case "close":
			c.Close()
		case "ack":
			c.trace("ACK", nil)
			select {
			case c.ack <- struct{}{}:
			default:
			}
		case "message":
			msg := toBytes(evt.Get("data").Index(1))
			c.trace("RCV", msg)
			select {
			case c.recv <- msg:
			default:
				fmt.Println("RECEIVE BUFFER FULL, CLOSING CONNECTION")
				c.Close()
			}
			c.messagePort.Call("postMessage", []interface{}{"ack"})
		default:
			panic(fmt.Sprintf("method %s is not implemented", method))
		}
	})
	return c
}

func (c *ackedMessagePortConn) Read(b []byte) (n int, err error) {
	return c.rdr.Read(b)
}

func (c *ackedMessagePortConn) Write(b []byte) (n int, err error) {
	c.trace("SND", b)
	buf := js.NewArrayBuffer(b)
	c.messagePort.Call("postMessage", []interface{}{
		"message",
		buf,
	}, []interface{}{buf})
	<-c.ack
	return len(b), nil
}

func (c *ackedMessagePortConn) Close() error {
	panic("CLOSE!")
}

func (c *ackedMessagePortConn) LocalAddr() net.Addr {
	return addr{c.srcPort}
}

func (c *ackedMessagePortConn) RemoteAddr() net.Addr {
	return addr{c.dstPort}
}

func (c *ackedMessagePortConn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}

func (c *ackedMessagePortConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *ackedMessagePortConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *ackedMessagePortConn) trace(typ string, data []byte) {
	//fmt.Printf("%s %05d:%05d %x\n", typ, c.srcPort, c.dstPort, data)
}
