package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Printf("starting http server port=%s\n", port)
	http.ListenAndServe("127.0.0.1:"+port, nil)
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

	src, err := yamux.Server(wsc, yamux.DefaultConfig())
	if err != nil {
		log.Println("failed to create session", err)
	}
	defer src.Close()

	log.Println("started listener", src.Addr())
	defer log.Println("closed listener", src.Addr())

	for {
		srcc, err := src.Accept()
		if err != nil {
			log.Println("error accepting connection:", err)
			break
		}

		dst, err := net.Dial("tcp", "127.0.0.1:5002")
		if err != nil {
			srcc.Close()
			log.Println("error opening connection:", err)
			break
		}

		go func() {
			for range time.Tick(time.Second) {
				if src.IsClosed() {
					srcc.Close()
					dst.Close()
					return
				}
			}
		}()

		go func() {
			defer srcc.Close()
			defer dst.Close()
			err := proxy(dst, srcc)
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
