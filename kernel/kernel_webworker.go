package kernel

import (
	"net"

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

func (k *webWorkerKernel) Read(handle Handle, data []byte) (int, error) {
	res, err := k.client.Invoke("Read", []interface{}{handle, len(data)}, nil)
	if err != nil {
		return 0, err
	}
	b := toBytes(res[0])
	copy(data, b)
	return len(b), nil
}

func (k *webWorkerKernel) Write(handle Handle, data []byte) (int, error) {
	b := js.NewArrayBuffer(data)
	res, err := k.client.Invoke("Write", []interface{}{handle, b}, []interface{}{b})
	if err != nil {
		return 0, err
	}
	return res[0].Int(), nil
}

func (k *webWorkerKernel) Close(handle Handle) error {
	_, err := k.client.Invoke("Close", []interface{}{handle}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (k *webWorkerKernel) StartProcess(name string, env []string) (handle Handle, err error) {
	res, err := k.client.Invoke("StartProcess", []interface{}{name, env}, nil)
	if err != nil {
		return 0, err
	}
	return Handle(res[0].Int64()), nil
}
