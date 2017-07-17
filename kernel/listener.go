package kernel

import (
	"fmt"
	"net"

	"github.com/gopherjs/gopherjs/js"
)

type messagePortListener struct {
	messagePort *js.Object
	networkPort Handle
	incoming    chan net.Conn
	onClose     func()
}

func newMessagePortListener(networkPort Handle, messagePort *js.Object, onClose func()) *messagePortListener {
	l := &messagePortListener{
		messagePort: messagePort,
		networkPort: networkPort,
		incoming:    make(chan net.Conn, 64),
		onClose:     onClose,
	}
	l.messagePort.Set("onmessage", func(evt *js.Object) {
		method := evt.Get("data").Index(0).String()
		switch method {
		case "close":
			l.Close()
		case "connection":
			connPort := Handle(evt.Get("data").Index(1).Int64())
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
	return l.networkPort
}
