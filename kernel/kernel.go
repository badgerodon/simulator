package kernel

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

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
	Read(fd uintptr, data []byte) (int, error)
	Write(fd uintptr, data []byte) (int, error)
}

var defaultKernel kernel

type coreKernel struct {
	listeners map[int]*js.Object
	nextPort  int64
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

	port := int(atomic.AddInt64(&k.nextPort, 1))

	obj := js.Global.Get("MessageChannel").New()
	conn := newAckedMessagePortConn(port, networkPort, obj.Get("port1"), func() {
		liPort.Call("postMessage", []interface{}{
			"close",
			port,
		})
	})
	liPort.Call("postMessage", []interface{}{
		"connection",
		port,
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
	li := newMessagePortListener(networkPort, port1, func() {

	})
	k.listeners[networkPort] = port2
	return li, nil
}

func (k *coreKernel) Read(fd uintptr, data []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (k *coreKernel) Write(fd uintptr, data []byte) (int, error) {
	switch fd {
	case 1:
		s := bufio.NewScanner(bytes.NewReader(data))
		for s.Scan() {
			js.Global.Get("console").Call("log", s.Text())
		}
		return len(data), nil
	case 2:
		s := bufio.NewScanner(bytes.NewReader(data))
		for s.Scan() {
			js.Global.Get("console").Call("warn", s.Text())
		}
		return len(data), nil
	}
	return 0, errors.New("not implemented")
}

type workerKernel struct {
	context  *js.Object
	nextPort int64

	mu        sync.Mutex
	listeners map[int]*js.Object
}

func newWorkerKernel() *workerKernel {
	return &workerKernel{
		context:  js.Global,
		nextPort: 10000,
	}
}

func (k *workerKernel) Dial(networkPort int) (net.Conn, error) {
	obj := js.Global.Get("MessageChannel").New()

	port := int(atomic.AddInt64(&k.nextPort, 1))

	conn := newAckedMessagePortConn(port, networkPort, obj.Get("port1"), func() {
		k.context.Call("postMessage", []interface{}{
			"close",
			port,
		})
	})
	k.context.Call("postMessage", []interface{}{
		"dial",
		networkPort,
		port,
		obj.Get("port2"),
	}, []interface{}{
		obj.Get("port2"),
	})
	return conn, nil
}

func (k *workerKernel) Listen(networkPort int) (net.Listener, error) {
	if networkPort == 0 {
		networkPort = int(atomic.AddInt64(&k.nextPort, 1))
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	if _, ok := k.listeners[networkPort]; ok {
		return nil, fmt.Errorf("port already bound")
	}

	obj := k.context.Get("MessageChannel").New()
	port1, port2 := obj.Get("port1"), obj.Get("port2")
	k.listeners[networkPort] = port2
	li := newMessagePortListener(networkPort, port1, func() {
		k.mu.Lock()
		delete(k.listeners, networkPort)
		k.mu.Unlock()
	})
	return li, nil
}
