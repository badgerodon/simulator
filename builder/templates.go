package main

import "html/template"

var (
	mainTemplate = template.Must(template.New("").Parse(`// +build !js

package main

import (
    "google.golang.org/grpc"
    "net"
)

import pb1 "github.com/badgerodon/grpcsimulator/examples/ping/pb"
import impl "github.com/badgerodon/grpcsimulator/examples/ping"

func main() {
    li, err := net.Listen("tcp", "127.0.0.1:5000")
    if err != nil {
        panic(err)
    }
    defer li.Close()

    server := grpc.NewServer()
    pb1.RegisterPingServiceServer(server, impl.New())
    err = server.Serve(li)
    if err != nil {
        panic(err)
    }
}
`))

	mainJSTemplate = template.Must(template.New("").Parse(`// +build js

package main

import (
    "google.golang.org/grpc"
    "net"
)

import pb1 "github.com/badgerodon/grpcsimulator/examples/ping/pb"
import impl "github.com/badgerodon/grpcsimulator/examples/ping"

func main() {
    li, err := net.Listen("tcp", "127.0.0.1:5000")
    if err != nil {
        panic(err)
    }
    defer li.Close()

    server := grpc.NewServer()
    pb1.RegisterPingServiceServer(server, impl.New())
    err = server.Serve(li)
    if err != nil {
        panic(err)
    }
}
`))
	envJSTemplate = template.Must(template.New("").Parse(`// +build js
package main

import (
    "strings"
    "os"
    "fmt"
    "github.com/cathalgarvey/fmtless/net/url"
    "github.com/gopherjs/gopherjs/js"
)

func init() {
    u, _ := url.Parse(js.Global.Get("location").Get("href").String())
    for k, vs := range u.Query() {
        os.Setenv(k, strings.Join(vs, ","))
    }
    fmt.Println(os.Environ())
}
`))
)
