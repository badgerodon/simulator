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

type kernel interface {
	Dial(networkPort int) (net.Conn, error)
	Listen(networkPort int) (net.Listener, error)
}

type coreKernel struct {
	listeners map[int]*js.Object
	nextPort  int
}

func newCoreKernel() *coreKernel {
	return &coreKernel{
		listeners: make(map[int]*js.Object),
		nextPort:  10000,
	}
}

func (k *coreKernel) Dial(networkPort int) (net.Conn, error) {
	liPort, ok := k.listeners[networkPort]
	if !ok {
		return nil, fmt.Errorf("unavailable")
	}
	obj := js.Global.Get("MessageChannel").New()
	conn := newAckedMessagePortConn(k.nextPort, networkPort, obj.Get("port1"))
	liPort.Call("postMessage", []interface{}{
		"connection",
		k.nextPort,
		obj.Get("port2"),
	}, []interface{}{
		obj.Get("port2"),
	})
	k.nextPort++
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

var defaultKernel kernel
