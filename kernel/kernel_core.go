package kernel

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"
	"syscall"

	"github.com/gopherjs/gopherjs/js"
)

type coreKernelWorker struct {
	*js.Object
	waiter chan struct{}
}

type coreKernel struct {
	sync.Mutex
	listeners   map[Handle]*js.Object
	workers     map[Handle]coreKernelWorker
	connections map[Handle]*js.Object
	readers     map[int]io.Reader
	writers     map[int]io.WriteCloser
	nextPort    int64
	nextPID     int
}

func newCoreKernel() *coreKernel {
	return &coreKernel{
		listeners:   make(map[Handle]*js.Object),
		workers:     make(map[Handle]coreKernelWorker),
		connections: make(map[Handle]*js.Object),
		readers:     make(map[int]io.Reader),
		writers:     make(map[int]io.WriteCloser),
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
		k.Close(int(localPort))
	})
	js.Global.Get("console").Call("log", "CK", "Dial", networkPort, localPort)
	return conn, nil
}

func (k *coreKernel) Listen(networkPort Handle) (net.Listener, error) {
	port, err := k.listen(networkPort)
	if err != nil {
		return nil, err
	}
	li := newMessagePortListener(networkPort, port, func() {
		k.Close(int(networkPort))
	})
	js.Global.Get("console").Call("log", "CK", "Listen", networkPort)
	return li, nil
}

func (k *coreKernel) Pipe() (r, w int, err error) {
	c := make(chan []byte, 1)
	cr := NewChannelReader(c)
	cw := NewChannelWriter(c)

	r = int(NextHandle())
	w = int(NextHandle())

	k.Lock()
	k.readers[r] = cr
	k.writers[w] = cw
	k.Unlock()

	js.Global.Get("console").Call("log", "CK", "Pipe", r, w)

	return r, w, nil
}

func (k *coreKernel) Read(fd int, p []byte) (int, error) {
	k.Lock()
	r, ok := k.readers[fd]
	k.Unlock()
	if ok {
		return r.Read(p)
	}

	js.Global.Get("console").Call("log", "read", fd)

	return 0, errors.New("Read not implemented")
}

func (k *coreKernel) Wait(pid int) error {
	js.Global.Get("console").Call("log", "CK", "Wait", pid)
	k.Lock()
	w, ok := k.workers[Handle(pid)]
	k.Unlock()
	if !ok {
		return errors.New("process not found")
	}
	<-w.waiter
	return nil
}

func (k *coreKernel) Write(fd int, p []byte) (int, error) {
	js.Global.Get("console").Call("log", "CK", "Write", fd, p)
	k.Lock()
	w, ok := k.writers[fd]
	k.Unlock()
	if ok {
		return w.Write(p)
	}

	switch fd {
	case 1:
		s := bufio.NewScanner(bytes.NewReader(p))
		for s.Scan() {
			js.Global.Get("console").Call("log", s.Text())
		}
		return len(p), nil
	case 2:
		s := bufio.NewScanner(bytes.NewReader(p))
		for s.Scan() {
			js.Global.Get("console").Call("warn", s.Text())
		}
		return len(p), nil
	}
	return 0, errors.New("Write not implemented")
}

