package router

import "github.com/gopherjs/gopherjs/js"

// A Router routes traffic within the browser to simulate network connections and
// stdin/stdout between processes
type Router interface {
	BindWorker(object *js.Object)
	Dial(network, address string) (net.Conn, error)
	Listen(net, laddr string) (net.Listener, error)
}

// DefaultRouter is the global router used within the current execution context
var DefaultRouter Router

func init() {
	DefaultRouter = &MasterRouter{}

	net.DefaultDialFunction = func(network, address string) (net.Conn, error) {
		return DefaultRouter.Dial(network, address)
	}
	net.DefaultListenFunction = func(net, laddr string) (net.Listener, error) {
		return DefaultRouter.Listen(net, laddr)
	}
}
