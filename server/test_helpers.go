package server

import (
	"fmt"
	"net"
)

// Get a free address for a test tendermint server
// protocol is either tcp, http, etc
func FreeTCPAddr() (addr, port string, err error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", "", err
	}

	closer := func() {
		err := l.Close()
		if err != nil {
			// TODO: Handle with #870
			panic(err)
		}
	}

	defer closer()

	portI := l.Addr().(*net.TCPAddr).Port
	port = fmt.Sprintf("%d", portI)
	addr = fmt.Sprintf("tcp://0.0.0.0:%s", port)
	return
}
