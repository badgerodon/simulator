package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	addr := "127.0.0.1:" + port
	log.Println("starting http server on", addr)
	err := http.ListenAndServe(addr, http.HandlerFunc(handle))
	if err != nil {
		log.Fatalln(err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/build"):
		handleBuild(w, r)
	case strings.HasPrefix(r.URL.Path, "/run/"):
		io.WriteString(w, `<!DOCTYPE html>
		<html>
			<head>
				<meta charset="UTF-8" />
				<meta http-equiv="X-UA-Compatible" content="IE=edge" />
				<meta http-equiv="Content-Type" content="text/html;charset=utf-8" >
				<title>Badgerodon Simulator</title>
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
	case strings.HasPrefix(r.URL.Path, "/srv/"):
		p := filepath.Join(builderDataDir, filepath.FromSlash(path.Clean(r.URL.Path[len("/srv/"):])))
		fmt.Println(p)
		http.ServeFile(w, r, p)
	case strings.HasPrefix(r.URL.Path, "/ui/"):
		path := r.URL.Path[1:]
		info, err := AssetInfo(path)
		if err != nil {
			http.Error(w, "unknown page", http.StatusNotFound)
			return
		}
		data, err := Asset(path)
		if err != nil {
			http.Error(w, "unknown page", http.StatusNotFound)
			return
		}
		http.ServeContent(w, r, info.Name(), info.ModTime(), bytes.NewReader(data))
	default:
		http.Error(w, "unknown page", http.StatusNotFound)
	}
}
