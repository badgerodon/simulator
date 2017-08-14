package kernel

import (
	"context"
	"errors"
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
	if js.Global.Get("DedicatedWorkerGlobalScope").Bool() &&
		jsbuiltin.InstanceOf(js.Global.Get("self"), js.Global.Get("DedicatedWorkerGlobalScope")) {
		defaultKernel = newWebWorkerKernel(NewRPCMessageChannelClient(js.Global.Get("self")))
	} else {
		defaultKernel = newCoreKernel()
	}

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
			return defaultKernel.Dial(Handle(iport))
		default:
			return nil, fmt.Errorf("only tcp is supported")
		}
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
			return defaultKernel.Listen(Handle(iport))
		default:
			return nil, fmt.Errorf("only tcp is supported")
		}
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
	syscall.DefaultReadFunction = func(fd uintptr, b []byte) (int, error) {
		return defaultKernel.Read(Handle(fd), b)
	}
	syscall.DefaultWriteFunction = func(fd uintptr, b []byte) (int, error) {
		return defaultKernel.Write(Handle(fd), b)
	}
	syscall.DefaultStartProcessFunction = func(argv0 string, argv []string, attr *syscall.ProcAttr) (pid int, handle uintptr, err error) {
		panic(errors.New("starting processes is not supported in gopherjs"))
	}
}

func allowHost(host string) bool {
	return host == "127.0.0.1" ||
		host == "" ||
		host == "0.0.0.0" ||
		host == "locahost"
}
