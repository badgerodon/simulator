// +build !js

package main

import (
	"log"

	"github.com/badgerodon/grpcsimulator/mesh"
)

func main() {
	log.SetFlags(0)

	m := mesh.New()
	err := m.Serve()
	if err != nil {
		log.Fatalln(err)
	}
}
