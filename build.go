package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/badgerodon/grpcsimulator/builder"
)

func getBuildServer() *builder.Server {
	working := os.Getenv("BUILDER_WORKING_DIR")
	if working == "" {
		working = filepath.Join(os.TempDir(), "builder-data")
		os.MkdirAll(working, 0755)
	}
	log.Println("builder working directory:", working)
	return builder.NewServer(working, "doxsey-1", "badgerodon-gosimulator", "dev")
}
