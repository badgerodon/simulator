package kernel

import (
	"net"
	"syscall"

	"github.com/gopherjs/gopherjs/js"
)

type webWorkerKernel struct {
	client *RPCMessageChannelClient
}

func newWebWorkerKernel(client *RPCMessageChannelClient) *webWorkerKernel {
	return &webWorkerKernel{
		client: client,
	}
}

func (k *webWorkerKernel) Dial(networkPort Handle) (net.Conn, error) {
	handle := NextHandle()
	res, err := k.client.Invoke("Dial", []interface{}{networkPort}, nil)
	if err != nil {
		return nil, err
	}
	conn := newAckedMessagePortConn(handle, networkPort, res[0], func() {
		k.client.Invoke("Close", []interface{}{handle}, nil)
	})
	return conn, nil
}

func (k *webWorkerKernel) Listen(networkPort Handle) (net.Listener, error) {
	handle := NextHandle()
	res, err := k.client.Invoke("Listen", []interface{}{networkPort}, []interface{}{})
	if err != nil {
		return nil, err
	}
	li := newMessagePortListener(networkPort, res[0], func() {
		k.client.Invoke("Close", []interface{}{handle}, nil)
	})
	return li, nil
}

func (k *webWorkerKernel) Pipe() (r, w int, err error) {
	res, err := k.client.Invoke("Pipe", nil, nil)
	if err != nil {
		return 0, 0, err
	}
	return res[0].Int(), res[1].Int(), nil
}

func (k *webWorkerKernel) Read(fd int, p []byte) (int, error) {
	res, err := k.client.Invoke("Read", []interface{}{fd, len(p)}, nil)
	if err != nil {
		return 0, err
	}
	b := toBytes(res[0])
	copy(p, b)
	return len(b), nil
}

func (k *webWorkerKernel) Write(fd int, p []byte) (int, error) {
	js.Global.Get("console").Call("log", "WW", "Write", fd, p)
	b := js.NewArrayBuffer(p)
	res, err := k.client.Invoke("Write", []interface{}{fd, b}, []interface{}{b})
	if err != nil {
		return 0, err
	}
	return res[0].Int(), nil
}

func (k *webWorkerKernel) Close(fd int) error {
	js.Global.Get("console").Call("log", "WW", "Close", fd)
	_, err := k.client.Invoke("Close", []interface{}{fd}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (k *webWorkerKernel) StartProcess(argv0 string, argv []string, attr *syscall.ProcAttr) (pid int, handle uintptr, err error) {
	res, err := k.client.Invoke("StartProcess", []interface{}{argv0, argv, attr}, nil)
	if err != nil {
		return 0, 0, err
	}
	pid = int(res[0].Int64())
	return pid, uintptr(pid), nil
}
