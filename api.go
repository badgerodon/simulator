package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"path"

	"github.com/badgerodon/simulator/builder/builderpb"
	"google.golang.org/grpc"
)

func init() {
	http.HandleFunc("/api/build", handleAPIBuild)
}

func handleAPIBuild(w http.ResponseWriter, r *http.Request) {
	importPath := r.URL.Query().Get("import_path")
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

	var buildResult = struct {
		Location string `json:"location"`
	}{
		Location: path.Join("/srv", res.GetLocation()[len(builderDataDir):]),
	}
	json.NewEncoder(w).Encode(buildResult)
}
