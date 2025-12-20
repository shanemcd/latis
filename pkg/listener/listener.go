// Package listener defines the interface for accepting inbound connections from units.
package listener

import (
	"context"

	"github.com/shanemcd/latis/pkg/connector"
)

// Listener accepts inbound connections from units that dial out to cmdr.
type Listener interface {
	// Listen starts accepting connections. Returns a channel of new connections.
	Listen(ctx context.Context) (<-chan connector.Connector, error)

	// Addr returns the address the listener is bound to.
	Addr() string

	// Close stops the listener.
	Close() error
}
