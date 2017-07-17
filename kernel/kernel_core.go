package kernel

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/gopherjs/gopherjs/js"
)

type coreKernel struct {
	sync.Mutex
	listeners   map[Handle]*js.Object
	workers     map[Handle]*js.Object
	connections map[Handle]*js.Object
	nextPort    int64
	nextPID     int
}

func newCoreKernel() *coreKernel {
	return &coreKernel{
		listeners:   make(map[Handle]*js.Object),
		workers:     make(map[Handle]*js.Object),
		connections: make(map[Handle]*js.Object),
		nextPort:    10000,
		nextPID:     1000,
	}
}

func (k *coreKernel) Dial(networkPort Handle) (net.Conn, error) {
	localPort := NextHandle()

	port, err := k.dial(networkPort)
	if err != nil {
		return nil, err
	}

	k.Lock()
	k.connections[localPort] = port
	k.Unlock()

	conn := newAckedMessagePortConn(localPort, networkPort, port, func() {
		k.Close(localPort)
	})
	return conn, nil
}

func (k *coreKernel) Listen(networkPort Handle) (net.Listener, error) {
	port, err := k.listen(networkPort)
	if err != nil {
		return nil, err
	}
	return newMessagePortListener(networkPort, port, func() {
		k.Close(networkPort)
	}), nil
}

func (k *coreKernel) Read(handle Handle, data []byte) (int, error) {
	return 0, errors.New("Read not implemented")
}

func (k *coreKernel) Write(handle Handle, data []byte) (int, error) {
	switch handle {
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
	return 0, errors.New("Write not implemented")
}

func (k *coreKernel) Close(handle Handle) error {
	k.Lock()
	li, ok := k.listeners[handle]
	delete(k.listeners, handle)
	k.Unlock()
	if ok {
		li.Call("close")
		return nil
	}

	k.Lock()
	w, ok := k.workers[handle]
	delete(k.workers, handle)
	k.Unlock()
	if ok {
		w.Call("terminate")
		return nil
	}

	k.Lock()
	c, ok := k.connections[handle]
	delete(k.connections, handle)
	k.Unlock()
	if ok {
		c.Call("close")
		return nil
	}

	return nil
}

func (k *coreKernel) StartProcess(name string, env []string) (handle Handle, err error) {
	handle = NextHandle()

	worker := js.Global.Get("Worker").New(name)
	NewRPCMessageChannelServer(worker, func(method string, arguments []*js.Object) (results, transfer []*js.Object, err error) {
		switch method {
		case "Close":
			handle := Handle(arguments[0].Int64())
			err = k.Close(handle)
			return nil, nil, err
		case "Dial":
			networkPort := Handle(arguments[0].Int64())
			port, err := k.dial(networkPort)
			if err != nil {
				return nil, nil, err
			}
			return []*js.Object{port}, []*js.Object{port}, nil
		case "Listen":
			networkPort := Handle(arguments[0].Int64())
			port, err := k.listen(networkPort)
			if err != nil {
				return nil, nil, err
			}
			return []*js.Object{port}, []*js.Object{port}, nil
		case "Read":
			handle := Handle(arguments[0].Int64())
			sz := arguments[1].Int()
			buf := make([]byte, sz)
			n, err := k.Read(handle, buf)
			if err != nil {
				return nil, nil, err
			}
			js.NewArrayBuffer(buf[:n])
			jbuf := js.NewArrayBuffer(buf)
			return []*js.Object{jbuf}, []*js.Object{jbuf}, nil
		case "Write":
			handle := Handle(arguments[0].Int64())
			buf := toBytes(arguments[1])
			n, err := k.Write(handle, buf)
			if err != nil {
				return nil, nil, err
			}
			return []*js.Object{js.InternalObject(n)}, nil, nil
		default:
			js.Global.Get("console").Call("warn", fmt.Sprintf("%s not implemented", method))
		}
		return nil, nil, fmt.Errorf("%s not implemented", method)
	})

	k.Lock()
	k.workers[handle] = worker
	k.Unlock()

	return handle, nil
}

func (k *coreKernel) dial(networkPort Handle) (port *js.Object, err error) {
	k.Lock()
	li, ok := k.listeners[networkPort]
	k.Unlock()
	if !ok {
		return nil, fmt.Errorf("unavailable")
	}

	obj := js.Global.Get("MessageChannel").New()
	port1, port2 := obj.Get("port1"), obj.Get("port2")
	li.Call("postMessage", []interface{}{
		"connection",
		0,
		port2,
	}, []interface{}{
		port2,
	})
	return port1, nil
}

func (k *coreKernel) listen(networkPort Handle) (port *js.Object, err error) {
	k.Lock()
	defer k.Unlock()

	_, ok := k.listeners[networkPort]
	if ok {
		return nil, fmt.Errorf("port already bound")
	}

	obj := js.Global.Get("MessageChannel").New()
	port1, port2 := obj.Get("port1"), obj.Get("port2")

	k.listeners[networkPort] = port2
	return port1, nil
}
