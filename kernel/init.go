package kernel

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/badgerodon/simulator/kernel/vfs"
	"github.com/gopherjs/gopherjs/js"
)

func init() {
	if js.Global.Get("DedicatedWorkerGlobalScope").Bool() &&
		!js.Global.Get("document").Bool() {
		client := NewRPCMessageChannelClient(js.Global.Get("self"))
		defaultKernel = newWebWorkerKernel(client)
		js.Global.Set("ATEXIT", func() {
			client.Invoke("Exit", nil, nil)
		})
	} else {
		defaultKernel = newCoreKernel()
	}

	js.Global.Get("console").Call("log", "LOCATION", js.Global.Get("location"))

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
	syscall.DefaultStartProcessFunction = func(argv0 string, argv []string, attr *syscall.ProcAttr) (pid int, handle uintptr, err error) {
		return defaultKernel.StartProcess(argv0, argv, attr)
	}
	syscall.DefaultReadFunction = func(fd int, p []byte) (int, error) {
		return defaultKernel.Read(fd, p)
	}
	syscall.DefaultWriteFunction = func(fd int, p []byte) (int, error) {
		return defaultKernel.Write(fd, p)
	}

	// setup the file system

	fs, err := vfs.New()
	if err != nil {
		log.Fatal(err)
	}

	syscall.DefaultCloseFunction = func(fd int) error {
		return fs.Close(fd)
	}

	syscall.DefaultOpenFunction = func(path string, mode int, perm uint32) (fd int, err error) {
		return fs.Open(path, mode, perm)
	}

	syscall.DefaultFCNTLFunction = func(fd int, cmd int, arg int) (val int, err error) {
		return fs.FCNTL(fd, cmd, arg)
	}

	syscall.DefaultPipe2Function = func(p []int, flags int) error {
		r, w, err := defaultKernel.Pipe()
		if err != nil {
			return err
		}
		p[0] = r
		p[1] = w
		return nil
	}

	syscall.DefaultWait4Function = func(pid int, wstatus *syscall.WaitStatus, options int, rusage *syscall.Rusage) (wpid int, err error) {
		return 0, defaultKernel.Wait(pid)
	}

	u, err := url.Parse(js.Global.Get("location").Get("href").String())
	if err == nil {
		if fd0, _ := strconv.Atoi(u.Query().Get("files[0]")); fd0 > 0 {
			os.Stdin = os.NewFile(uintptr(fd0), "stdin")
		}
		if fd1, _ := strconv.Atoi(u.Query().Get("files[1]")); fd1 > 0 {
			os.Stdout = os.NewFile(uintptr(fd1), "stdout")
		}
		if fd2, _ := strconv.Atoi(u.Query().Get("files[2]")); fd2 > 0 {
			os.Stderr = os.NewFile(uintptr(fd2), "stderr")
		}
	}
}

func allowHost(host string) bool {
	return host == "127.0.0.1" ||
		host == "" ||
		host == "0.0.0.0" ||
		host == "locahost"
}