func (k *coreKernel) Close(fd int) error {
	js.Global.Get("console").Call("log", "CK", "Close", fd)
	handle := Handle(fd)

	k.Lock()
	li, ok := k.listeners[handle]
	delete(k.listeners, handle)
	k.Unlock()
	if ok {
		js.Global.Get("console").Call("log", "CK", "Close", "Listener", handle, li)
		li.Call("close")
		return nil
	}

	k.Lock()
	w, ok := k.workers[handle]
	delete(k.workers, handle)
	k.Unlock()
	if ok {
		js.Global.Get("console").Call("log", "CK", "Close", "Worker", handle, w)
		close(w.waiter)
		w.Call("terminate")
		return nil
	}

	// close connections
	k.Lock()
	c, ok := k.connections[handle]
	delete(k.connections, handle)
	k.Unlock()
	if ok {
		js.Global.Get("console").Call("log", "CK", "Close", "Connection", handle, c)
		c.Call("close")
		return nil
	}

	// close readers
	{
		k.Lock()
		r, ok := k.readers[int(handle)]
		delete(k.readers, int(handle))
		k.Unlock()
		if ok {
			js.Global.Get("console").Call("log", "CK", "Close", "Reader", handle, r)
			return nil
		}
	}

	// close writers
	{
		k.Lock()
		w, ok := k.writers[int(handle)]
		delete(k.writers, int(handle))
		k.Unlock()
		if ok {
			js.Global.Get("console").Call("log", "CK", "Close", "Writer", handle, w)
			w.Close()
			return nil
		}
	}

	return nil
}

func (k *coreKernel) StartProcess(argv0 string, argv []string, attr *syscall.ProcAttr) (pid int, handle uintptr, err error) {
	js.Global.Get("console").Call("log", "CK", "StartProcess", argv0, argv, attr)
	pid = int(NextHandle())

	type Result struct {
		location string
		err      error
	}
	c := make(chan Result, 1)
	vs := make(url.Values)
	vs.Set("import_path", argv0)
	vs.Set("branch", "master")
	js.Global.Get("fetch").Invoke("/api/build?"+vs.Encode()).
		Call("then", func(res *js.Object) *js.Object {
			status := res.Get("status").Int() / 100
			if status != 2 && status != 3 {
				res.Call("text").Call("then", func(txt *js.Object) {
					c <- Result{err: errors.New(txt.String())}
				})
				return nil
			}
			return res.Call("json")
		}).
		Call("then", func(res *js.Object) {
			c <- Result{location: res.Get("location").String()}
		}).
		Call("catch", func(err *js.Object) {
			c <- Result{err: errors.New(err.String())}
		})
	result := <-c
	if result.err != nil {
		return 0, handle, result.err
	}

	vs = make(url.Values)
	for i, fd := range attr.Files {
		vs.Set(fmt.Sprintf("files[%d]", i), fmt.Sprint(fd))
	}
	u := result.location + "?" + vs.Encode()

	worker := js.Global.Get("Worker").New(u)
	NewRPCMessageChannelServer(worker, func(method string, arguments []*js.Object) (results, transfer []*js.Object, err error) {
		switch method {
		case "Close":
			handle := Handle(arguments[0].Int64())
			err = k.Close(int(handle))
			return nil, nil, err
		case "Dial":
			networkPort := Handle(arguments[0].Int64())
			port, err := k.dial(networkPort)
			if err != nil {
				return nil, nil, err
			}
			return []*js.Object{port}, []*js.Object{port}, nil
		case "Exit":
			k.Close(pid)
			return []*js.Object{}, []*js.Object{}, nil
		case "Listen":
			networkPort := Handle(arguments[0].Int64())
			port, err := k.listen(networkPort)
			if err != nil {
				return nil, nil, err
			}
			return []*js.Object{port}, []*js.Object{port}, nil
		case "NextHandle":
			h := int64(NextHandle())
			return []*js.Object{js.InternalObject(h)}, nil, nil
		case "Read":
			handle := Handle(arguments[0].Int64())
			sz := arguments[1].Int()
			buf := make([]byte, sz)
			n, err := k.Read(int(handle), buf)
			if err != nil {
				return nil, nil, err
			}
			js.NewArrayBuffer(buf[:n])
			jbuf := js.NewArrayBuffer(buf)
			return []*js.Object{jbuf}, []*js.Object{jbuf}, nil
		case "Write":
			fd := arguments[0].Int()
			buf := toBytes(arguments[1])
			n, err := k.Write(fd, buf)
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
	k.workers[Handle(pid)] = coreKernelWorker{
		Object: worker,
		waiter: make(chan struct{}),
	}
	k.Unlock()

	return pid, uintptr(pid), nil
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
