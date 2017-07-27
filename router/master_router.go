package router

import (
	"fmt"
	"net"

	"github.com/badgerodon/simulator/router/routerpb"
	"github.com/gopherjs/gopherjs/js"
)

// A MasterRouter is a router within the outermost window context
type MasterRouter struct {
	messages chan *routerpb.Message
}

// NewMasterRouter creates a new MasterRouter
func NewMasterRouter() *MasterRouter {
	r := &MasterRouter{
		messages: make(chan *routerpb.Message, 1),
	}
	go func() {
		for msg := range r.messages {
			fmt.Println(msg)
		}
	}()
	return r
}

// BindWorker binds a worker javascript object to the master router
func (r *MasterRouter) BindWorker(object *js.Object) {
	object.Set("onmessage", func(evt *js.Object) {
		data := evt.Get("data").Interface()
		fmt.Println("DATA", data)
	})
}

// Dial creates a new connection
func (r *MasterRouter) Dial(network, address string) (net.Conn, error) {
	return nil, fmt.Errorf("not implemented")
}

// Listen creates a new listener
func (r *MasterRouter) Listen(net, laddr string) (net.Listener, error) {
	return nil, fmt.Errorf("not implemented")
}
