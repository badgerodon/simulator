package main

import "github.com/badgerodon/grpcsimulator/kernel"

func main() {
	kernel.StartProcess("/github.com/badgerodon/grpcsimulator/examples/ping/client/client.js", []string{
		"TEST=1",
	})
	kernel.StartProcess("/github.com/badgerodon/grpcsimulator/examples/ping/server/server.js", []string{
		"TEST=1",
	})
}
