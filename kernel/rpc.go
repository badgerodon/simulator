package kernel

import (
	"errors"
	"math"

	"github.com/gopherjs/gopherjs/js"
)

type rcpResult struct {
	results []*js.Object
	err     error
}

// An RPCMessageChannelClient implements RPC over a message channel
type RPCMessageChannelClient struct {
	port     *js.Object
	incoming chan rcpResult
}

// NewRPCMessageChannelClient creates a new RPCMessageChannelClient
func NewRPCMessageChannelClient(port *js.Object) *RPCMessageChannelClient {
	client := &RPCMessageChannelClient{
		port:     port,
		incoming: make(chan rcpResult, 1),
	}
	port.Set("onmessage", func(evt *js.Object) {
		switch evt.Get("data").Index(0).String() {
		case "error":
			err := errors.New(evt.Get("data").Index(1).String())
			client.incoming <- rcpResult{err: err}
		case "results":
			results := make([]*js.Object, 0, evt.Get("data").Length()-1)
			for i := 1; i < evt.Get("data").Length(); i++ {
				results = append(results, evt.Get("data").Index(i))
			}
			client.incoming <- rcpResult{results: results}
		default:
			panic("unknown message type")
		}
	})
	return client
}

// Invoke invokes a method on the RPC Server
func (client *RPCMessageChannelClient) Invoke(name string, arguments []interface{}, transfer []interface{}) (results []*js.Object, err error) {
	if transfer == nil {
		client.port.Call("postMessage", append([]interface{}{name}, arguments...))
	} else {
		client.port.Call("postMessage", append([]interface{}{name}, arguments...), transfer)
	}
	res := <-client.incoming
	return res.results, res.err
}

type rpcRequest struct {
	method    string
	arguments []*js.Object
}

// An RPCMessageChannelServer implements an RPC server over a message channel
type RPCMessageChannelServer struct {
	port     *js.Object
	incoming chan rpcRequest
}

// NewRPCMessageChannelServer creates a new RPCMessageChannelServer
func NewRPCMessageChannelServer(port *js.Object, handler func(method string, arguments []*js.Object) (results, transfer []*js.Object, err error)) *RPCMessageChannelServer {
	server := &RPCMessageChannelServer{
		port:     port,
		incoming: make(chan rpcRequest, math.MaxInt16),
	}
	port.Set("onmessage", func(evt *js.Object) {
		arguments := make([]*js.Object, 0, evt.Get("data").Length()-1)
		for i := 1; i < evt.Get("data").Length(); i++ {
			arguments = append(arguments, evt.Get("data").Index(i))
		}
		server.incoming <- rpcRequest{
			method:    evt.Get("data").Index(0).String(),
			arguments: arguments,
		}
	})
	go func() {
		for req := range server.incoming {
			res, transfer, err := handler(req.method, req.arguments)
			if err != nil {
				port.Call("postMessage", []interface{}{"error", err.Error()})
			} else {
				port.Call("postMessage", append([]*js.Object{js.InternalObject("results")}, res...), transfer)
			}
		}
	}()
	return server
}
