package router

import (
	"fmt"
	"net"

	"github.com/gopherjs/gopherjs/js"
)

// A WorkerRouter forwards all data to master router via postMessage
type WorkerRouter struct {
	object *js.Object
}

func (r *WorkerRouter) Dial(network, address string) (net.Conn, error) {


	return nil, fmt.Errorf("not implemented")
}

func (r *WorkerRouter) Listen(net, laddr string) (net.Listener, error) {
	return nil, fmt.Errorf("not implemented")
}
