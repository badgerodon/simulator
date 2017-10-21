package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	log.SetFlags(0)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Println("starting server on 127.0.0.1:" + port)
	li, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Fatalln(err)
	}
	defer li.Close()

	for {
		conn, err := li.Accept()
		if err != nil {
			log.Fatalln(err)
		}

		handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()

	s := bufio.NewScanner(conn)
	for s.Scan() {
		_, err := fmt.Fprintln(conn, s.Text())
		if err != nil {
			log.Printf("error writing line: %v\n", err)
			return
		}
	}
}
