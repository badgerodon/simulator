package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	_ "github.com/badgerodon/simulator/kernel"
)

func main() {
	li, err := net.Listen("tcp", "127.0.0.1:7000")
	if err != nil {
		panic(err)
	}

	go func() {
		res, err := http.Get("http://127.0.0.1:7000")
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		bs, _ := ioutil.ReadAll(res.Body)
		fmt.Println("RESPONSE:", string(bs), res.Header)
	}()

	http.Serve(li, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello World")
	}))
}
