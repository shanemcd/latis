// Package connector defines the interface for transport connectors.
// Connectors move protocol messages between cmdr and units over
// various transports (SSH, WebSocket, TCP, etc.).
package connector

import "io"

// Connector represents a bidirectional connection to a unit.
type Connector interface {
	io.ReadWriteCloser
}
