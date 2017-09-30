package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/hashicorp/yamux"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	log.SetFlags(0)

	http.HandleFunc("/listen", handleListen)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:5000"
	}
	log.Printf("starting http server addr=%s\n", addr)
	http.ListenAndServe(addr, nil)
}

func handleListen(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("failed to upgrade connection to websocket:", err)
		return
	}
	defer ws.Close()

	wsc := &binaryWSConn{
		Conn: ws,
	}

	dst, err := yamux.Client(wsc, yamux.DefaultConfig())
	if err != nil {
		log.Println("failed to create session", err)
	}
	defer dst.Close()

	src, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Println("failed to create new TCP listener:", err)
		return
	}
	defer src.Close()

	log.Println("started listener", src.Addr())

	for {
		srcc, err := src.Accept()
		if err != nil {
			log.Println("failed to accept connection:", err)
			return
		}

		log.Println("received connection", srcc.RemoteAddr())

		dstc, err := dst.Open()
		if err != nil {
			srcc.Close()
			log.Println("failed to open connection:", err)
			return
		}

		log.Println("opened connection", dstc.LocalAddr())

		go func() {
			defer srcc.Close()
			defer dstc.Close()
			err := proxy(dstc, srcc)
			if err != nil {
				log.Println("error handling connection:", err)
			}
		}()
	}
}

func proxy(dst, src net.Conn) error {
	var eg errgroup.Group
	eg.Go(func() error {
		_, err := io.Copy(dst, src)
		return err
	})
	eg.Go(func() error {
		_, err := io.Copy(src, dst)
		return err
	})
	return eg.Wait()
}
