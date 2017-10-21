package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"time"

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

	ws := js.Global.Get("WebSocket").New("ws://localhost:5001/listen")
	ws.Call("addEventListener", "open", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "open", evt)
	})
	ws.Call("addEventListener", "message", func(evt *js.Object) {
		//js.Global.Get("console").Call("log", "message", evt.Get("data"))
	})
	ws.Call("addEventListener", "error", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "error", evt)
	})
	ws.Call("addEventListener", "close", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "close", evt)
	})

	wsc := newWSConn(ws)
	defer wsc.Close()

	session, err := yamux.Client(wsc, yamux.DefaultConfig())
	if err != nil {
		log.Fatal("failed to get session", err)
	}
	defer session.Close()

	c, err := session.Open()
	if err != nil {
		log.Fatal("failed to get connection", err)
	}
	defer c.Close()

	go func() {
		for range time.Tick(time.Second * 2) {
			io.WriteString(c, "Hello World\n")
		}
	}()

	s := bufio.NewScanner(c)
	for s.Scan() {
		js.Global.Get("document").Call("write", "FROM SERVER: <tt>"+s.Text()+"</tt><br>")
	}
}
