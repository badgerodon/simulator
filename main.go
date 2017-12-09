package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"unicode"

	"google.golang.org/grpc"

	"github.com/badgerodon/httpcompression"
	"github.com/badgerodon/simulator/builder/builderpb"
)

var (
	grpcServer *grpc.Server
)

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	log.SetFlags(0)

	grpcServer = grpc.NewServer()
	builderpb.RegisterServiceServer(grpcServer, getBuildServer())
	go func() {
		li, err := net.Listen("tcp", "127.0.0.1:5001")
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("starting gRPC server on :5001")
		grpcServer.Serve(li)
	}()

	http.Handle("/ui/assets/", http.StripPrefix("/ui/assets/", httpcompression.FileServer(http.Dir("./ui/assets/"))))
	http.Handle("/srv/", http.StripPrefix("/srv/", httpcompression.FileServer(http.Dir(builderDataDir))))
	http.Handle("/", http.HandlerFunc(handleCatchAll))

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:5000"
	}
	log.Println("starting server on " + addr)
	return http.ListenAndServe(addr, nil)
}

func makeFileHandler(filePath string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, filePath)
	}
}

type nopWriteCloser struct {
	io.Writer
}

func (w nopWriteCloser) Close() error {
	return nil
}

func acceptsBrotli(r *http.Request) bool {
	fs := strings.FieldsFunc(r.Header.Get("Accept-Encoding"), func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	})
	for _, f := range fs {
		if f == "br" {
			return true
		}
	}
	return false
}

func handleCatchAll(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
		grpcServer.ServeHTTP(w, r)
		return
	}

	handleUI(w, r)
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8" />
		<meta http-equiv="X-UA-Compatible" content="IE=edge" />
		<meta http-equiv="Content-Type" content="text/html;charset=utf-8" >
		<title>Simulator</title>
		<link rel="stylesheet" type="text/css" href="/ui/assets/css/normalize.css" />
		<link rel="stylesheet" type="text/css" href="/ui/assets/css/xterm.css" />
		<link rel="stylesheet" type="text/css" href="/ui/assets/css/main.css" />
	</head>
	<body>
		<header>
			<h1>Simulator</h1>
		</header>

		<section id="terminal-container">

		</section>

		<footer>
			&copy;2017 Caleb Doxsey
		</footer>

		<script src="/ui/assets/js/xterm.js"></script>
		<script src="/ui/assets/js/main.js"></script>
	</body>
</html>`)
}
