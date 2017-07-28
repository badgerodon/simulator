package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
	} {
		http.Handle("/ui/"+uiPath, makeFileHandler("./ui/"+uiPath))
	}
	http.Handle("/build/", http.HandlerFunc(handleBuild))
	http.Handle("/", http.HandlerFunc(handleCatchAll))

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Println("starting server on :" + port)
	return http.ListenAndServe(":"+port, nil)
}

func makeFileHandler(filePath string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, filePath)
	}
}

func handleBuild(w http.ResponseWriter, r *http.Request) {
	importPath := r.URL.Path[len("/build/"):]
	branch := r.URL.Query().Get("branch")
	if branch == "" {
		branch = "master"
	}

	cc, err := grpc.Dial("127.0.0.1:5001", grpc.WithInsecure())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cc.Close()

	c := builderpb.NewServiceClient(cc)
	res, err := c.Build(context.Background(), &builderpb.BuildRequest{
		ImportPath: importPath,
		Branch:     branch,
	})
	if err != nil {
		log.Printf("failed to build: %v\n", err)
		http.Error(w, err.Error(), 500)
		return
	}

	location := res.GetLocation()
	http.Redirect(w, r, location[len(builderDataDir):], http.StatusTemporaryRedirect)

	// if strings.HasSuffix(r.URL.Path, ".map") {
	// 	location += ".map"
	// }

	// f, err := os.Open(location)
	// if err != nil {
	// 	http.Error(w, err.Error(), 500)
	// 	return
	// }
	// defer f.Close()

	// var out io.WriteCloser = nopWriteCloser{w}
	// switch {
	// case acceptsBrotli(r):
	// 	w.Header().Set("Content-Encoding", "br")
	// 	out = cbrotli.NewWriter(out, cbrotli.WriterOptions{
	// 		Quality: 6,
	// 	})
	// }

	// log.Println("RESULT", location)

	// w.Header().Set("Content-Type", "text/javascript")
	// io.Copy(out, f)
	// out.Close()
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
	log.Println("HANDLE", r.URL, r.Header)
	if strings.HasSuffix(r.URL.Path, ".js") ||
		strings.HasSuffix(r.URL.Path, ".map") {
		builderPath := filepath.Join(builderDataDir, r.URL.Path)
		if _, err := os.Stat(builderPath); err == nil {
			http.ServeFile(w, r, builderPath)
			return
		}
	}

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
		<title>gRPC Simulator</title>
	</head>
	<body>
		<div id="root"></div>
		<script src="/ui/dist/bundle.js"></script>
	</body>
</html>`)
}
