// +build js

package kernel

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
)

func init() {
	k := newCoreKernel()
	net.DefaultDialContextFunction = func(ctx context.Context, network, address string) (net.Conn, error) {
		switch network {
		case "tcp", "tcp4":
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, fmt.Errorf("invalid listen address: %v, %v", err, address)
			}
			if !allowHost(host) {
				return nil, fmt.Errorf("only localhost is supported for listening")
			}
			iport, err := strconv.Atoi(port)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %v, %v", err, port)
			}
			return k.Dial(iport)
		default:
			return nil, fmt.Errorf("only tcp is supported")
		}
		return nil, fmt.Errorf("not implemented")
	}
	net.DefaultListenFunction = func(network, address string) (net.Listener, error) {
		switch network {
		case "tcp", "tcp4":
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, fmt.Errorf("invalid listen address: %v, %v", err, address)
			}
			if !allowHost(host) {
				return nil, fmt.Errorf("only localhost is supported for listening")
			}
			iport, err := strconv.Atoi(port)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %v, %v", err, port)
			}
			return k.Listen(iport)
		default:
			return nil, fmt.Errorf("only tcp is supported")
		}
		return nil, fmt.Errorf("not implemented")
	}
	http.DefaultTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: false,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	syscall.DefaultReadFunction = k.Read
	syscall.DefaultWriteFunction = k.Write

	js.Global.Get("console").Call("log", jsbuiltin.InstanceOf(js.Global.Get("self"), js.Global.Get("Worker")))
}

func allowHost(host string) bool {
	return host == "127.0.0.1" ||
		host == "" ||
		host == "0.0.0.0" ||
		host == "locahost"
}
