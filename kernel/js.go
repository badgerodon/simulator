// +build js

import (
	"net"
)

func init() {
	k := newCoreKernel()
	net.SetDefaultListen(func(net, addr string) (net.Listener, error) {
		return nil, fmt.Errorf("not implemented")
	})
}
