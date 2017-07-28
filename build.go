package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/badgerodon/simulator/builder"
)

var builderDataDir string

func init() {
	builderDataDir = os.Getenv("BUILDER_DATA_DIR")
	if builderDataDir == "" {
		builderDataDir = filepath.Join(os.TempDir(), "builder-data")
		os.MkdirAll(builderDataDir, 0755)
	}
}

func getBuildServer() *builder.Server {
	log.Println("builder working directory:", builderDataDir)
	server, err := builder.NewServer(builderDataDir)
	if err != nil {
		log.Fatalln(err)
	}
	return server
}
