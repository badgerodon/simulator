package main

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	for _, uiPath := range []string{
		"node_modules/react/dist/react.js",
		"node_modules/react-dom/dist/react-dom.js",
		"dist/bundle.js",
		"dist/bundle.js.map",
	} {
		http.Handle("/ui/"+uiPath, makeFileHandler("./ui/"+uiPath))
	}
	http.Handle("/api/", http.HandlerFunc(handleAPI))
	http.Handle("/", http.HandlerFunc(handleUI))

	return http.ListenAndServe("127.0.0.1:5000", nil)
}

func makeFileHandler(filePath string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, filePath)
	}
}

func handleAPI(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	if rand.Float64() > 0.5 {
		res.WriteHeader(500)
		json.NewEncoder(res).Encode(struct {
			Error string `json:"error"`
		}{
			Error: "randomly failing",
		})
		return
	}

	switch req.URL.Path {
	case "/api/projects":
		type Project struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		type ProjectListResponse struct {
			Projects []Project `json:"projects"`
		}
		json.NewEncoder(res).Encode(&ProjectListResponse{
			Projects: []Project{
				{ID: "helloworld", Name: "Hello World"},
			},
		})
	default:
		http.Error(res, "not found", 404)
	}
}

func handleUI(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, `<!DOCTYPE html>
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
