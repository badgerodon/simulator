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

	for _, uiPath := range []string{
		"node_modules/react/dist/react.js",
		"node_modules/react-dom/dist/react-dom.js",
		"dist/bundle.js",
		"dist/bundle.js.map",
		"dist/monaco-editor-worker-loader-proxy.js",
	} {
		http.Handle("/ui/"+uiPath, makeFileHandler("./ui/"+uiPath))
	}
	http.Handle("/srv/", http.StripPrefix("/srv/", http.FileServer(http.Dir(builderDataDir))))
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
		<link rel="stylesheet" type="text/css" href="https://unpkg.com/highlight.js@9.12.0/styles/github.css" />
		<link rel="stylesheet" type="text/css" href="https://unpkg.com/bulma@0.5.1/css/bulma.css" />
	</head>
	<body>
		<div id="root"></div>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/require.js/2.3.4/require.min.js"></script>
		<script>
			requirejs.config({
				paths: {
					'github-api': 'https://unpkg.com/github-api@3.0.0/dist/GitHub.bundle.min',
					'highlight': 'https://unpkg.com/highlight.js@9.12.0/lib/highlight',
					'highlight-go': 'https://unpkg.com/highlight-languages@9.8.1/go'
				}
			});
		</script>
		<script src="/ui/dist/bundle.js"></script>
	</body>
</html>`)
}
