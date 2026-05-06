package sandbox

import (
	"fmt"
	"net"
)

func findFreePort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("find free port: %w", err)
	}
	defer listener.Close()

	address := listener.Addr().(*net.TCPAddr)
	return address.Port, nil
}
