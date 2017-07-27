package main

import "github.com/badgerodon/simulator/kernel"

func main() {
	kernel.StartProcess("/github.com/badgerodon/simulator/examples/ping/client/client.js", []string{
		"TEST=1",
	})
	kernel.StartProcess("/github.com/badgerodon/simulator/examples/ping/server/server.js", []string{
		"TEST=1",
	})
}
