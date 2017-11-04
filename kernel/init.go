package kernel

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
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

	os.Setenv("TERM", "xterm-256color")
	ti, _ := base64.StdEncoding.DecodeString(`
		GgElACYADwCdAb8FeHRlcm0tMjU2Y29sb3J8eHRlcm0gd2l0aCAyNTYgY29sb3JzAAABAAABAAAA
		AQAAAAABAQAAAAAAAAABAAABAAEBAAAAAAAAAAABAFAACAAYAP//////////////////////////
		AAH/fwAABAAGAAgAGQAeACYAKgAuAP//OQBKAEwAUABXAP//WQBmAP//agBuAHgAfAD/////gACE
		AIkAjgD//5cAnAChAP//pgCrALAAtQC+AMIAyQD//9IA1wDdAOMA////////9QD///////8HAf//
		CwH///////8NAf//EgH//////////xYBGgEgASQBKAEsATIBOAE+AUQBSgFOAf//UwH//1cBXAFh
		AWUBbAH//3MBdwF/Af///////////////////////////////////////4cBkAGZAaIBqwG0Ab0B
		xgHPAdgB////////4QHlAeoB///vAfgB/////woCDQIYAhsCHQIgAn0C//+AAv//////////////
		/4IC//////////+GAv//uwL/////vwLFAv/////////////////////////////LAs8C////////
		///////////////////////////////////////////////////////////TAv/////aAv//////
		////4QLoAu8C//////YC///9Av///////wQD/////////////wsDEQMXAx4DJQMsAzMDOwNDA0sD
		UwNbA2MDawNzA3oDgQOIA48DlwOfA6cDrwO3A78DxwPPA9YD3QPkA+sD8wP7AwMECwQTBBsEIwQr
		BDIEOQRABEcETwRXBF8EZwRvBHcEfwSHBI4ElQScBP//////////////////////////////////
		//////////////////////////+hBKwEsQS5BL0ExgTNBP////////////////////////////8r
		Bf///////////////////////zAF////////////////////////////////////////////////
		////////////////////////////////////////NgX///////86BXkF////////////////////
		////////////////////////////////////////////////////////////////////////////
		/////////////////////////////////////7kFvAUbW1oABwANABtbJWklcDElZDslcDIlZHIA
		G1szZwAbW0gbWzJKABtbSwAbW0oAG1slaSVwMSVkRwAbWyVpJXAxJWQ7JXAyJWRIAAoAG1tIABtb
		PzI1bAAIABtbPzEybBtbPzI1aAAbW0MAG1tBABtbPzEyOzI1aAAbW1AAG1tNABsoMAAbWzVtABtb
		MW0AG1s/MTA0OWgAG1sybQAbWzRoABtbOG0AG1s3bQAbWzdtABtbNG0AG1slcDElZFgAGyhCABso
		QhtbbQAbWz8xMDQ5bAAbWzRsABtbMjdtABtbMjRtABtbPzVoJDwxMDAvPhtbPzVsABtbIXAbWz8z
		OzRsG1s0bBs+ABtbTAB/ABtbM34AG09CABtPUAAbWzIxfgAbT1EAG09SABtPUwAbWzE1fgAbWzE3
		fgAbWzE4fgAbWzE5fgAbWzIwfgAbT0gAG1syfgAbT0QAG1s2fgAbWzV+ABtPQwAbWzE7MkIAG1sx
		OzJBABtPQQAbWz8xbBs+ABtbPzFoGz0AG1slcDElZFAAG1slcDElZE0AG1slcDElZEIAG1slcDEl
		ZEAAG1slcDElZFMAG1slcDElZEwAG1slcDElZEQAG1slcDElZEMAG1slcDElZFQAG1slcDElZEEA
		G1tpABtbNGkAG1s1aQAbYxtdMTA0BwAbWyFwG1s/Mzs0bBtbNGwbPgAbOAAbWyVpJXAxJWRkABs3
		AAoAG00AJT8lcDkldBsoMCVlGyhCJTsbWzAlPyVwNiV0OzElOyU/JXA1JXQ7MiU7JT8lcDIldDs0
		JTslPyVwMSVwMyV8JXQ7NyU7JT8lcDQldDs1JTslPyVwNyV0OzglO20AG0gACQAbT0UAYGBhYWZm
		Z2dpaWpqa2tsbG1tbm5vb3BwcXFycnNzdHR1dXZ2d3d4eHl5enp7e3x8fX1+fgAbW1oAG1s/N2gA
		G1s/N2wAG09GABtPTQAbWzM7Mn4AG1sxOzJGABtbMTsySAAbWzI7Mn4AG1sxOzJEABtbNjsyfgAb
		WzU7Mn4AG1sxOzJDABtbMjN+ABtbMjR+ABtbMTsyUAAbWzE7MlEAG1sxOzJSABtbMTsyUwAbWzE1
		OzJ+ABtbMTc7Mn4AG1sxODsyfgAbWzE5OzJ+ABtbMjA7Mn4AG1syMTsyfgAbWzIzOzJ+ABtbMjQ7
		Mn4AG1sxOzVQABtbMTs1UQAbWzE7NVIAG1sxOzVTABtbMTU7NX4AG1sxNzs1fgAbWzE4OzV+ABtb
		MTk7NX4AG1syMDs1fgAbWzIxOzV+ABtbMjM7NX4AG1syNDs1fgAbWzE7NlAAG1sxOzZRABtbMTs2
		UgAbWzE7NlMAG1sxNTs2fgAbWzE3OzZ+ABtbMTg7Nn4AG1sxOTs2fgAbWzIwOzZ+ABtbMjE7Nn4A
		G1syMzs2fgAbWzI0OzZ+ABtbMTszUAAbWzE7M1EAG1sxOzNSABtbMTszUwAbWzE1OzN+ABtbMTc7
		M34AG1sxODszfgAbWzE5OzN+ABtbMjA7M34AG1syMTszfgAbWzIzOzN+ABtbMjQ7M34AG1sxOzRQ
		ABtbMTs0UQAbWzE7NFIAG1sxSwAbWyVpJWQ7JWRSABtbNm4AG1s/MTsyYwAbW2MAG1szOTs0OW0A
		G10xMDQHABtdNDslcDElZDtyZ2I6JXAyJXsyNTV9JSolezEwMDB9JS8lMi4yWC8lcDMlezI1NX0l
		KiV7MTAwMH0lLyUyLjJYLyVwNCV7MjU1fSUqJXsxMDAwfSUvJTIuMlgbXAAbWzNtABtbMjNtABtb
		TQAbWyU/JXAxJXs4fSU8JXQzJXAxJWQlZSVwMSV7MTZ9JTwldDklcDElezh9JS0lZCVlMzg7NTsl
		cDElZCU7bQAbWyU/JXAxJXs4fSU8JXQ0JXAxJWQlZSVwMSV7MTZ9JTwldDEwJXAxJXs4fSUtJWQl
		ZTQ4OzU7JXAxJWQlO20AG2wAG20AAAIAAAA+AH4A7gIBAQAABwATABgAKgAwADoAQQBIAE8AVgBd
		AGQAawByAHkAgACHAI4AlQCcAKMAqgCxALgAvwDGAM0A1ADbAOIA6QDwAPcA/gAFAQwBEwEaASEB
		KAEvATYBPQFEAUsBUgFZAWABZwFuAXUBfAGDAYoBkQGYAZ8B//////////8AAAMABgAJAAwADwAS
		ABUAGAAdACIAJwAsADEANQA6AD8ARABJAE4AVABaAGAAZgBsAHIAeAB+AIQAigCPAJQAmQCeAKMA
		qQCvALUAuwDBAMcAzQDTANkA3wDlAOsA8QD3AP0AAwEJAQ8BFQEbAR8BJAEpAS4BMwE4ATwBQAFE
		ARtdMTEyBwAbXTEyOyVwMSVzBwAbWzNKABtdNTI7JXAxJXM7JXAyJXMHABtbMiBxABtbJXAxJWQg
		cQAbWzM7M34AG1szOzR+ABtbMzs1fgAbWzM7Nn4AG1szOzd+ABtbMTsyQgAbWzE7M0IAG1sxOzRC
		ABtbMTs1QgAbWzE7NkIAG1sxOzdCABtbMTszRgAbWzE7NEYAG1sxOzVGABtbMTs2RgAbWzE7N0YA
		G1sxOzNIABtbMTs0SAAbWzE7NUgAG1sxOzZIABtbMTs3SAAbWzI7M34AG1syOzR+ABtbMjs1fgAb
		WzI7Nn4AG1syOzd+ABtbMTszRAAbWzE7NEQAG1sxOzVEABtbMTs2RAAbWzE7N0QAG1s2OzN+ABtb
		Njs0fgAbWzY7NX4AG1s2OzZ+ABtbNjs3fgAbWzU7M34AG1s1OzR+ABtbNTs1fgAbWzU7Nn4AG1s1
		Ozd+ABtbMTszQwAbWzE7NEMAG1sxOzVDABtbMTs2QwAbWzE7N0MAG1sxOzJBABtbMTszQQAbWzE7
		NEEAG1sxOzVBABtbMTs2QQAbWzE7N0EAQVgAWFQAQ3IAQ3MARTMATXMAU2UAU3MAa0RDMwBrREM0
		AGtEQzUAa0RDNgBrREM3AGtETgBrRE4zAGtETjQAa0RONQBrRE42AGtETjcAa0VORDMAa0VORDQA
		a0VORDUAa0VORDYAa0VORDcAa0hPTTMAa0hPTTQAa0hPTTUAa0hPTTYAa0hPTTcAa0lDMwBrSUM0
		AGtJQzUAa0lDNgBrSUM3AGtMRlQzAGtMRlQ0AGtMRlQ1AGtMRlQ2AGtMRlQ3AGtOWFQzAGtOWFQ0
		AGtOWFQ1AGtOWFQ2AGtOWFQ3AGtQUlYzAGtQUlY0AGtQUlY1AGtQUlY2AGtQUlY3AGtSSVQzAGtS
		SVQ0AGtSSVQ1AGtSSVQ2AGtSSVQ3AGtVUABrVVAzAGtVUDQAa1VQNQBrVVA2AGtVUDcAa2EyAGti
		MQBrYjMAa2MyAA==
	`)
	ioutil.WriteFile("/usr/share/terminfo/x/xterm-256color", ti, 0777)
}

func allowHost(host string) bool {
	return host == "127.0.0.1" ||
		host == "" ||
		host == "0.0.0.0" ||
		host == "locahost"
}
