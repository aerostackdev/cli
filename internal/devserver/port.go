package devserver

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// CheckPortAvailable checks if a given TCP port is free to bind.
// Returns nil if the port is free, or a clear error with guidance if it's busy.
func CheckPortAvailable(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		// Connection refused → port is free
		return nil
	}
	conn.Close()

	// Port is occupied — give the developer a clear, actionable error
	msg := fmt.Sprintf(
		"port %d is already in use.\n\n"+
			"  A previous 'aerostack dev' session may still be running.\n"+
			"  To find and kill the process:\n"+
			"    lsof -ti :%d | xargs kill -9\n\n"+
			"  Or start on a different port:\n"+
			"    aerostack dev --port <other-port>",
		port, port,
	)
	return errors.New(msg)
}
