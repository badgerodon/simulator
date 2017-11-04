package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/badgerodon/simulator/builder/builderpb"

	"github.com/badgerodon/simulator/builder"
)

var (
	builderDataDir string
	builderServer  builderpb.ServiceServer
)

func init() {
	var err error

	builderDataDir = os.Getenv("BUILDER_DATA_DIR")
	if builderDataDir == "" {
		builderDataDir = filepath.Join(os.TempDir(), "builder-data")
		os.MkdirAll(builderDataDir, 0755)
	}

	log.Println("builder working directory:", builderDataDir)
	builderServer, err = builder.NewServer(builderDataDir)
	if err != nil {
		log.Fatalln(err)
	}
}

func handleBuild(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	req := &builderpb.BuildRequest{
		ImportPath: r.URL.Query().Get("import_path"),
		Branch:     r.URL.Query().Get("branch"),
	}
	res, err := builderServer.Build(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Location = filepath.Join("/srv", res.Location[len(builderDataDir):])
	json.NewEncoder(w).Encode(res)
}
