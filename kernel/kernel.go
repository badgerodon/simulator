package kernel

import (
	"fmt"
	"net"

	"github.com/gopherjs/gopherjs/js"
)

type addr struct {
	networkPort int
}

func (a addr) Network() string {
	return "message-port"
}

func (a addr) String() string {
	return fmt.Sprint(a.networkPort)
}

type coreKernel struct {
	listeners map[int]*js.Object
}

func newCoreKernel() *coreKernel {
	return &coreKernel{
		listeners: make(map[int]*js.Object),
	}
}

func (k *coreKernel) Dial(networkPort int) (net.Conn, error) {
	liPort, ok := k.listeners[networkPort]
	if !ok {
		return nil, fmt.Errorf("unavailable")
	}
	obj := js.Global.Get("MessageChannel").New()
	conn := newMessagePortConn(-1, networkPort, obj.Get("port1"))
	liPort.Call("postMessage", []interface{}{
		"connection",
		obj.Get("port2"),
	}, []interface{}{
		obj.Get("port2"),
	})
	return conn, nil
}

func (k *coreKernel) Listen(networkPort int) (net.Listener, error) {
	_, ok := k.listeners[networkPort]
	if ok {
		return nil, fmt.Errorf("port already bound")
	}

	obj := js.Global.Get("MessageChannel").New()
	port1, port2 := obj.Get("port1"), obj.Get("port2")
	li := newMessagePortListener(networkPort, port1)
	k.listeners[networkPort] = port2
	return li, nil
}
