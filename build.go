package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/badgerodon/simulator/builder"
)

func getBuildServer() *builder.Server {
	working := os.Getenv("BUILDER_WORKING_DIR")
	if working == "" {
		working = filepath.Join(os.TempDir(), "builder-data")
		os.MkdirAll(working, 0755)
	}
	log.Println("builder working directory:", working)
	server, err := builder.NewServer(working, "badgerodon-173120", "gosimulator-build", "dev")
	if err != nil {
		log.Fatalln(err)
	}
	return server
}
