package kernel

import (
	"fmt"
	"math"
	"net"
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

// A Handle is a pointer to an object in the system
type Handle int64

// Network returns the network
func (h Handle) Network() string {
	return "js"
}

func (h Handle) String() string {
	return fmt.Sprint(int64(h))
}

var handleCounter int64 = math.MaxUint16

// NextHandle generates the next Handle
func NextHandle() Handle {
	return Handle(atomic.AddInt64(&handleCounter, 1))
}

type kernel interface {
	Dial(networkPort Handle) (net.Conn, error)
	Listen(networkPort Handle) (net.Listener, error)
	Read(handle Handle, data []byte) (int, error)
	Write(handle Handle, data []byte) (int, error)
	StartProcess(name string, env []string) (handle Handle, err error)
	Close(handle Handle) error
}

var defaultKernel kernel

// StartProcess starts a web worker with the given name
func StartProcess(name string, env []string) (handle Handle, err error) {
	return defaultKernel.StartProcess(name, env)
}

func toBytes(obj *js.Object) []byte {
	return js.Global.Get("Uint8Array").New(obj).Interface().([]byte)
}
