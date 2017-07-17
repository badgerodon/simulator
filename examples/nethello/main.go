package main

import (
	"bufio"
	"fmt"
	"io"
	"net"

	_ "github.com/badgerodon/grpcsimulator/kernel"
)

func main() {
	li, err := net.Listen("tcp", "127.0.0.1:7000")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			conn, err := li.Accept()
			if err != nil {
				panic("failed to accept connection:" + err.Error())
			}

			s := bufio.NewScanner(conn)
			s.Scan()
			io.WriteString(conn, s.Text()+"\n")
			conn.Close()
		}
	}()

	conn, err := net.Dial("tcp", "127.0.0.1:7000")
	if err != nil {
		panic("failed to dial server:" + err.Error())
	}
	defer conn.Close()

	go func() {
		io.WriteString(conn, "Hello World\n")
	}()

	s := bufio.NewScanner(conn)
	for s.Scan() {
		fmt.Println(s.Text())
	}
}
