package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/hashicorp/yamux"

	"github.com/gopherjs/gopherjs/js"
)

func init() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello")
	})
}

func main() {
	log.SetFlags(0)

	ws := js.Global.Get("WebSocket").New("ws://local.rtctunnel.com:5000/listen")
	ws.Call("addEventListener", "open", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "open", evt)
	})
	ws.Call("addEventListener", "message", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "message", evt.Get("data"))
	})
	ws.Call("addEventListener", "error", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "error", evt)
	})
	ws.Call("addEventListener", "close", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "close", evt)
	})

	conn := newWSConn(ws)
	defer conn.Close()

	li, err := yamux.Server(conn, yamux.DefaultConfig())
	if err != nil {
		log.Fatal("failed to get session", err)
	}
	defer li.Close()

	err = http.Serve(li, nil)
	if err != nil {
		log.Fatal(err)
	}
	// for {
	// 	conn, err := li.Accept()
	// 	if err != nil {
	// 		log.Fatal("failed to accept connection", err)
	// 	}

	// 	go handle(conn)
	// }
}

func handle(conn net.Conn) {
	defer conn.Close()

	s := bufio.NewScanner(conn)
	for s.Scan() {
		log.Println("recv", s.Text())

		args := strings.Fields(s.Text())
		if len(args) == 0 {
			continue
		}
		switch args[0] {
		case "get":
			io.WriteString(conn, "TEST\r\n")
		}
	}
}
